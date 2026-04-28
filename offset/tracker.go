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

// Tracker manages email delivery offset tracking for resumable campaigns.
//
// Concurrency model:
//   - GetOffset / ShouldSkip / GetJobID / GetInfo: read-only; safe from any goroutine.
//   - UpdateOffset / SetJobID: writer methods used by the CLI before/after dispatch.
//   - MarkComplete: hot-path call from worker goroutines. Maintains a contiguous
//     high-water mark so the persisted offset advances only past indices that
//     have all been completed (correct under concurrent out-of-order completion).
//   - Save: serializes to disk; the file write happens outside the lock so
//     readers and MarkComplete callers do not block on I/O.
type Tracker struct {
	mu       sync.Mutex
	filePath string
	offset   int // contiguous high-water mark: all indices < offset are completed
	jobID    string
	dirty    bool

	// pending holds completed indices that arrived out of order (i.e. with
	// idx > offset at the time of MarkComplete). The map is drained as the
	// contiguous prefix advances.
	pending map[int]struct{}
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
	}
}

// Load reads the offset from the file
func (t *Tracker) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	file, err := os.Open(t.filePath)
	if err != nil {
		if os.IsNotExist(err) {
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

		parts := strings.Split(line, ":")
		switch len(parts) {
		case 2:
			t.jobID = parts[0]
			if offset, err := strconv.Atoi(parts[1]); err == nil {
				t.offset = offset
			} else {
				t.offset = 0
				t.jobID = ""
				t.dirty = true
				return fmt.Errorf("invalid offset format in file: %s (resetting to start)", line)
			}
		case 1:
			if offset, err := strconv.Atoi(parts[0]); err == nil {
				t.offset = offset
				t.jobID = ""
			} else {
				t.offset = 0
				t.jobID = ""
				t.dirty = true
				return fmt.Errorf("invalid offset format in file: %s (resetting to start)", line)
			}
		default:
			t.offset = 0
			t.jobID = ""
			t.dirty = true
			return fmt.Errorf("invalid offset format in file: %s (resetting to start)", line)
		}
	}

	return scanner.Err()
}

// GetOffset returns the current offset (contiguous high-water mark).
func (t *Tracker) GetOffset() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.offset
}

// GetJobID returns the current job ID
func (t *Tracker) GetJobID() string {
	t.mu.Lock()
	defer t.mu.Unlock()
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

// UpdateOffset sets the offset directly. Used by the CLI to seed the tracker
// from a persisted state and by tests; production worker code should use
// MarkComplete instead. Calling UpdateOffset clears any pending out-of-order
// completions because they no longer correspond to a meaningful baseline.
func (t *Tracker) UpdateOffset(newOffset int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.offset != newOffset {
		t.offset = newOffset
		t.dirty = true
	}
	t.pending = nil
}

// MarkComplete records that the absolute task index `idx` has been delivered.
// The persisted offset advances to one past the highest index N such that all
// indices in [t.offset, N] have been marked complete. Indices that arrive out
// of order are buffered in t.pending and drained when the gap closes.
//
// Indices that arrive below the current offset (e.g. duplicates from a retry
// after a flusher has already moved past) are ignored.
func (t *Tracker) MarkComplete(idx int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if idx < t.offset {
		return
	}
	if idx == t.offset {
		t.offset++
		if t.pending != nil {
			for {
				if _, ok := t.pending[t.offset]; !ok {
					break
				}
				delete(t.pending, t.offset)
				t.offset++
			}
			if len(t.pending) == 0 {
				t.pending = nil
			}
		}
		t.dirty = true
		return
	}

	if t.pending == nil {
		t.pending = make(map[int]struct{})
	}
	t.pending[idx] = struct{}{}
}

// Save writes the current offset to disk atomically (write to temp + rename).
// The disk I/O happens outside the tracker lock so concurrent MarkComplete
// callers are not blocked on fsync.
func (t *Tracker) Save() error {
	t.mu.Lock()
	if !t.dirty {
		t.mu.Unlock()
		return nil
	}
	offset := t.offset
	jobID := t.jobID
	filePath := t.filePath
	t.dirty = false
	t.mu.Unlock()

	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.markDirty()
			return fmt.Errorf("failed to create offset directory: %w", err)
		}
	}

	tempFile := filePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		t.markDirty()
		return fmt.Errorf("failed to create temporary offset file: %w", err)
	}

	var content string
	if jobID != "" {
		content = fmt.Sprintf("%s:%d\n", jobID, offset)
	} else {
		content = fmt.Sprintf("%d\n", offset)
	}

	if _, err := file.WriteString(content); err != nil {
		file.Close()
		os.Remove(tempFile)
		t.markDirty()
		return fmt.Errorf("failed to write offset: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempFile)
		t.markDirty()
		return fmt.Errorf("failed to sync offset file: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(tempFile)
		t.markDirty()
		return fmt.Errorf("failed to close offset file: %w", err)
	}
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile)
		t.markDirty()
		return fmt.Errorf("failed to rename offset file: %w", err)
	}
	return nil
}

// markDirty re-flags the tracker so a future Save retries. Used when a Save
// fails after we've already cleared the dirty bit; without this, a transient
// I/O error would silently lose the next offset advance.
func (t *Tracker) markDirty() {
	t.mu.Lock()
	t.dirty = true
	t.mu.Unlock()
}

// Reset clears the offset and removes the offset file
func (t *Tracker) Reset() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.offset = 0
	t.jobID = ""
	t.dirty = false
	t.pending = nil

	if err := os.Remove(t.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove offset file: %w", err)
	}
	return nil
}

// GetInfo returns current offset information
func (t *Tracker) GetInfo() OffsetInfo {
	t.mu.Lock()
	defer t.mu.Unlock()
	return OffsetInfo{
		JobID:  t.jobID,
		Offset: t.offset,
	}
}

// ShouldSkip returns true if the given index should be skipped based on offset
func (t *Tracker) ShouldSkip(index int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return index < t.offset
}
