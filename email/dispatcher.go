package email

import (
	"context"
	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/internal/metrics"
	"github.com/bravo1goingdark/mailgrid/internal/ratelimit"
	"github.com/bravo1goingdark/mailgrid/parser"
	"log"
	"sync"
	"time"
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
}

type worker struct {
	ID          int
	TaskQueue   <-chan Task
	RetryChan   chan<- Task
	Config      config.SMTPConfig
	Wg          *sync.WaitGroup
	RetryWg     *sync.WaitGroup
	BatchSize   int
	RateLimiter *ratelimit.RateLimiter
	Metrics     *metrics.Metrics
	Ctx         context.Context
}

// StartDispatcher spawns workers and processes email tasks with retries, rate limiting, and batch-mode dispatch.
func StartDispatcher(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int) {
	StartDispatcherWithContext(context.Background(), tasks, cfg, concurrency, batchSize, 0, 0)
}

// StartDispatcherWithContext provides enhanced dispatcher with context, rate limiting, and metrics
func StartDispatcherWithContext(ctx context.Context, tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int, rateLimit int, burstLimit int) {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize rate limiter and metrics
	rateLimiter := ratelimit.NewRateLimiter(rateLimit, burstLimit)
	metricsInstance := metrics.GetMetrics()

	// Channel sizes based on task count for better memory efficiency
	taskChanSize := min(len(tasks), 1000)
	retryChanSize := min(len(tasks)/2, 500)

	taskChan := make(chan Task, taskChanSize)
	retryChan := make(chan Task, retryChanSize)

	var wg sync.WaitGroup
	var retryWg sync.WaitGroup

	// Spawn workers with enhanced features
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		metricsInstance.RecordWorkerStart()
		go startWorker(worker{
			ID:          i + 1,
			TaskQueue:   taskChan,
			RetryChan:   retryChan,
			Config:      cfg,
			Wg:          &wg,
			RetryWg:     &retryWg,
			BatchSize:   batchSize,
			RateLimiter: rateLimiter,
			Metrics:     metricsInstance,
			Ctx:         ctx,
		})
	}

	// Dispatch initial tasks with graceful shutdown
	go func() {
		defer close(taskChan)
		for _, task := range tasks {
			select {
			case taskChan <- task:
			case <-ctx.Done():
				log.Printf("Task dispatch cancelled due to context cancellation")
				return
			}
		}
	}()

	// Handle retries with context awareness
	go func() {
		for {
			select {
			case task, ok := <-retryChan:
				if !ok {
					return
				}
				if task.Retries > 0 {
					retryWg.Add(1)
					go func(t Task) {
						defer retryWg.Done()
						select {
						case taskChan <- t:
						case <-ctx.Done():
							log.Printf("Task retry cancelled: %s", t.Recipient.Email)
						}
					}(task)
				} else {
					log.Printf("Permanent failure: %s", task.Recipient.Email)
					metricsInstance.RecordEmailFailed()
				}
			case <-ctx.Done():
				log.Printf("Retry handler cancelled")
				return
			}
		}
	}()

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(retryChan)
		retryWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("All email tasks completed successfully")
	case <-ctx.Done():
		log.Printf("Email dispatch cancelled: %v", ctx.Err())
		cancel() // Ensure all workers are notified
		// Wait a bit for graceful shutdown
		shutdownTimer := time.NewTimer(5 * time.Second)
		select {
		case <-done:
			log.Printf("Graceful shutdown completed")
		case <-shutdownTimer.C:
			log.Printf("Forced shutdown after timeout")
		}
		shutdownTimer.Stop()
	}

	// Clean up metrics
	for i := 0; i < concurrency; i++ {
		metricsInstance.RecordWorkerStop()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
