package logger

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureLogOutput(fn func()) string {
	var buf bytes.Buffer
	oldOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOutput)

	fn()
	return buf.String()
}

func TestLogSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	email := "test@example.com"
	subject := "Test Subject"

	// Capture log output
	output := captureLogOutput(func() {
		LogSuccess(email, subject)
	})

	// Check log output
	assert.Contains(t, output, "Sent to test@example.com")

	// Check CSV file
	csvPath := filepath.Join(tmpDir, "success.csv")
	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	expectedCSV := "test@example.com,Test Subject,OK\n"
	assert.Equal(t, expectedCSV, string(content))
}

func TestLogFailure(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	email := "failed@example.com"
	subject := "Failed Subject"

	// Capture log output
	output := captureLogOutput(func() {
		LogFailure(email, subject)
	})

	// Check log output
	assert.Contains(t, output, "Failed permanently: failed@example.com")

	// Check CSV file
	csvPath := filepath.Join(tmpDir, "failed.csv")
	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	expectedCSV := "failed@example.com,Failed Subject,Failed\n"
	assert.Equal(t, expectedCSV, string(content))
}

func TestAppendToCSV_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")

	// Add multiple entries
	appendToCSV(csvPath, "user1@example.com", "Subject 1", "OK")
	appendToCSV(csvPath, "user2@example.com", "Subject 2", "Failed")
	appendToCSV(csvPath, "user3@example.com", "Subject 3", "OK")

	// Read file content
	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 3)
	assert.Equal(t, "user1@example.com,Subject 1,OK", lines[0])
	assert.Equal(t, "user2@example.com,Subject 2,Failed", lines[1])
	assert.Equal(t, "user3@example.com,Subject 3,OK", lines[2])
}

func TestAppendToCSV_FileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "new_file.csv")

	// File should not exist initially
	_, err := os.Stat(csvPath)
	assert.True(t, os.IsNotExist(err))

	// Append to CSV should create the file
	appendToCSV(csvPath, "test@example.com", "Test", "OK")

	// File should now exist
	_, err = os.Stat(csvPath)
	assert.NoError(t, err)

	// Check content
	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com,Test,OK\n", string(content))
}

func TestAppendToCSV_InvalidPath(t *testing.T) {
	// Test with invalid path - should log error but not panic
	output := captureLogOutput(func() {
		appendToCSV("/invalid/path/test.csv", "test@example.com", "Test", "OK")
	})

	assert.Contains(t, output, "Could not write to log file")
}

func TestErrorf(t *testing.T) {
	output := captureLogOutput(func() {
		Errorf("This is an error: %s", "test error")
	})

	assert.Contains(t, output, "ERROR: This is an error: test error")
}

func TestWarnf(t *testing.T) {
	output := captureLogOutput(func() {
		Warnf("This is a warning: %d", 42)
	})

	assert.Contains(t, output, "WARNING: This is a warning: 42")
}

func TestErrorf_NoFormatting(t *testing.T) {
	output := captureLogOutput(func() {
		Errorf("Simple error message")
	})

	assert.Contains(t, output, "ERROR: Simple error message")
}

func TestWarnf_NoFormatting(t *testing.T) {
	output := captureLogOutput(func() {
		Warnf("Simple warning message")
	})

	assert.Contains(t, output, "WARNING: Simple warning message")
}

func TestCSVContentEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "escape_test.csv")

	// Test with commas and special characters in content
	email := "test@example.com"
	subject := "Subject with, comma and \"quotes\""
	status := "OK"

	appendToCSV(csvPath, email, subject, status)

	content, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	expected := "test@example.com,Subject with, comma and \"quotes\",OK\n"
	assert.Equal(t, expected, string(content))
}