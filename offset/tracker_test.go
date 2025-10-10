package offset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTracker(t *testing.T) {
	tracker := NewTracker("")
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	if tracker.filePath != DefaultOffsetFile {
		t.Errorf("Expected default file path %s, got %s", DefaultOffsetFile, tracker.filePath)
	}

	if tracker.offset != 0 {
		t.Errorf("Expected initial offset 0, got %d", tracker.offset)
	}
}

func TestTracker_LoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewTracker(filepath.Join(tmpDir, "nonexistent.offset"))

	err := tracker.Load()
	if err != nil {
		t.Errorf("Load should not fail for non-existent file: %v", err)
	}

	if tracker.GetOffset() != 0 {
		t.Errorf("Expected offset 0 for non-existent file, got %d", tracker.GetOffset())
	}
}

func TestTracker_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	offsetFile := filepath.Join(tmpDir, "test.offset")

	tracker := NewTracker(offsetFile)
	tracker.SetJobID("test-job-123")
	tracker.UpdateOffset(42)

	// Save the offset
	err := tracker.Save()
	if err != nil {
		t.Fatalf("Failed to save offset: %v", err)
	}

	// Create new tracker and load
	tracker2 := NewTracker(offsetFile)
	err = tracker2.Load()
	if err != nil {
		t.Fatalf("Failed to load offset: %v", err)
	}

	if tracker2.GetOffset() != 42 {
		t.Errorf("Expected loaded offset 42, got %d", tracker2.GetOffset())
	}

	if tracker2.GetJobID() != "test-job-123" {
		t.Errorf("Expected loaded job ID 'test-job-123', got '%s'", tracker2.GetJobID())
	}
}

func TestTracker_BackwardCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	offsetFile := filepath.Join(tmpDir, "test.offset")

	// Create old format file (just offset number)
	err := os.WriteFile(offsetFile, []byte("25\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tracker := NewTracker(offsetFile)
	err = tracker.Load()
	if err != nil {
		t.Fatalf("Failed to load old format file: %v", err)
	}

	if tracker.GetOffset() != 25 {
		t.Errorf("Expected offset 25, got %d", tracker.GetOffset())
	}

	if tracker.GetJobID() != "" {
		t.Errorf("Expected empty job ID for old format, got '%s'", tracker.GetJobID())
	}
}

func TestTracker_Reset(t *testing.T) {
	tmpDir := t.TempDir()
	offsetFile := filepath.Join(tmpDir, "test.offset")

	tracker := NewTracker(offsetFile)
	tracker.SetJobID("test-job")
	tracker.UpdateOffset(10)

	err := tracker.Save()
	if err != nil {
		t.Fatalf("Failed to save offset: %v", err)
	}

	// Reset should clear everything
	err = tracker.Reset()
	if err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}

	if tracker.GetOffset() != 0 {
		t.Errorf("Expected offset 0 after reset, got %d", tracker.GetOffset())
	}

	if tracker.GetJobID() != "" {
		t.Errorf("Expected empty job ID after reset, got '%s'", tracker.GetJobID())
	}

	// File should be removed
	if _, err := os.Stat(offsetFile); !os.IsNotExist(err) {
		t.Error("Offset file should be removed after reset")
	}
}

func TestTracker_ShouldSkip(t *testing.T) {
	tracker := NewTracker("")
	tracker.UpdateOffset(5)

	tests := []struct {
		index    int
		expected bool
	}{
		{0, true},   // Should skip (0 < 5)
		{3, true},   // Should skip (3 < 5)
		{4, true},   // Should skip (4 < 5)
		{5, false},  // Should not skip (5 >= 5)
		{10, false}, // Should not skip (10 >= 5)
	}

	for _, test := range tests {
		result := tracker.ShouldSkip(test.index)
		if result != test.expected {
			t.Errorf("ShouldSkip(%d) = %v, expected %v (offset=%d)",
				test.index, result, test.expected, tracker.GetOffset())
		}
	}
}

func TestTracker_UpdateOffset(t *testing.T) {
	tracker := NewTracker("")

	// Initial state
	if tracker.GetOffset() != 0 {
		t.Errorf("Expected initial offset 0, got %d", tracker.GetOffset())
	}

	// Update offset
	tracker.UpdateOffset(15)
	if tracker.GetOffset() != 15 {
		t.Errorf("Expected offset 15, got %d", tracker.GetOffset())
	}

	// Update with same value should work
	tracker.UpdateOffset(15)
	if tracker.GetOffset() != 15 {
		t.Errorf("Expected offset 15 after same update, got %d", tracker.GetOffset())
	}

	// Update with higher value
	tracker.UpdateOffset(20)
	if tracker.GetOffset() != 20 {
		t.Errorf("Expected offset 20, got %d", tracker.GetOffset())
	}
}

func TestTracker_GetInfo(t *testing.T) {
	tracker := NewTracker("")
	tracker.SetJobID("test-job-456")
	tracker.UpdateOffset(30)

	info := tracker.GetInfo()
	if info.JobID != "test-job-456" {
		t.Errorf("Expected job ID 'test-job-456', got '%s'", info.JobID)
	}

	if info.Offset != 30 {
		t.Errorf("Expected offset 30, got %d", info.Offset)
	}
}

func TestTracker_AtomicSave(t *testing.T) {
	tmpDir := t.TempDir()
	offsetFile := filepath.Join(tmpDir, "atomic.offset")

	tracker := NewTracker(offsetFile)
	tracker.SetJobID("atomic-test")
	tracker.UpdateOffset(100)

	// Save should be atomic
	err := tracker.Save()
	if err != nil {
		t.Fatalf("Failed to save offset: %v", err)
	}

	// Temporary file should not exist
	tempFile := offsetFile + ".tmp"
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temporary file should not exist after atomic save")
	}

	// Final file should exist and be readable
	content, err := os.ReadFile(offsetFile)
	if err != nil {
		t.Fatalf("Failed to read offset file: %v", err)
	}

	expected := "atomic-test:100\n"
	if string(content) != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, string(content))
	}
}