package email

import (
	"log"
	"math/rand"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/monitor"
)

var (
	retryLimit = 2
	maxBackoff = 10 * time.Second // maximum wait before retry
	retryMu    sync.RWMutex
)

// SetRetryLimit sets the max retry attempts per failed email, with exponential backoff.
func SetRetryLimit(limit int) {
	retryMu.Lock()
	defer retryMu.Unlock()
	retryLimit = limit
}

// SetMaxBackoff sets the maximum backoff delay for retries
func SetMaxBackoff(d time.Duration) {
	retryMu.Lock()
	defer retryMu.Unlock()
	maxBackoff = d
}

// GetRetryLimit returns the current retry limit (thread-safe)
func GetRetryLimit() int {
	retryMu.RLock()
	defer retryMu.RUnlock()
	return retryLimit
}

// GetMaxBackoff returns the current max backoff duration (thread-safe)
func GetMaxBackoff() time.Duration {
	retryMu.RLock()
	defer retryMu.RUnlock()
	return maxBackoff
}

// retryDelay computes the backoff duration for a given attempt number.
// Caps the bit-shift at 8 to prevent integer overflow (max 2^8=256s before clamp).
func retryDelay(attempt int) time.Duration {
	retryMu.RLock()
	cap := maxBackoff
	retryMu.RUnlock()

	shift := attempt
	if shift < 0 {
		shift = 0
	}
	if shift > 8 {
		shift = 8
	}
	d := time.Duration(1<<uint(shift)) * time.Second
	if d > cap {
		d = cap
	}
	// Add up to 1000ms of jitter using a non-crypto source (fine for backoff).
	d += time.Duration(rand.Intn(1000)) * time.Millisecond
	return d
}

// extractSMTPCode attempts to extract SMTP response code from error message.
func extractSMTPCode(errMsg string) string {
	if len(errMsg) < 3 {
		return ""
	}

	// Fast path: code at the start of the message
	prefix := errMsg[:3]
	switch prefix {
	case "421", "450", "451", "452", "550", "551", "552", "553", "554":
		return prefix
	}

	// Fallback: search for embedded codes
	codes := []string{"421", "450", "451", "452", "550", "551", "552", "553", "554"}
	for _, code := range codes {
		if strings.Contains(errMsg, code) {
			return code
		}
	}
	return ""
}

// isConnectionError checks if the error indicates a connection issue requiring reconnection.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())

	connectionErrors := []string{
		"eof",
		"connection reset",
		"use of closed network connection",
		"broken pipe",
		"network is unreachable",
		"no route to host",
		"i/o timeout",
		"temporary failure",
	}
	for _, pattern := range connectionErrors {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// SMTP codes that indicate connection-level failures
	code := extractSMTPCode(err.Error())
	for _, c := range []string{"421", "451", "554"} {
		if code == c {
			return true
		}
	}
	return false
}

// startWorker handles email sending using a persistent SMTP connection and batch-mode dispatch.
// Retries are handled inline (sleep + retry within the worker), avoiding the goroutine leak
// and channel-close race conditions of the previous time.AfterFunc approach.
func startWorker(w worker) {
	defer w.Wg.Done()

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

	batch := make([]Task, 0, w.BatchSize)

	for {
		select {
		case <-w.Ctx.Done():
			if len(batch) > 0 {
				processBatch(w, &client, batch)
			}
			log.Printf("[Worker %d] Context cancelled, stopping", w.ID)
			return

		case task, ok := <-w.TaskQueue:
			if !ok {
				if len(batch) > 0 {
					processBatch(w, &client, batch)
				}
				return
			}

			batch = append(batch, task)
			if len(batch) >= w.BatchSize {
				processBatch(w, &client, batch)
				batch = batch[:0]
			}
		}
	}
}

// processBatch sends a batch of tasks with inline retry logic.
// On failure the worker sleeps for the backoff duration and retries up to retryLimit
// times — no goroutines or channels needed, which eliminates all prior race conditions.
func processBatch(w worker, clientPtr **smtp.Client, batch []Task) {
	currentLimit := GetRetryLimit()

	for _, task := range batch {
		if w.Ctx.Err() != nil {
			log.Printf("[Worker %d] Context cancelled during batch processing", w.ID)
			return
		}

		// Inner retry loop: attempt → fail → sleep → retry (inline)
		for {
			start := time.Now()
			w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusSending, 0, "")

			err := SendWithClient(*clientPtr, w.Config, task, w.AttachmentCache)

			// Reconnect on connection errors and retry once immediately
			if isConnectionError(err) {
				log.Printf("[Worker %d] Connection error, reconnecting: %v", w.ID, err)
				if quitErr := (*clientPtr).Quit(); quitErr != nil {
					log.Printf("[Worker %d] Quit failed: %v", w.ID, quitErr)
				}
				newClient, reconnErr := ConnectSMTPWithContext(w.Ctx, w.Config)
				if reconnErr != nil {
					log.Printf("[Worker %d] Reconnection failed: %v", w.ID, reconnErr)
				} else {
					*clientPtr = newClient
					start = time.Now()
					err = SendWithClient(*clientPtr, w.Config, task, w.AttachmentCache)
				}
			}

			duration := time.Since(start)

			if err == nil {
				logger.LogSuccess(task.Recipient.Email, task.Subject)
				w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusSent, duration, "")
				w.Monitor.AddSMTPResponse("250")
				w.Sent.Add(1)

				// Mark this index complete; the tracker maintains a contiguous
				// high-water mark so resume is correct under concurrency. The
				// disk write is coalesced by the dispatcher's flusher.
				if w.Tracker != nil {
					w.Tracker.MarkComplete(task.Index)
				}
				break
			}

			// Send failed
			log.Printf("[Worker %d] Failed to send to %s: %v", w.ID, task.Recipient.Email, err)
			if code := extractSMTPCode(err.Error()); code != "" {
				w.Monitor.AddSMTPResponse(code)
			} else {
				w.Monitor.AddSMTPResponse("error")
			}

			if task.Retries >= currentLimit {
				// Exhausted retries — permanent failure
				logger.LogFailure(task.Recipient.Email, task.Subject)
				w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusFailed, duration, err.Error())
				w.Failed.Add(1)
				break
			}

			// Schedule inline retry: increment counter, sleep with backoff, loop
			task.Retries++
			delay := retryDelay(task.Retries)
			log.Printf("[Worker %d] Retrying %s in %v (attempt %d/%d)", w.ID, task.Recipient.Email, delay, task.Retries, currentLimit)
			w.Monitor.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusRetry, duration, err.Error())

			select {
			case <-time.After(delay):
				// Continue to next iteration of retry loop
			case <-w.Ctx.Done():
				log.Printf("[Worker %d] Context cancelled during retry wait for %s", w.ID, task.Recipient.Email)
				return
			}
		}
	}
}
