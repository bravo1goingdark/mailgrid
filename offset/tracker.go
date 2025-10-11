package offset

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const DefaultOffsetFile = ".mailgrid.offset"

// Tracker manages email delivery offset tracking for resumable campaigns
type Tracker struct {
	mu       sync.RWMutex
	filePath string
	offset   int
	jobID    string
	dirty    bool // Track if offset needs to be written
}

// OffsetInfo contains information about the current offset state
type OffsetInfo struct {
	JobID  string `json:"job_id"`
	Offset int    `json:"offset"`
	Total  int    `json:"total,omitempty"`
}

// NewTracker creates a new offset tracker
func NewTracker(filePath string) *Tracker {
	if filePath == "" {
		filePath = DefaultOffsetFile
	}

	return &Tracker{
		filePath: filePath,
		offset:   0,
		dirty:    false,
	}
}

// Load reads the offset from the file
func (t *Tracker) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	file, err := os.Open(t.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start from beginning
			t.offset = 0
			t.jobID = ""
			return nil
		}
		return fmt.Errorf("failed to open offset file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			t.offset = 0
			t.jobID = ""
			return nil
		}

		// Parse line format: "jobID:offset" or just "offset" for backward compatibility
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			t.jobID = parts[0]
			if offset, err := strconv.Atoi(parts[1]); err == nil {
				t.offset = offset
			} else {
				return fmt.Errorf("invalid offset format in file: %s", line)
			}
		} else if len(parts) == 1 {
			// Backward compatibility: just offset number
			if offset, err := strconv.Atoi(parts[0]); err == nil {
				t.offset = offset
				t.jobID = ""
			} else {
				return fmt.Errorf("invalid offset format in file: %s", line)
			}
		} else {
			return fmt.Errorf("invalid offset format in file: %s", line)
		}
	}

	return scanner.Err()
}

// GetOffset returns the current offset
func (t *Tracker) GetOffset() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.offset
}

// GetJobID returns the current job ID
func (t *Tracker) GetJobID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.jobID
}

// SetJobID sets the job ID for this campaign
func (t *Tracker) SetJobID(jobID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.jobID != jobID {
		t.jobID = jobID
		t.dirty = true
	}
}

// UpdateOffset updates the offset and marks it for saving
func (t *Tracker) UpdateOffset(newOffset int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.offset != newOffset {
		t.offset = newOffset
		t.dirty = true
	}
}

// Save writes the current offset to file (buffered writes)
func (t *Tracker) Save() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.dirty {
		return nil // No changes to save
	}

	// Ensure directory exists
	dir := filepath.Dir(t.filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create offset directory: %w", err)
		}
	}

	// Write to temporary file first for atomic operation
	tempFile := t.filePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary offset file: %w", err)
	}

	var content string
	if t.jobID != "" {
		content = fmt.Sprintf("%s:%d\n", t.jobID, t.offset)
	} else {
		content = fmt.Sprintf("%d\n", t.offset)
	}

	if _, err := file.WriteString(content); err != nil {
		file.Close()
		os.Remove(tempFile)
		return fmt.Errorf("failed to write offset: %w", err)
	}

	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempFile)
		return fmt.Errorf("failed to sync offset file: %w", err)
	}

	file.Close()

	// Atomic rename
	if err := os.Rename(tempFile, t.filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename offset file: %w", err)
	}

	t.dirty = false
	return nil
}

// Reset clears the offset and removes the offset file
func (t *Tracker) Reset() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.offset = 0
	t.jobID = ""
	t.dirty = false

	// Remove the offset file
	if err := os.Remove(t.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove offset file: %w", err)
	}

	return nil
}

// GetInfo returns current offset information
func (t *Tracker) GetInfo() OffsetInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return OffsetInfo{
		JobID:  t.jobID,
		Offset: t.offset,
	}
}

// ShouldSkip returns true if the given index should be skipped based on offset
func (t *Tracker) ShouldSkip(index int) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return index < t.offset
}
