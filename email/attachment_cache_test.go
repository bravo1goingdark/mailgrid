package email

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAttachmentCache_GetCachesPayload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	want := []byte("hello world this is a test attachment")
	if err := os.WriteFile(path, want, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cache := NewAttachmentCache(0)

	first, err := cache.Get(path)
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if first == nil {
		t.Fatal("first Get returned nil entry")
	}

	second, err := cache.Get(path)
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if second != first {
		t.Fatalf("expected cache hit to return the same entry pointer")
	}

	expectedLen := base64.StdEncoding.EncodedLen(len(want))
	if len(first.data) != expectedLen {
		t.Fatalf("expected encoded length %d, got %d", expectedLen, len(first.data))
	}

	dec, err := base64.StdEncoding.DecodeString(string(first.data))
	if err != nil {
		t.Fatalf("decode cached payload: %v", err)
	}
	if !bytes.Equal(dec, want) {
		t.Fatalf("decoded payload mismatch:\n got %q\nwant %q", dec, want)
	}
}

func TestAttachmentCache_RespectsCap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.bin")
	// 200-byte file → ~272 bytes base64-encoded; cap below that forces fallback.
	if err := os.WriteFile(path, make([]byte, 200), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cache := NewAttachmentCache(64) // small cap

	if _, err := cache.Get(path); err == nil {
		t.Fatal("expected error when payload exceeds cap")
	}
}

func TestAttachmentCache_ConcurrentGetIsRaceFree(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cache := NewAttachmentCache(0)

	const N = 64
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			if _, err := cache.Get(path); err != nil {
				t.Errorf("concurrent Get: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestAttachmentCache_Reset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cache := NewAttachmentCache(0)
	if _, err := cache.Get(path); err != nil {
		t.Fatalf("Get: %v", err)
	}
	cache.Reset()
	if cache.totalSz != 0 {
		t.Fatalf("expected totalSz 0 after reset, got %d", cache.totalSz)
	}
	if len(cache.entries) != 0 {
		t.Fatalf("expected empty entries after reset, got %d", len(cache.entries))
	}
}
