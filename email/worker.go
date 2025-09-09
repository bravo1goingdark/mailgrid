package email

import (
	"github.com/bravo1goingdark/mailgrid/logger"
	"log"
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
	defer func() {
		w.Wg.Done()
		w.Metrics.RecordWorkerStop()
	}()

	w.Metrics.RecordWorkerStart()
	log.Printf("[Worker %d] Starting with batch size %d", w.ID, w.BatchSize)

	client, err := ConnectSMTP(w.Config)
	if err != nil {
		log.Printf("[Worker %d] SMTP connection failed: %v", w.ID, err)
		w.Metrics.RecordError("smtp_connection_failed")
		return
	}
	w.Metrics.RecordSMTPConnection()
	
	defer func() {
		w.Metrics.RecordSMTPDisconnection()
		if err := client.Quit(); err != nil {
			log.Printf("[Worker %d] Failed to quit SMTP session: %v", w.ID, err)
		}
	}()

	batch := make([]Task, 0, w.BatchSize)
	batchTimeout := time.NewTimer(30 * time.Second) // Flush batch after 30s
	defer batchTimeout.Stop()

	for {
		select {
		case task, ok := <-w.TaskQueue:
			if !ok {
				// Channel closed, process remaining batch
				if len(batch) > 0 {
					processBatch(w, client, batch)
				}
				return
			}

			batch = append(batch, task)

			if len(batch) >= w.BatchSize {
				processBatch(w, client, batch)
				batch = batch[:0]
				batchTimeout.Reset(30 * time.Second)
			}
			
		case <-batchTimeout.C:
			// Flush partial batch on timeout
			if len(batch) > 0 {
				log.Printf("[Worker %d] Flushing partial batch of %d emails due to timeout", w.ID, len(batch))
				processBatch(w, client, batch)
				batch = batch[:0]
			}
			batchTimeout.Reset(30 * time.Second)
			
		case <-w.Ctx.Done():
			log.Printf("[Worker %d] Shutting down due to context cancellation", w.ID)
			// Process remaining batch before shutdown
			if len(batch) > 0 {
				log.Printf("[Worker %d] Processing final batch of %d emails", w.ID, len(batch))
				processBatch(w, client, batch)
			}
			return
		}
	}
}

// processBatch handles the sending of a batch of emails with retry logic, rate limiting, and metrics.
func processBatch(w worker, client *smtp.Client, batch []Task) {
	log.Printf("[Worker %d] Processing batch of %d emails", w.ID, len(batch))
	start := time.Now()
	
	for i, task := range batch {
		// Apply rate limiting
		if w.RateLimiter != nil {
			if err := w.RateLimiter.Wait(w.Ctx); err != nil {
				log.Printf("[Worker %d] Rate limiter cancelled: %v", w.ID, err)
				return
			}
		}
		
		// Record attempt
		start := time.Now()
		err := SendWithClient(client, w.Config, task)
		duration := time.Since(start)
		w.Metrics.RecordResponseTime("email_send", duration)
		
		if err != nil {
			log.Printf("[Worker %d] Failed to send to %s: %v", w.ID, task.Recipient.Email, err)
			w.Metrics.RecordError("email_send_failed")

			if task.Retries < retryLimit {
				task.Retries++
				w.Metrics.RecordEmailRetried()

				// Exponential backoff with jitter
				backoff := time.Duration(1<<task.Retries) * time.Second
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				delay := backoff + jitter

				log.Printf("[Worker %d] Retrying %s in %v (attempt %d/%d)", w.ID, task.Recipient.Email, delay, task.Retries, retryLimit)

				// Use a cancellable timer instead of time.AfterFunc
				w.RetryWg.Add(1)
				go func(t Task) {
					defer w.RetryWg.Done()
					timer := time.NewTimer(delay)
					defer timer.Stop()
					
					select {
					case <-timer.C:
						select {
						case w.RetryChan <- t:
						case <-w.Ctx.Done():
							log.Printf("[Worker %d] Retry cancelled for %s", w.ID, t.Recipient.Email)
						}
					case <-w.Ctx.Done():
						log.Printf("[Worker %d] Retry timer cancelled for %s", w.ID, t.Recipient.Email)
					}
				}(task)
			} else {
				log.Printf("[Worker %d] Permanent failure after %d attempts: %s", w.ID, task.Retries, task.Recipient.Email)
				logger.LogFailure(task.Recipient.Email, task.Subject)
				w.Metrics.RecordEmailFailed()
			}
		} else {
			log.Printf("[Worker %d] ✅ Successfully sent to %s (%d/%d in batch)", w.ID, task.Recipient.Email, i+1, len(batch))
			logger.LogSuccess(task.Recipient.Email, task.Subject)
			w.Metrics.RecordEmailSent()
		}
	}
	
	batchDuration := time.Since(start)
	w.Metrics.RecordResponseTime("batch_process", batchDuration)
	log.Printf("[Worker %d] Completed batch of %d emails in %v", w.ID, len(batch), batchDuration)
}
