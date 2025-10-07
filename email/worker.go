package email

import (
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"log"
	"math/rand"
	"net/smtp"
	"strings"
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
		start := time.Now()

		// Update status to sending
		w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusSending, 0, "")

		err := SendWithClient(client, w.Config, task)
		duration := time.Since(start)

		if err != nil {
			log.Printf("[Worker %d] Failed to send to %s: %v", w.ID, task.Recipient.Email, err)

			// Extract SMTP response code from error message
			if code := extractSMTPCode(err.Error()); code != "" {
				w.Monitor.AddSMTPResponse(code)
			} else {
				w.Monitor.AddSMTPResponse("error")
			}

			if task.Retries < retryLimit {
				task.Retries++

				backoff := time.Duration(1<<task.Retries) * time.Second
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				delay := backoff + jitter

				log.Printf("[Worker %d] Retrying %s in %v (attempt %d)", w.ID, task.Recipient.Email, delay, task.Retries)

				// Update status to retry
				w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusRetry, duration, err.Error())

				w.RetryWg.Add(1)
				time.AfterFunc(delay, func() {
					defer w.RetryWg.Done()
					w.RetryChan <- task
				})
			} else {
				logger.LogFailure(task.Recipient.Email, task.Subject)
				// Update status to failed
				w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusFailed, duration, err.Error())
			}
		} else {
			logger.LogSuccess(task.Recipient.Email, task.Subject)
			// Update status to sent
			w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusSent, duration, "")
			w.Monitor.AddSMTPResponse("250") // Standard success code
		}
	}
}

// extractSMTPCode attempts to extract SMTP response code from error message
func extractSMTPCode(errMsg string) string {
	// Common SMTP error patterns
	if strings.Contains(errMsg, "421") {
		return "421"
	}
	if strings.Contains(errMsg, "450") {
		return "450"
	}
	if strings.Contains(errMsg, "451") {
		return "451"
	}
	if strings.Contains(errMsg, "452") {
		return "452"
	}
	if strings.Contains(errMsg, "550") {
		return "550"
	}
	if strings.Contains(errMsg, "551") {
		return "551"
	}
	if strings.Contains(errMsg, "552") {
		return "552"
	}
	if strings.Contains(errMsg, "553") {
		return "553"
	}
	if strings.Contains(errMsg, "554") {
		return "554"
	}
	return ""
}
