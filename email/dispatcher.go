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
	Body        string // HTML body (from --template)
	PlainText   string // Plain-text body (from --text, for multipart/alternative)
	Retries     int
	Attachments []string
	CC          []string
	BCC         []string
	Index       int // Position in original task list for offset tracking
}

// OffsetTracker interface for tracking email delivery progress.
//
// MarkComplete records that the absolute task index `idx` was successfully
// delivered. Implementations are expected to maintain a contiguous high-water
// mark — the saved offset advances only past indices that have all been
// completed, even when workers finish out of order.
//
// UpdateOffset and Save remain on the interface for backward compatibility
// with external callers (CLI cleanup paths, tests).
type OffsetTracker interface {
	UpdateOffset(offset int)
	MarkComplete(idx int)
	Save() error
}

type worker struct {
	ID              int
	TaskQueue       <-chan Task
	Config          config.SMTPConfig
	Wg              *sync.WaitGroup
	BatchSize       int
	Monitor         monitor.Monitor
	Tracker         OffsetTracker
	AttachmentCache *AttachmentCache
	Ctx             context.Context
	Sent            *atomic.Int64
	Failed          *atomic.Int64
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
	Context context.Context
	Monitor monitor.Monitor
	Tracker OffsetTracker
	// StartOffset is retained for API compatibility but is no longer consulted
	// for offset arithmetic — Task.Index is treated as an absolute position
	// and Tracker.MarkComplete handles concurrency-safe advancement.
	StartOffset int
	// AttachmentCache is reused across all sends in this dispatch. When nil,
	// the dispatcher creates one with DefaultAttachmentCacheLimit. Pass a
	// pre-populated cache to share across multiple dispatches.
	AttachmentCache *AttachmentCache
	// OffsetSaveInterval controls how often the tracker is flushed to disk.
	// Zero falls back to 200ms. Negative disables periodic flushing (only the
	// final save at dispatch end is performed).
	OffsetSaveInterval time.Duration
	// PendingEmails seeds the monitor dashboard with all expected recipients
	// before any worker runs. StartDispatcher populates this from the task
	// slice automatically; streaming callers should supply it from their
	// recipient list so the dashboard shows everyone in Pending state.
	PendingEmails []string
}

// DispatchResult holds summary statistics from a dispatch run.
type DispatchResult struct {
	Sent   int
	Failed int
}

// StartDispatcher sends emails using a worker pool. It is the bulk entry point
// for slice-based task lists; for streaming sources see StartDispatcherStream.
//
// Per-success offset persistence is coalesced through a background flusher so
// the disk syncs do not serialize the worker hot path. Offset semantics use a
// contiguous high-water mark (Tracker.MarkComplete) so resume is correct under
// concurrency.
func StartDispatcher(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int, opts *DispatchOptions) DispatchResult {
	if concurrency < 1 {
		concurrency = 1
	}
	if batchSize < 1 {
		batchSize = 1
	}

	if opts == nil {
		opts = &DispatchOptions{}
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.Monitor == nil {
		opts.Monitor = monitor.NewNoOpMonitor()
	}

	ctx := opts.Context

	// Bulk-seed the monitor from the task slice when the caller did not
	// pre-supply PendingEmails. runDispatch performs the actual seeding so
	// streaming entry points share the same path.
	if opts.PendingEmails == nil && len(tasks) > 0 {
		emails := make([]string, 0, len(tasks))
		for i := range tasks {
			emails = append(emails, tasks[i].Recipient.Email)
		}
		opts.PendingEmails = emails
	}

	// Buffer sized to hold all tasks so the dispatch goroutine never blocks.
	taskBufSize := maxInt(len(tasks), concurrency*batchSize)
	if taskBufSize > 5000 {
		taskBufSize = 5000
	} else if taskBufSize < concurrency {
		taskBufSize = concurrency
	}

	taskChan := make(chan Task, taskBufSize)

	// Producer: feed tasks into the channel.
	go func() {
		for i := range tasks {
			select {
			case taskChan <- tasks[i]:
			case <-ctx.Done():
				log.Printf("Context cancelled, stopping task dispatch")
				close(taskChan)
				return
			}
		}
		close(taskChan)
	}()

	return runDispatch(ctx, taskChan, cfg, concurrency, batchSize, opts)
}

// StartDispatcherStream consumes tasks from a channel rather than a slice.
// Use it when the recipient list is large enough that holding all rendered
// tasks in RAM is undesirable. The caller is responsible for closing taskCh
// when no more tasks will be produced.
//
// All workers will return once taskCh is drained or ctx is cancelled.
func StartDispatcherStream(ctx context.Context, taskCh <-chan Task, cfg config.SMTPConfig, concurrency int, batchSize int, opts *DispatchOptions) DispatchResult {
	if concurrency < 1 {
		concurrency = 1
	}
	if batchSize < 1 {
		batchSize = 1
	}
	if opts == nil {
		opts = &DispatchOptions{}
	}
	if opts.Monitor == nil {
		opts.Monitor = monitor.NewNoOpMonitor()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return runDispatch(ctx, taskCh, cfg, concurrency, batchSize, opts)
}

// runDispatch implements the worker pool, attachment cache, and offset flusher
// shared by both the slice- and channel-based entry points.
func runDispatch(ctx context.Context, taskCh <-chan Task, cfg config.SMTPConfig, concurrency, batchSize int, opts *DispatchOptions) DispatchResult {
	var sent atomic.Int64
	var failed atomic.Int64

	mon := opts.Monitor
	tracker := opts.Tracker

	if len(opts.PendingEmails) > 0 {
		mon.InitializePending(opts.PendingEmails)
	}

	cache := opts.AttachmentCache
	if cache == nil {
		cache = NewAttachmentCache(0)
	}

	// Start the offset flusher when a tracker is supplied. The hot path will
	// only call tracker.MarkComplete; the flusher takes care of disk syncs.
	stopFlusher := startOffsetFlusher(tracker, opts.OffsetSaveInterval)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go startWorker(worker{
			ID:              i + 1,
			TaskQueue:       taskCh,
			Config:          cfg,
			Wg:              &wg,
			BatchSize:       batchSize,
			Monitor:         mon,
			Tracker:         tracker,
			AttachmentCache: cache,
			Ctx:             ctx,
			Sent:            &sent,
			Failed:          &failed,
		})
	}

	wg.Wait()

	// Stop the flusher and wait for it to fully exit before doing the final
	// synchronous save. This prevents a concurrent Save from racing on the
	// temp-file path used by the atomic rename.
	if stopFlusher != nil {
		stopFlusher()
	}
	if tracker != nil {
		if err := tracker.Save(); err != nil {
			log.Printf("offset: final save failed: %v", err)
		}
	}

	return DispatchResult{
		Sent:   int(sent.Load()),
		Failed: int(failed.Load()),
	}
}

// startOffsetFlusher kicks off a background goroutine that periodically calls
// tracker.Save while sends are in progress. Returns a stop function that
// signals the flusher and waits for it to exit; the flusher does not perform
// a final save (the caller does that synchronously).
//
// When tracker is nil or the interval is negative the flusher is not started
// and nil is returned so callers can skip the stop call.
func startOffsetFlusher(tracker OffsetTracker, interval time.Duration) func() {
	if tracker == nil {
		return nil
	}
	if interval == 0 {
		interval = 200 * time.Millisecond
	}
	if interval < 0 {
		return nil
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if err := tracker.Save(); err != nil {
					log.Printf("offset: periodic save failed: %v", err)
				}
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}
