package email

import (
	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/parser"
	"log"
	"sync"
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
	ID        int
	TaskQueue <-chan Task
	RetryChan chan<- Task
	Config    config.SMTPConfig
	Wg        *sync.WaitGroup
	RetryWg   *sync.WaitGroup
	BatchSize int
	Monitor   monitor.Monitor
}

// StartDispatcher spawns workers and processes email tasks with retries and batch-mode dispatch.
func StartDispatcher(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int) {
	StartDispatcherWithMonitor(tasks, cfg, concurrency, batchSize, monitor.NewNoOpMonitor())
}

// StartDispatcherWithMonitor spawns workers with monitoring support.
func StartDispatcherWithMonitor(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int, mon monitor.Monitor) {
	taskChan := make(chan Task, 1000)
	retryChan := make(chan Task, 500)

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
			ID:        i + 1,
			TaskQueue: taskChan,
			RetryChan: retryChan,
			Config:    cfg,
			Wg:        &wg,
			RetryWg:   &retryWg,
			BatchSize: batchSize,
			Monitor:   mon,
		})
	}

	// Dispatch initial tasks
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// Handle retries
	go func() {
		for task := range retryChan {
			if task.Retries > 0 {
				retryWg.Add(1)
				go func(t Task) {
					defer retryWg.Done()
					taskChan <- t
				}(task)
			} else {
				log.Printf("Permanent failure: %s", task.Recipient.Email)
			}
		}
	}()

	wg.Wait()
	close(retryChan)
	retryWg.Wait()
}
