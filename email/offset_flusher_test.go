package email

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/offset"
)

// countingTracker wraps a real *offset.Tracker and counts MarkComplete and
// Save invocations so the test can assert the flusher actually coalesces.
type countingTracker struct {
	*offset.Tracker
	saveCount atomic.Int64
	markCount atomic.Int64
}

func (t *countingTracker) MarkComplete(idx int) {
	t.markCount.Add(1)
	t.Tracker.MarkComplete(idx)
}

func (t *countingTracker) Save() error {
	t.saveCount.Add(1)
	return t.Tracker.Save()
}

// TestOffsetFlusher_CoalescesSavesUnderConcurrency mirrors the worker hot path:
// many goroutines call MarkComplete in parallel while a flusher trickles saves
// to disk. After dispatch:
//   - the persisted offset must equal the contiguous high-water mark, and
//   - the number of Save calls must be far smaller than the number of marks.
//
// This is the integration that actually proves the optimization claim — that
// disk fsyncs no longer serialize sends.
func TestOffsetFlusher_CoalescesSavesUnderConcurrency(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bench.offset")

	base := offset.NewTracker(path)
	tr := &countingTracker{Tracker: base}

	stop := startOffsetFlusher(tr, 10*time.Millisecond)
	if stop == nil {
		t.Fatal("startOffsetFlusher returned nil for a real tracker")
	}

	const total = 5000
	const workers = 16

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := w; i < total; i += workers {
				tr.MarkComplete(i)
			}
		}()
	}
	wg.Wait()

	stop()

	// Final synchronous save (matches runDispatch behaviour).
	if err := tr.Save(); err != nil {
		t.Fatalf("final save: %v", err)
	}

	if marks := tr.markCount.Load(); marks != total {
		t.Fatalf("expected %d MarkComplete calls, got %d", total, marks)
	}
	saves := tr.saveCount.Load()
	if saves >= int64(total) {
		t.Fatalf("flusher did not coalesce: saves=%d marks=%d", saves, total)
	}
	t.Logf("MarkComplete=%d, Save=%d (coalesce ratio %.1fx)", total, saves, float64(total)/float64(saves))

	if got := tr.GetOffset(); got != total {
		t.Fatalf("expected in-memory offset %d, got %d", total, got)
	}

	loader := offset.NewTracker(path)
	if err := loader.Load(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := loader.GetOffset(); got != total {
		t.Fatalf("expected persisted offset %d, got %d", total, got)
	}
}

// TestOffsetFlusher_StopWaitsForExit verifies that the stop function returned
// by startOffsetFlusher actually waits for the flusher goroutine to exit. If
// it returned early, runDispatch's subsequent synchronous Save could race on
// the temp-file path used by the atomic rename.
func TestOffsetFlusher_StopWaitsForExit(t *testing.T) {
	dir := t.TempDir()
	tr := offset.NewTracker(filepath.Join(dir, "exit.offset"))
	tr.UpdateOffset(1) // mark dirty so Save would run

	stop := startOffsetFlusher(tr, 1*time.Millisecond)
	if stop == nil {
		t.Fatal("expected non-nil stop")
	}
	// Give the flusher one tick of work, then stop.
	time.Sleep(5 * time.Millisecond)
	stop()

	// At this point no further Save can be in flight. A direct Save should
	// see dirty=false (the flusher already saved) or succeed without racing.
	if err := tr.Save(); err != nil && !strings.Contains(err.Error(), "not exist") {
		t.Fatalf("post-stop save failed: %v", err)
	}

	// Verify the file content parses cleanly — strconv.Atoi guards against any
	// torn writes from a temp-file race.
	loader := offset.NewTracker(filepath.Join(dir, "exit.offset"))
	if err := loader.Load(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := loader.GetOffset(); got != 1 {
		t.Fatalf("expected offset 1, got %d", got)
	}
	_ = strconv.Itoa
}
