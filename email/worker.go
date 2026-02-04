package email

import (
	"crypto/rand"
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"log"
	"math/big"
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

// SetMaxBackoff sets the maximum backoff delay for retries
func SetMaxBackoff(d time.Duration) {
	maxBackoff = d
}

// extractSMTPCode attempts to extract SMTP response code from error message
// Optimized to search only once and use efficient string operations
func extractSMTPCode(errMsg string) string {
	if len(errMsg) < 3 {
		return ""
	}

	// First check if the error message is long enough to contain a code
	// Fast path: check for codes at the beginning of the message
	prefix := errMsg[:3]
	// Common SMTP codes that indicate errors
	switch prefix {
	case "421", "450", "451", "452", "550", "551", "552", "553", "554":
		return prefix
	}

	// Fallback to contains search for embedded codes
	codes := []string{"421", "450", "451", "452", "550", "551", "552", "553", "554"}
	for _, code := range codes {
		if strings.Contains(errMsg, code) {
			return code
		}
	}

	return ""
}

// startWorker handles email sending using persistent SMTP connection and batch-mode dispatch.
func startWorker(w worker) {
	defer w.Wg.Done()

	// Check for context cancellation before starting
	if w.Ctx.Err() != nil {
		log.Printf("[Worker %d] Context cancelled before starting", w.ID)
		return
	}

	client, err := ConnectSMTPWithContext(w.Ctx, w.Config)
	if err != nil {
		log.Printf("[Worker %d] SMTP connection failed: %v", w.ID, err)
		return
	}
	defer func() {
		if err := client.Quit(); err != nil {
			log.Printf("[Worker %d] Failed to quit SMTP session: %v", w.ID, err)
		}
	}()

	// Pre-allocate batch slice to avoid reallocations
	batch := make([]Task, 0, w.BatchSize)

	for {
		select {
		case <-w.Ctx.Done():
			// Context cancelled, process remaining batch and exit
			if len(batch) > 0 {
				processBatch(w, client, batch)
			}
			log.Printf("[Worker %d] Context cancelled, stopping", w.ID)
			return

		case task, ok := <-w.TaskQueue:
			if !ok {
				// Channel closed, process remaining batch and exit
				if len(batch) > 0 {
					processBatch(w, client, batch)
				}
				return
			}

			batch = append(batch, task)

			if len(batch) >= w.BatchSize {
				processBatch(w, client, batch)
				batch = batch[:0] // Reset slice but keep capacity
			}
		}
	}
}

// processBatch handles the sending of a batch of emails with retry logic and async backoff.
func processBatch(w worker, client *smtp.Client, batch []Task) {
	for _, task := range batch {
		// Check for context cancellation before processing each task
		if w.Ctx.Err() != nil {
			log.Printf("[Worker %d] Context cancelled during batch processing", w.ID)
			return
		}

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

				backoff := time.Duration(1<<uint(task.Retries)) * time.Second
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				jitterMs, _ := rand.Int(rand.Reader, big.NewInt(1000))
				jitter := time.Duration(jitterMs.Int64()) * time.Millisecond
				delay := backoff + jitter

				log.Printf("[Worker %d] Retrying %s in %v (attempt %d)", w.ID, task.Recipient.Email, delay, task.Retries)

				// Update status to retry
				w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusRetry, duration, err.Error())

				w.RetryWg.Add(1)
				time.AfterFunc(delay, func() {
					defer w.RetryWg.Done()
					select {
					case w.RetryChan <- task:
					case <-w.Ctx.Done():
						log.Printf("[Worker %d] Retry cancelled for %s due to context", w.ID, task.Recipient.Email)
					}
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

			// Update offset tracker if available
			if w.Tracker != nil {
				newOffset := w.StartOffset + task.Index + 1
				w.Tracker.UpdateOffset(newOffset)
				// Periodically save offset (buffered writes for performance)
				if newOffset%10 == 0 {
					if err := w.Tracker.Save(); err != nil {
						log.Printf("️ Warning: Failed to save offset: %v", err)
					}
				}
			}
		}
	}
}
