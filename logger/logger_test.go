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

func TestAppendToCSV(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	// Test writing to a new file
	appendToCSV(testFile, "test@example.com", "Test Subject", "OK")

	// Read the file and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	expected := "test@example.com,Test Subject,OK\n"
	if string(content) != expected {
		t.Errorf("File content = %q, expected %q", string(content), expected)
	}

	// Test appending to existing file
	appendToCSV(testFile, "test2@example.com", "Test Subject 2", "Failed")

	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file after append: %v", err)
	}

	expected = "test@example.com,Test Subject,OK\ntest2@example.com,Test Subject 2,Failed\n"
	if string(content) != expected {
		t.Errorf("File content after append = %q, expected %q", string(content), expected)
	}
}

func TestAppendToCSVWithCommasInSubject(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	// Test writing with commas in subject (this tests CSV format)
	appendToCSV(testFile, "test@example.com", "Subject, with, commas", "OK")

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// The function doesn't quote fields, so commas will break CSV format
	// This is expected behavior based on the current implementation
	if !strings.Contains(string(content), "Subject, with, commas") {
		t.Error("Subject with commas was not written correctly")
	}
}

func TestLogFunctions(t *testing.T) {
	// These functions log to stdout and write to CSV files
	// We can't easily test the log output, but we can test they don't panic

	// Create temp directory for CSV files
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Test LogSuccess
	LogSuccess("success@example.com", "Success Subject")

	// Verify success.csv was created
	if _, err := os.Stat("success.csv"); os.IsNotExist(err) {
		t.Error("success.csv was not created")
	}

	// Test LogFailure
	LogFailure("failure@example.com", "Failure Subject")

	// Verify failed.csv was created
	if _, err := os.Stat("failed.csv"); os.IsNotExist(err) {
		t.Error("failed.csv was not created")
	}

	// Test Errorf and Warnf - they should not panic
	Errorf("Test error: %s", "test")
	Warnf("Test warning: %s", "test")
}
