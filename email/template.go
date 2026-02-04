package email

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TemplateCache provides thread-safe caching of parsed email templates
type TemplateCache struct {
	mu sync.RWMutex
	// Map of template hash to parsed template
	templates map[string]*template.Template
	// Map of template hash to last access time
	lastAccess map[string]time.Time
	// Maximum age of cached templates
	maxAge time.Duration
	// Maximum number of templates to cache
	maxSize int
	// Current number of cached templates
	currentSize int
	// Channel to stop cleanup goroutine
	stopCh chan struct{}
}

// NewTemplateCache creates a new template cache with given configuration
func NewTemplateCache(maxAge time.Duration, maxSize int) *TemplateCache {
	if maxAge <= 0 {
		maxAge = 1 * time.Hour
	}
	if maxSize <= 0 {
		maxSize = 100
	}

	c := &TemplateCache{
		templates:  make(map[string]*template.Template),
		lastAccess: make(map[string]time.Time),
		maxAge:     maxAge,
		maxSize:    maxSize,
		currentSize: 0,
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.periodicCleanup()

	return c
}

// Get retrieves a template from cache or parses it if not found
func (c *TemplateCache) Get(path string) (*template.Template, error) {
	// Calculate template hash
	hash, err := c.hashFile(path)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	c.mu.RLock()
	if tmpl, ok := c.templates[hash]; ok {
		c.lastAccess[hash] = time.Now()
		c.mu.RUnlock()
		return tmpl, nil
	}
	c.mu.RUnlock()

	// Parse template
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.mu.Lock()
	defer c.mu.Unlock()

	// Enforce size limit BEFORE adding new template
	if c.currentSize >= c.maxSize {
		c.evictOldest()
	}

	// Double-check after eviction in case multiple threads are racing
	if c.currentSize < c.maxSize {
		c.templates[hash] = tmpl
		c.lastAccess[hash] = time.Now()
		c.currentSize++
	}

	return tmpl, nil
}

// Clear removes all templates from cache
func (c *TemplateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates = make(map[string]*template.Template)
	c.lastAccess = make(map[string]time.Time)
	c.currentSize = 0
}

// Stop stops cleanup goroutine and cleans up resources
func (c *TemplateCache) Stop() {
	close(c.stopCh)
	c.Clear()
}

// GetCurrentSize returns the current number of cached templates
func (c *TemplateCache) GetCurrentSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

func (c *TemplateCache) hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func (c *TemplateCache) periodicCleanup() {
	ticker := time.NewTicker(c.maxAge / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

func (c *TemplateCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var hashesToRemove []string

	// Find expired templates
	for hash, lastAccess := range c.lastAccess {
		if now.Sub(lastAccess) > c.maxAge {
			hashesToRemove = append(hashesToRemove, hash)
		}
	}

	// Remove expired templates and update size
	for _, hash := range hashesToRemove {
		delete(c.templates, hash)
		delete(c.lastAccess, hash)
		c.currentSize--
	}
}

func (c *TemplateCache) evictOldest() {
	var oldestHash string
	var oldestTime time.Time

	// Find oldest template
	first := true
	for hash, accessTime := range c.lastAccess {
		if first || accessTime.Before(oldestTime) {
			oldestHash = hash
			oldestTime = accessTime
			first = false
		}
	}

	// Remove it
	if oldestHash != "" {
		delete(c.templates, oldestHash)
		delete(c.lastAccess, oldestHash)
		c.currentSize--
	}
}

// AttachmentProcessor handles efficient processing of email attachments
type AttachmentProcessor struct {
	// Pool of buffers for copying attachment data
	bufferPool sync.Pool
	// Maximum size of a single attachment
	maxSize int64
	// List of allowed MIME types
	allowedTypes map[string]struct{}
}

// NewAttachmentProcessor creates a new attachment processor
func NewAttachmentProcessor(maxSize int64) *AttachmentProcessor {
	return &AttachmentProcessor{
		bufferPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 32*1024) // 32KB chunks
				return &buf
			},
		},
		maxSize: maxSize,
		allowedTypes: map[string]struct{}{
			"application/pdf":           {},
			"image/jpeg":                {},
			"image/png":                 {},
			"image/gif":                 {},
			"text/plain":                {},
			"text/plain; charset=utf-8": {},
			"application/zip":           {},
			"application/octet-stream":  {},
		},
	}
}

// ProcessAttachment handles a single attachment efficiently
func (p *AttachmentProcessor) ProcessAttachment(path string) (io.Reader, string, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}

	// Check size
	if info.Size() > p.maxSize {
		return nil, "", ErrAttachmentTooLarge
	}

	// Check MIME type
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		// Try to detect type
		file, err := os.Open(path)
		if err != nil {
			return nil, "", err
		}
		defer file.Close()

		// Read first 512 bytes for MIME detection
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, "", err
		}
		mimeType = http.DetectContentType(buffer)
	}

	if _, ok := p.allowedTypes[mimeType]; !ok {
		return nil, "", ErrUnsupportedAttachmentType
	}

	// Create efficient reader
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}

	// Use buffered reader from pool
	buf := p.bufferPool.Get().(*[]byte)
	reader := bufio.NewReaderSize(file, len(*buf))

	// Create a closer that returns buffer to pool
	closer := &attachmentReadCloser{
		Reader: reader,
		closer: func() error {
			p.bufferPool.Put(buf)
			return file.Close()
		},
	}

	return closer, mimeType, nil
}

type attachmentReadCloser struct {
	*bufio.Reader
	closer func() error
}

func (r *attachmentReadCloser) Close() error {
	return r.closer()
}

var (
	ErrAttachmentTooLarge        = errors.New("attachment exceeds maximum allowed size")
	ErrUnsupportedAttachmentType = errors.New("unsupported attachment type")
)
