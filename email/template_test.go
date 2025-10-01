package email

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestTemplateCache_Basic(t *testing.T) {
	cache := NewTemplateCache(1*time.Hour, 10)
	defer cache.Clear()

	// Create a temporary template file
	content := "Hello {{.Name}}!"
	tmpfile, err := ioutil.TempFile("", "template*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test cache miss (first load)
	tmpl1, err := cache.Get(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}
	if tmpl1 == nil {
		t.Fatal("Template is nil")
	}

	// Test cache hit (second load)
	tmpl2, err := cache.Get(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to get cached template: %v", err)
	}
	if tmpl2 != tmpl1 {
		t.Error("Expected same template instance from cache")
	}
}

func TestTemplateCache_Expiration(t *testing.T) {
	cache := NewTemplateCache(10*time.Millisecond, 10)
	defer cache.Clear()

	// Create a temporary template file
	tmpfile, err := ioutil.TempFile("", "template*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load template
	_, err = cache.Get(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Template should still be accessible but might be refreshed
	_, err = cache.Get(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
}

func TestTemplateCache_SizeLimit(t *testing.T) {
	cache := NewTemplateCache(1*time.Hour, 2) // Small cache size
	defer cache.Clear()

	files := make([]*os.File, 3)
	for i := range files {
		tmpfile, err := ioutil.TempFile("", "template*.html")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		
		if _, err := tmpfile.Write([]byte("test")); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}
		files[i] = tmpfile
	}

	// Load templates to exceed cache size
	for _, f := range files {
		_, err := cache.Get(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}

	// Cache should have evicted old entries
	cache.mu.RLock()
	size := len(cache.templates)
	cache.mu.RUnlock()
	
	if size > 2 {
		t.Errorf("Cache size %d exceeds limit of 2", size)
	}
}

func TestAttachmentProcessor_Basic(t *testing.T) {
	processor := NewAttachmentProcessor(1024 * 1024) // 1MB limit

	// Create a test file
	tmpfile, err := ioutil.TempFile("", "attachment*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "This is a test attachment"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test processing
	reader, mimeType, err := processor.ProcessAttachment(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to process attachment: %v", err)
	}

	if reader == nil {
		t.Fatal("Reader is nil")
	}

	if mimeType == "" {
		t.Fatal("MIME type is empty")
	}

	// Clean up
	if closer, ok := reader.(interface{ Close() error }); ok {
		closer.Close()
	}
}

func TestAttachmentProcessor_SizeLimit(t *testing.T) {
	processor := NewAttachmentProcessor(10) // 10 byte limit

	// Create a file larger than limit
	tmpfile, err := ioutil.TempFile("", "large*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "This is definitely more than 10 bytes"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Should fail due to size limit
	_, _, err = processor.ProcessAttachment(tmpfile.Name())
	if err != ErrAttachmentTooLarge {
		t.Errorf("Expected ErrAttachmentTooLarge, got %v", err)
	}
}