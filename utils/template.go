package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
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
type TemplateCache struct {
	mu    sync.RWMutex
	cache map[string]cachedTemplate
	order []string // Track insertion order for LRU eviction
}

// Global template cache instance
var templateCache = &TemplateCache{
	cache: make(map[string]cachedTemplate),
	order: []string{},
}

func (tc *TemplateCache) Get(path string) (*template.Template, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	entry, ok := tc.cache[path]
	if !ok {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.loadedAt) > cacheEntryTTL {
		return nil, false
	}

	return entry.template, true
}

func (tc *TemplateCache) Set(path string, tmpl *template.Template) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if cache is full and needs eviction
	if len(tc.cache) >= maxCacheSize && tc.cache[path] == (cachedTemplate{}) {
		// Evict oldest entry (LRU)
		if len(tc.order) > 0 {
			oldest := tc.order[0]
			delete(tc.cache, oldest)
			tc.order = tc.order[1:]
		}
	}

	tc.cache[path] = cachedTemplate{
		template: tmpl,
		loadedAt: time.Now(),
	}

	// Update order for LRU
	found := false
	for i, p := range tc.order {
		if p == path {
			tc.order = append(tc.order[:i], tc.order[i+1:]...)
			found = true
			break
		}
	}
	tc.order = append(tc.order, path)
	_ = found // suppress unused warning
}

// LoadTemplate parses and caches an HTML template file by its path.
//
// If the template has been parsed before and is not expired, it returns the cached version.
// Otherwise, it loads and parses the template and stores it in memory.
func LoadTemplate(path string) (*template.Template, error) {
	if tmpl, ok := templateCache.Get(path); ok {
		return tmpl, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", path)
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	templateCache.Set(path, tmpl)
	return tmpl, nil
}

// RenderTemplate renders an HTML template with given recipient's data.
//
// It uses caching internally for performance. Recipient values can be accessed in template via:
//   - {{ .email }} for recipient email
//   - {{ .name }}, {{ .age }}, etc. for CSV fields (same as subject templates)
func RenderTemplate(recipient parser.Recipient, templatePath string) (string, error) {
	tmpl, err := LoadTemplate(templatePath)
	if err != nil {
		return "", err
	}

	// Flatten data structure for consistent template access
	// Include email field and all CSV data fields at top level
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
