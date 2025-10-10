package email

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/bravo1goingdark/mailgrid/email"
)

func TestNewOffsetTracker(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}

	// Should start with 0 sent emails
	if count := tracker.GetSentCount(); count != 0 {
		t.Errorf("Expected 0 sent emails, got %d", count)
	}
}

func TestOffsetTracker_MarkEmailSent(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Mark first email as sent
	err = tracker.MarkEmailSent("alice@example.com")
	if err != nil {
		t.Fatalf("Failed to mark email as sent: %v", err)
	}

	// Check it's marked as sent
	if !tracker.IsEmailSent("alice@example.com") {
		t.Error("Expected alice@example.com to be marked as sent")
	}

	// Check count
	if count := tracker.GetSentCount(); count != 1 {
		t.Errorf("Expected 1 sent email, got %d", count)
	}

	// Mark more emails
	emails := []string{"bob@example.com", "carol@example.com"}
	for _, email := range emails {
		err = tracker.MarkEmailSent(email)
		if err != nil {
			t.Fatalf("Failed to mark email %s as sent: %v", email, err)
		}
	}

	// Check all are marked as sent
	allEmails := append([]string{"alice@example.com"}, emails...)
	for _, email := range allEmails {
		if !tracker.IsEmailSent(email) {
			t.Errorf("Expected %s to be marked as sent", email)
		}
	}

	if count := tracker.GetSentCount(); count != 3 {
		t.Errorf("Expected 3 sent emails, got %d", count)
	}
}

func TestOffsetTracker_IsEmailSent(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Email should not be sent initially
	if tracker.IsEmailSent("test@example.com") {
		t.Error("Expected test@example.com to not be marked as sent initially")
	}

	// Mark as sent
	err = tracker.MarkEmailSent("test@example.com")
	if err != nil {
		t.Fatalf("Failed to mark email as sent: %v", err)
	}

	// Now should be sent
	if !tracker.IsEmailSent("test@example.com") {
		t.Error("Expected test@example.com to be marked as sent after marking")
	}

	// Other emails should still not be sent
	if tracker.IsEmailSent("other@example.com") {
		t.Error("Expected other@example.com to not be marked as sent")
	}
}

func TestOffsetTracker_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	// Create a pre-existing offset file
	content := "alice@example.com\nbob@example.com\ncarol@example.com\n"
	err := os.WriteFile(offsetFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test offset file: %v", err)
	}

	// Create tracker - should load existing data
	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Should have loaded 3 emails
	if count := tracker.GetSentCount(); count != 3 {
		t.Errorf("Expected 3 sent emails after loading, got %d", count)
	}

	// Check each email is marked as sent
	expectedEmails := []string{"alice@example.com", "bob@example.com", "carol@example.com"}
	for _, email := range expectedEmails {
		if !tracker.IsEmailSent(email) {
			t.Errorf("Expected %s to be marked as sent after loading", email)
		}
	}
}

func TestOffsetTracker_Reset(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Mark some emails as sent
	emails := []string{"alice@example.com", "bob@example.com"}
	for _, email := range emails {
		err = tracker.MarkEmailSent(email)
		if err != nil {
			t.Fatalf("Failed to mark email as sent: %v", err)
		}
	}

	// Force flush to create the file
	err = tracker.Flush()
	if err != nil {
		t.Fatalf("Failed to flush tracker: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(offsetFile); os.IsNotExist(err) {
		t.Error("Expected offset file to exist after flush")
	}

	// Reset
	err = tracker.Reset()
	if err != nil {
		t.Fatalf("Failed to reset tracker: %v", err)
	}

	// Should have 0 sent emails
	if count := tracker.GetSentCount(); count != 0 {
		t.Errorf("Expected 0 sent emails after reset, got %d", count)
	}

	// Emails should no longer be marked as sent
	for _, email := range emails {
		if tracker.IsEmailSent(email) {
			t.Errorf("Expected %s to not be marked as sent after reset", email)
		}
	}

	// File should be removed
	if _, err := os.Stat(offsetFile); !os.IsNotExist(err) {
		t.Error("Expected offset file to be removed after reset")
	}
}

func TestOffsetTracker_Flush(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Mark emails as sent (less than buffer size to test manual flush)
	emails := []string{"alice@example.com", "bob@example.com"}
	for _, email := range emails {
		err = tracker.MarkEmailSent(email)
		if err != nil {
			t.Fatalf("Failed to mark email as sent: %v", err)
		}
	}

	// File should not exist yet (buffered)
	if _, err := os.Stat(offsetFile); !os.IsNotExist(err) {
		t.Error("Expected offset file to not exist before flush")
	}

	// Manual flush
	err = tracker.Flush()
	if err != nil {
		t.Fatalf("Failed to flush tracker: %v", err)
	}

	// File should now exist
	if _, err := os.Stat(offsetFile); os.IsNotExist(err) {
		t.Error("Expected offset file to exist after flush")
	}

	// Read file content
	content, err := os.ReadFile(offsetFile)
	if err != nil {
		t.Fatalf("Failed to read offset file: %v", err)
	}

	expectedContent := "alice@example.com\nbob@example.com\n"
	if string(content) != expectedContent {
		t.Errorf("Expected file content %q, got %q", expectedContent, string(content))
	}
}

func TestOffsetTracker_BufferedWrites(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Mark 15 emails as sent (more than buffer size of 10)
	for i := 1; i <= 15; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		err = tracker.MarkEmailSent(email)
		if err != nil {
			t.Fatalf("Failed to mark email as sent: %v", err)
		}
	}

	// File should exist after buffer overflow
	if _, err := os.Stat(offsetFile); os.IsNotExist(err) {
		t.Error("Expected offset file to exist after buffer overflow")
	}

	// Should have at least 10 emails written (first batch)
	content, err := os.ReadFile(offsetFile)
	if err != nil {
		t.Fatalf("Failed to read offset file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 10 {
		t.Errorf("Expected at least 10 lines in offset file, got %d", len(lines))
	}
}

func TestOffsetTracker_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	// Test concurrent writes
	var wg sync.WaitGroup
	numWorkers := 10
	emailsPerWorker := 20

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < emailsPerWorker; i++ {
				email := fmt.Sprintf("worker%d-email%d@example.com", workerID, i)
				err := tracker.MarkEmailSent(email)
				if err != nil {
					t.Errorf("Worker %d failed to mark email as sent: %v", workerID, err)
				}
			}
		}(worker)
	}

	wg.Wait()

	// Check all emails were marked correctly
	expectedCount := numWorkers * emailsPerWorker
	if count := tracker.GetSentCount(); count != expectedCount {
		t.Errorf("Expected %d sent emails, got %d", expectedCount, count)
	}

	// Test concurrent reads while writing
	var readWg sync.WaitGroup
	readWg.Add(1)
	go func() {
		defer readWg.Done()
		for i := 0; i < 100; i++ {
			_ = tracker.GetSentCount()
			_ = tracker.IsEmailSent("test@example.com")
		}
	}()

	// Continue writing while reading
	for i := 0; i < 10; i++ {
		email := fmt.Sprintf("concurrent%d@example.com", i)
		err := tracker.MarkEmailSent(email)
		if err != nil {
			t.Errorf("Failed to mark email as sent during concurrent test: %v", err)
		}
	}

	readWg.Wait()
}

func TestOffsetTracker_GetSentEmails(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	expectedEmails := []string{"alice@example.com", "bob@example.com", "carol@example.com"}

	// Mark emails as sent
	for _, email := range expectedEmails {
		err = tracker.MarkEmailSent(email)
		if err != nil {
			t.Fatalf("Failed to mark email as sent: %v", err)
		}
	}

	// Get sent emails
	sentEmails := tracker.GetSentEmails()

	if len(sentEmails) != len(expectedEmails) {
		t.Errorf("Expected %d sent emails, got %d", len(expectedEmails), len(sentEmails))
	}

	// Check all expected emails are present (order doesn't matter)
	sentEmailsMap := make(map[string]bool)
	for _, email := range sentEmails {
		sentEmailsMap[email] = true
	}

	for _, email := range expectedEmails {
		if !sentEmailsMap[email] {
			t.Errorf("Expected email %s to be in sent emails list", email)
		}
	}
}

func TestOffsetTracker_DuplicateEmails(t *testing.T) {
	tempDir := t.TempDir()
	offsetFile := filepath.Join(tempDir, "test.offset")

	tracker, err := email.NewOffsetTracker(offsetFile)
	if err != nil {
		t.Fatalf("Failed to create offset tracker: %v", err)
	}
	defer tracker.Close()

	email := "test@example.com"

	// Mark same email multiple times
	for i := 0; i < 5; i++ {
		err = tracker.MarkEmailSent(email)
		if err != nil {
			t.Fatalf("Failed to mark email as sent: %v", err)
		}
	}

	// Should only count once
	if count := tracker.GetSentCount(); count != 1 {
		t.Errorf("Expected 1 sent email after marking duplicates, got %d", count)
	}

	sentEmails := tracker.GetSentEmails()
	if len(sentEmails) != 1 {
		t.Errorf("Expected 1 unique email in sent list, got %d", len(sentEmails))
	}
}