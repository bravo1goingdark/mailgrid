package email

import (
	"log"
	"mailgrid/config"
	"mailgrid/parser"
	"net/smtp"
	"sync"
)

// Task represents an email send job with recipient data.
type Task struct {
	Recipient parser.Recipient
	Subject   string
	Body      string
	Retries   int
}

type worker struct {
	ID        int
	SMTPConn  *smtp.Client
	TaskQueue <-chan Task
	RetryChan chan<- Task
	Config    config.SMTPConfig
	Wg        *sync.WaitGroup
	RetryWg   *sync.WaitGroup
	BatchSize int
}

// StartDispatcher spawns workers and processes email tasks with retry and batch-mode dispatch.
func StartDispatcher(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int) {
	taskChan := make(chan Task, 100)
	retryChan := make(chan Task, 100)

	var wg sync.WaitGroup
	var retryWg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go startWorker(worker{
			ID:        i + 1,
			TaskQueue: taskChan,
			RetryChan: retryChan,
			Config:    cfg,
			Wg:        &wg,
			RetryWg:   &retryWg,
			BatchSize: batchSize,
		})
	}

	// Feed initial tasks
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan) // No more tasks to feed
	}()

	// Retry handler
	go func() {
		for task := range retryChan {
			if task.Retries > 0 {
				retryWg.Add(1)
				go func(t Task) {
					defer retryWg.Done()
					t.Retries--
					taskChan <- t
				}(task)
			} else {
				log.Printf("🚫 Permanent failure: %s", task.Recipient.Email)
			}
		}
	}()

	wg.Wait()        // Wait for all workers to finish
	close(retryChan) // Close retry channel
	retryWg.Wait()   // Wait for all retry submissions
}
