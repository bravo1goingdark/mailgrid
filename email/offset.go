package email

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// OffsetTracker manages resumable email delivery by tracking successfully sent emails
type OffsetTracker struct {
	filePath      string
	sentEmails    map[string]bool
	pendingWrites []string
	mutex         sync.RWMutex
	bufferSize    int
	file          *os.File
}

// NewOffsetTracker creates a new offset tracker
func NewOffsetTracker(filePath string) (*OffsetTracker, error) {
	tracker := &OffsetTracker{
		filePath:      filePath,
		sentEmails:    make(map[string]bool),
		pendingWrites: make([]string, 0),
		bufferSize:    10, // Buffer 10 successful sends before writing to disk
	}

	// Load existing offset data if file exists
	if err := tracker.loadFromFile(); err != nil {
		return nil, fmt.Errorf("failed to load offset data: %w", err)
	}

	return tracker, nil
}

// loadFromFile reads the offset file and populates the sentEmails map
func (o *OffsetTracker) loadFromFile() error {
	file, err := os.Open(o.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, which is fine for first run
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		email := scanner.Text()
		if email != "" {
			o.sentEmails[email] = true
		}
	}

	return scanner.Err()
}

// IsEmailSent checks if an email has already been successfully sent
func (o *OffsetTracker) IsEmailSent(email string) bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.sentEmails[email]
}

// MarkEmailSent marks an email as successfully sent and buffers the write
func (o *OffsetTracker) MarkEmailSent(email string) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Don't mark duplicates
	if o.sentEmails[email] {
		return nil
	}

	o.sentEmails[email] = true
	o.pendingWrites = append(o.pendingWrites, email)

	// Flush to disk if buffer is full
	if len(o.pendingWrites) >= o.bufferSize {
		return o.flushToDisk()
	}

	return nil
}

// flushToDisk writes pending emails to the offset file (caller must hold mutex)
func (o *OffsetTracker) flushToDisk() error {
	if len(o.pendingWrites) == 0 {
		return nil
	}

	// Open file in append mode for atomic writes
	file, err := os.OpenFile(o.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open offset file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, email := range o.pendingWrites {
		if _, err := writer.WriteString(email + "\n"); err != nil {
			return fmt.Errorf("failed to write email to offset file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush offset file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync offset file: %w", err)
	}

	// Clear pending writes
	o.pendingWrites = o.pendingWrites[:0]
	return nil
}

// Flush forces all pending writes to disk
func (o *OffsetTracker) Flush() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.flushToDisk()
}

// Reset clears the offset file and in-memory state
func (o *OffsetTracker) Reset() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Clear in-memory state
	o.sentEmails = make(map[string]bool)
	o.pendingWrites = o.pendingWrites[:0]

	// Remove the offset file
	if err := os.Remove(o.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove offset file: %w", err)
	}

	return nil
}

// GetSentCount returns the number of emails marked as sent
func (o *OffsetTracker) GetSentCount() int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return len(o.sentEmails)
}

// GetSentEmails returns a copy of all sent email addresses
func (o *OffsetTracker) GetSentEmails() []string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	emails := make([]string, 0, len(o.sentEmails))
	for email := range o.sentEmails {
		emails = append(emails, email)
	}
	return emails
}

// Close flushes any pending writes and closes the tracker
func (o *OffsetTracker) Close() error {
	return o.Flush()
}