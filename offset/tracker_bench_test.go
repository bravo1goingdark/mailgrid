package offset

import (
	"path/filepath"
	"testing"
)

// BenchmarkTracker_MarkComplete measures the per-success cost of advancing
// the contiguous high-water mark in memory. This is the hot-path call from
// every worker; it must stay sub-microsecond to keep up with bulk SMTP
// throughput (10k+ sends/sec).
func BenchmarkTracker_MarkComplete(b *testing.B) {
	tracker := NewTracker("")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.MarkComplete(i)
	}
}

// BenchmarkTracker_PerSuccessSave reproduces the old per-success persistence
// behavior — UpdateOffset followed by Save on every successful send. This is
// the worst case the new flusher coalesces away, and the gap between the two
// benchmarks is the real per-send win when offset tracking is enabled.
func BenchmarkTracker_PerSuccessSave(b *testing.B) {
	dir := b.TempDir()
	tracker := NewTracker(filepath.Join(dir, "bench.offset"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.UpdateOffset(i + 1)
		if err := tracker.Save(); err != nil {
			b.Fatalf("save: %v", err)
		}
	}
}
