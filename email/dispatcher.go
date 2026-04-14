package email

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/parser"
)

// Task represents an email send job with recipient data.
type Task struct {
	Recipient   parser.Recipient
	Subject     string
	Body        string // HTML body (from --template)
	PlainText   string // Plain-text body (from --text, for multipart/alternative)
	Retries     int
	Attachments []string
	CC          []string
	BCC         []string
	Index       int // Position in original task list for offset tracking
}

// OffsetTracker interface for tracking email delivery progress
type OffsetTracker interface {
	UpdateOffset(offset int)
	Save() error
}

type worker struct {
	ID          int
	TaskQueue   <-chan Task
	Config      config.SMTPConfig
	Wg          *sync.WaitGroup
	BatchSize   int
	Monitor     monitor.Monitor
	Tracker     OffsetTracker
	StartOffset int
	Ctx         context.Context
	Sent        *atomic.Int64
	Failed      *atomic.Int64
}

// maxInt returns the larger of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// DispatchOptions contains optional parameters for email dispatching
type DispatchOptions struct {
	Context     context.Context
	Monitor     monitor.Monitor
	Tracker     OffsetTracker
	StartOffset int
}

// DispatchResult holds summary statistics from a dispatch run.
type DispatchResult struct {
	Sent   int
	Failed int
}

// StartDispatcher sends emails using a worker pool. Retries are handled
// inline within each worker (sleep + retry), avoiding channel races and
// goroutine leaks from the previous time.AfterFunc approach.
func StartDispatcher(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int, opts *DispatchOptions) DispatchResult {
	var sent atomic.Int64
	var failed atomic.Int64

	if opts == nil {
		opts = &DispatchOptions{
			Context: context.Background(),
			Monitor: monitor.NewNoOpMonitor(),
		}
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.Monitor == nil {
		opts.Monitor = monitor.NewNoOpMonitor()
	}

	ctx := opts.Context
	mon := opts.Monitor
	tracker := opts.Tracker
	startOffset := opts.StartOffset

	// Buffer sized to hold all tasks so the dispatch goroutine never blocks.
	taskBufSize := maxInt(len(tasks), concurrency*batchSize)
	if taskBufSize > 5000 {
		taskBufSize = 5000
	} else if taskBufSize < concurrency {
		taskBufSize = concurrency
	}

	taskChan := make(chan Task, taskBufSize)

	var wg sync.WaitGroup

	// Initialize all recipients as pending in monitor
	for _, task := range tasks {
		mon.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusPending, 0, "")
	}

	// Spawn workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go startWorker(worker{
			ID:          i + 1,
			TaskQueue:   taskChan,
			Config:      cfg,
			Wg:          &wg,
			BatchSize:   batchSize,
			Monitor:     mon,
			Tracker:     tracker,
			StartOffset: startOffset,
			Ctx:         ctx,
			Sent:        &sent,
			Failed:      &failed,
		})
	}

	// Dispatch all tasks then close the channel so workers know when to stop.
	go func() {
		for _, task := range tasks {
			select {
			case taskChan <- task:
			case <-ctx.Done():
				log.Printf("Context cancelled, stopping task dispatch")
				close(taskChan)
				return
			}
		}
		close(taskChan)
	}()

	wg.Wait()

	return DispatchResult{
		Sent:   int(sent.Load()),
		Failed: int(failed.Load()),
	}
}
