package email

import (
	"encoding/base64"
	"errors"
	"mime"
	"os"
	"path/filepath"
	"sync"
)

// DefaultAttachmentCacheLimit caps the total bytes of base64-encoded attachment
// payloads held in cache per dispatch. Files whose encoded size would exceed
// the remaining budget are streamed inline instead.
const DefaultAttachmentCacheLimit int64 = 256 << 20 // 256 MiB

// cachedAttachment holds a fully MIME-prepared attachment ready for direct
// reuse across recipients within the same dispatch run.
type cachedAttachment struct {
	mimeType string
	safeName string
	data     []byte // base64-encoded payload (no line breaks, matches base64.StdEncoding)
}

// AttachmentCache memoizes base64-encoded attachment payloads keyed by absolute
// path. Population is lazy and concurrency-safe. A single shared cache should
// be created per dispatch and torn down with Reset when the run ends.
type AttachmentCache struct {
	mu       sync.RWMutex
	entries  map[string]*cachedAttachment
	totalSz  int64
	maxBytes int64
}

// NewAttachmentCache constructs a cache with the given total-size cap.
// A non-positive cap falls back to DefaultAttachmentCacheLimit.
func NewAttachmentCache(maxBytes int64) *AttachmentCache {
	if maxBytes <= 0 {
		maxBytes = DefaultAttachmentCacheLimit
	}
	return &AttachmentCache{
		entries:  make(map[string]*cachedAttachment),
		maxBytes: maxBytes,
	}
}

// Get returns the cached attachment for path, populating the cache on first
// access. Returns (nil, errCacheCapExceeded) when the file would push the
// cache past its byte cap; callers should fall back to streaming.
func (c *AttachmentCache) Get(path string) (*cachedAttachment, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	c.mu.RLock()
	if e, ok := c.entries[abs]; ok {
		c.mu.RUnlock()
		return e, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[abs]; ok {
		return e, nil
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	encodedLen := base64.StdEncoding.EncodedLen(len(raw))
	if int64(encodedLen)+c.totalSz > c.maxBytes {
		return nil, errCacheCapExceeded
	}

	encoded := make([]byte, encodedLen)
	base64.StdEncoding.Encode(encoded, raw)

	mt := mime.TypeByExtension(filepath.Ext(abs))
	if mt == "" {
		mt = "application/octet-stream"
	}

	entry := &cachedAttachment{
		mimeType: mt,
		safeName: sanitizeFilename(path),
		data:     encoded,
	}
	c.entries[abs] = entry
	c.totalSz += int64(encodedLen)
	return entry, nil
}

// Reset releases all cached payloads. Call once when the dispatch run ends.
func (c *AttachmentCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cachedAttachment)
	c.totalSz = 0
}

var errCacheCapExceeded = errors.New("attachment cache size limit exceeded")
