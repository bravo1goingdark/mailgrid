package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	logger := New("test")
	if logger == nil {
		t.Error("New() returned nil")
	}

	// Test that the logger implements the required interface
	logger.Infof("Test info message: %s", "test")
	logger.Warnf("Test warn message: %s", "test")
	logger.Errorf("Test error message: %s", "test")
}

// TestCSVLogger verifies that csvLogger writes and flushes correctly.
func TestCSVLogger(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	l, err := newCSVLogger(testFile)
	if err != nil {
		t.Fatalf("newCSVLogger() error: %v", err)
	}
	defer l.close()

	l.write("test@example.com", "Test Subject", "OK")
	l.flush()

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	expected := "test@example.com,Test Subject,OK\n"
	if string(content) != expected {
		t.Errorf("File content = %q, expected %q", string(content), expected)
	}

	// Append a second entry
	l.write("test2@example.com", "Test Subject 2", "Failed")
	l.flush()

	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file after append: %v", err)
	}

	expected = "test@example.com,Test Subject,OK\ntest2@example.com,Test Subject 2,Failed\n"
	if string(content) != expected {
		t.Errorf("File content after append = %q, expected %q", string(content), expected)
	}
}

// TestCSVLoggerWithCommasInSubject verifies that subjects with commas are written as-is.
func TestCSVLoggerWithCommasInSubject(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	l, err := newCSVLogger(testFile)
	if err != nil {
		t.Fatalf("newCSVLogger() error: %v", err)
	}
	defer l.close()

	l.write("test@example.com", "Subject, with, commas", "OK")
	l.flush()

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// The logger writes raw CSV without quoting; commas in subject fields are unescaped.
	if !strings.Contains(string(content), "Subject, with, commas") {
		t.Error("Subject with commas was not written correctly")
	}
}

func TestLogFunctions(t *testing.T) {
	// These functions log to stdout and write to CSV files.
	// Run in a temp directory so we don't pollute the working tree.
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	// Reset package-level loggers so they initialise fresh in tmpDir.
	func() {
		loggerMu.Lock()
		defer loggerMu.Unlock()
		if successLogger != nil {
			successLogger.close()
			successLogger = nil
		}
		if failedLogger != nil {
			failedLogger.close()
			failedLogger = nil
		}
	}()
	os.Chdir(tmpDir)
	defer func() {
		os.Chdir(originalDir)
		// Reset again after the test so other tests start clean.
		loggerMu.Lock()
		defer loggerMu.Unlock()
		if successLogger != nil {
			successLogger.close()
			successLogger = nil
		}
		if failedLogger != nil {
			failedLogger.close()
			failedLogger = nil
		}
	}()

	// Test LogSuccess
	LogSuccess("success@example.com", "Success Subject")
	FlushAndClose()

	// Verify success.csv was created and flushed
	if _, err := os.Stat("success.csv"); os.IsNotExist(err) {
		t.Error("success.csv was not created")
	}

	// Reset so LogFailure gets a fresh handle in the same dir
	loggerMu.Lock()
	successLogger = nil
	failedLogger = nil
	loggerMu.Unlock()

	// Test LogFailure
	LogFailure("failure@example.com", "Failure Subject")
	FlushAndClose()

	// Verify failed.csv was created
	if _, err := os.Stat("failed.csv"); os.IsNotExist(err) {
		t.Error("failed.csv was not created")
	}

	// Test Errorf and Warnf - they should not panic
	Errorf("Test error: %s", "test")
	Warnf("Test warning: %s", "test")
}
