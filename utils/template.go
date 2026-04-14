package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/parser"
)

const (
	maxCacheSize  = 100           // Maximum number of templates to cache
	cacheEntryTTL = 1 * time.Hour // Time-to-live for cached templates
)

type cachedTemplate struct {
	template *template.Template
	loadedAt time.Time
}

// TemplateCache is a bounded, TTL-based cache for parsed templates.
// Keys are canonical (absolute) file paths.
type TemplateCache struct {
	mu    sync.Mutex
	cache map[string]cachedTemplate
	order []string // Insertion order for FIFO eviction (LRU-approximation)
}

// Global template cache instance
var templateCache = &TemplateCache{
	cache: make(map[string]cachedTemplate),
}

// Get returns the cached template for path, or (nil, false) if absent or expired.
// Expired entries are removed on detection so they don't linger in the cache.
func (tc *TemplateCache) Get(path string) (*template.Template, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	entry, ok := tc.cache[path]
	if !ok {
		return nil, false
	}

	// Remove and report miss if entry has expired.
	if time.Since(entry.loadedAt) > cacheEntryTTL {
		delete(tc.cache, path)
		tc.removeFromOrder(path)
		return nil, false
	}

	return entry.template, true
}

// Set stores a parsed template under the given (already canonicalized) path.
// If the cache is full, the oldest entry is evicted first.
func (tc *TemplateCache) Set(path string, tmpl *template.Template) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// If path already exists, update in place without evicting.
	if _, exists := tc.cache[path]; exists {
		tc.cache[path] = cachedTemplate{template: tmpl, loadedAt: time.Now()}
		return
	}

	// Evict oldest entry if the cache is at capacity.
	if len(tc.cache) >= maxCacheSize && len(tc.order) > 0 {
		oldest := tc.order[0]
		delete(tc.cache, oldest)
		tc.order = tc.order[1:]
	}

	tc.cache[path] = cachedTemplate{template: tmpl, loadedAt: time.Now()}
	tc.order = append(tc.order, path)
}

// removeFromOrder removes a path from the ordering slice.
// Must be called with tc.mu held.
func (tc *TemplateCache) removeFromOrder(path string) {
	for i, p := range tc.order {
		if p == path {
			tc.order = append(tc.order[:i], tc.order[i+1:]...)
			return
		}
	}
}

// canonicalPath returns the absolute path, falling back to the original on error.
func canonicalPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// LoadTemplate parses and caches an HTML template file by its path.
// The path is canonicalized so "./template.html" and "template.html" share a cache entry.
func LoadTemplate(path string) (*template.Template, error) {
	abs := canonicalPath(path)

	if tmpl, ok := templateCache.Get(abs); ok {
		return tmpl, nil
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", path)
	}

	tmpl, err := template.ParseFiles(abs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	templateCache.Set(abs, tmpl)
	return tmpl, nil
}

// RenderTemplate renders an HTML template with the given recipient's data.
// Recipient fields are accessible in templates as {{ .email }}, {{ .name }}, etc.
func RenderTemplate(recipient parser.Recipient, templatePath string) (string, error) {
	tmpl, err := LoadTemplate(templatePath)
	if err != nil {
		return "", err
	}

	data := make(map[string]any)
	data["email"] = recipient.Email
	for key, value := range recipient.Data {
		data[key] = value
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return out.String(), nil
}
