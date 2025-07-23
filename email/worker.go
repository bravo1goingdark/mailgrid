package email

import (
	"log"
	"mailgrid/logger"
	"math/rand"
	"net/smtp"
	"time"
)

var retryLimit = 2
var maxBackoff = 10 * time.Second // maximum wait before retry

// SetRetryLimit sets the max retry attempts per failed email, with exponential backoff.
func SetRetryLimit(limit int) {
	retryLimit = limit
}

// startWorker handles email sending using persistent SMTP connection and batch-mode dispatch.
func startWorker(w worker) {
	defer w.Wg.Done()

	client, err := ConnectSMTP(w.Config)
	if err != nil {
		log.Printf("[Worker %d] SMTP connection failed: %v", w.ID, err)
		return
	}
	defer func() {
		if err := client.Quit(); err != nil {
			log.Printf("[Worker %d] Failed to quit SMTP session: %v", w.ID, err)
		}
	}()

	batch := make([]Task, 0, w.BatchSize)

	for {
		task, ok := <-w.TaskQueue
		if !ok {
			if len(batch) > 0 {
				processBatch(w, client, batch)
			}
			return
		}

		batch = append(batch, task)

		if len(batch) >= w.BatchSize {
			processBatch(w, client, batch)
			batch = batch[:0]
		}
	}
}

// processBatch handles the sending of a batch of emails with retry logic and async backoff.
func processBatch(w worker, client *smtp.Client, batch []Task) {
	for _, task := range batch {
		err := SendWithClient(client, w.Config, task)
		if err != nil {
			log.Printf("[Worker %d] Failed to send to %s: %v", w.ID, task.Recipient.Email, err)

			if task.Retries < retryLimit {
				task.Retries++

				backoff := time.Duration(1<<task.Retries) * time.Second
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				delay := backoff + jitter

				log.Printf("[Worker %d] Retrying %s in %v (attempt %d)", w.ID, task.Recipient.Email, delay, task.Retries)

				w.RetryWg.Add(1)
				time.AfterFunc(delay, func() {
					defer w.RetryWg.Done()
					w.RetryChan <- task
				})
			} else {
				logger.LogFailure(task.Recipient.Email, task.Subject)
			}
		} else {
			logger.LogSuccess(task.Recipient.Email, task.Subject)
		}
	}
}
