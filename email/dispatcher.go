package email

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/parser"
)

// Task represents an email send job with recipient data.
type Task struct {
	Recipient   parser.Recipient
	Subject     string
	Body        string
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
	RetryChan   chan<- Task
	Config      config.SMTPConfig
	Wg          *sync.WaitGroup
	RetryWg     *sync.WaitGroup
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

// StartDispatcher sends emails using worker pool with retries and monitoring.
// This is the main dispatcher function that all variants should use.
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

	// Calculate buffer sizes based on workload
	taskBufSize := maxInt(len(tasks)/2, concurrency*batchSize*2)
	if taskBufSize > 2000 {
		taskBufSize = 2000
	} else if taskBufSize < concurrency {
		taskBufSize = concurrency
	}
	retryBufSize := maxInt(len(tasks)/10, concurrency*5)
	if retryBufSize > 1000 {
		retryBufSize = 1000
	} else if retryBufSize < concurrency {
		retryBufSize = concurrency
	}

	taskChan := make(chan Task, taskBufSize)
	retryChan := make(chan Task, retryBufSize)

	var wg sync.WaitGroup
	var retryWg sync.WaitGroup

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
			RetryChan:   retryChan,
			Config:      cfg,
			Wg:          &wg,
			RetryWg:     &retryWg,
			BatchSize:   batchSize,
			Monitor:     mon,
			Tracker:     tracker,
			StartOffset: startOffset,
			Ctx:         ctx,
			Sent:        &sent,
			Failed:      &failed,
		})
	}

	// Dispatch initial tasks
	go func() {
		for _, task := range tasks {
			select {
			case taskChan <- task:
			case <-ctx.Done():
				log.Printf("Context cancelled, stopping task dispatch")
				return // return instead of break to exit the goroutine
			}
		}
		close(taskChan)
	}()

	// Handle retries
	go func() {
		for task := range retryChan {
			if task.Retries > 0 {
				retryWg.Add(1)
				select {
				case taskChan <- task:
					retryWg.Done()
				case <-ctx.Done():
					retryWg.Done()
					log.Printf("Context cancelled, stopping retry processing")
					return
				default:
					// Channel full, use a non-blocking approach with proper cleanup
					scheduled := make(chan struct{})
					go func(t Task) {
						defer func() {
							retryWg.Done()
							close(scheduled)
						}()
						select {
						case taskChan <- t:
						case <-ctx.Done():
							log.Printf("Retry cancelled for %s due to context", t.Recipient.Email)
						case <-time.After(5 * time.Second):
							log.Printf("Retry timeout for %s", t.Recipient.Email)
						}
					}(task)
					// Wait briefly for the goroutine to start, then continue
					select {
					case <-scheduled:
					case <-ctx.Done():
						// If context cancelled while waiting, don't block
					default:
					}
				}
			} else {
				log.Printf("Permanent failure: %s", task.Recipient.Email)
			}
		}
	}()

	wg.Wait()
	close(retryChan)
	retryWg.Wait()

	return DispatchResult{
		Sent:   int(sent.Load()),
		Failed: int(failed.Load()),
	}
}
