package email

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTemplateCache(t *testing.T) {
	cache := NewTemplateCache(10*time.Minute, 50)
	
	if cache == nil {
		t.Fatal("NewTemplateCache returned nil")
	}
	
	if cache.maxAge != 10*time.Minute {
		t.Errorf("Expected maxAge 10m, got %v", cache.maxAge)
	}
	
	if cache.maxSize != 50 {
		t.Errorf("Expected maxSize 50, got %d", cache.maxSize)
	}
	
	// Clean up
	cache.Stop()
}

func TestNewTemplateCacheDefaultValues(t *testing.T) {
	cache := NewTemplateCache(0, 0)
	
	if cache.maxAge != 1*time.Hour {
		t.Errorf("Expected default maxAge 1h, got %v", cache.maxAge)
	}
	
	if cache.maxSize != 100 {
		t.Errorf("Expected default maxSize 100, got %d", cache.maxSize)
	}
	
	// Clean up
	cache.Stop()
}

func TestTemplateCache_GetValidTemplate(t *testing.T) {
	// Create a temporary template file
	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "test_template.html")
	
	templateContent := `<html><body>Hello {{.name}}!</body></html>`
	err := os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}
	
	cache := NewTemplateCache(10*time.Minute, 50)
	defer cache.Stop()
	
	// Test getting the template
	tmpl, err := cache.Get(templateFile)
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}
	
	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestTemplateCache_GetNonExistentTemplate(t *testing.T) {
	cache := NewTemplateCache(10*time.Minute, 50)
	defer cache.Stop()
	
	_, err := cache.Get("non_existent_template.html")
	if err == nil {
		t.Error("Expected error when getting non-existent template")
	}
}

func TestTemplateCache_Clear(t *testing.T) {
	cache := NewTemplateCache(10*time.Minute, 50)
	defer cache.Stop()
	
	// Add some dummy data
	cache.templates["test"] = nil
	cache.lastAccess["test"] = time.Now()
	
	if len(cache.templates) == 0 {
		t.Error("Templates map should have entries before clear")
	}
	
	cache.Clear()
	
	if len(cache.templates) != 0 {
		t.Error("Templates map should be empty after clear")
	}
	
	if len(cache.lastAccess) != 0 {
		t.Error("LastAccess map should be empty after clear")
	}
}

func TestNewAttachmentProcessor(t *testing.T) {
	maxSize := int64(10 * 1024 * 1024) // 10MB
	processor := NewAttachmentProcessor(maxSize)
	
	if processor == nil {
		t.Fatal("NewAttachmentProcessor returned nil")
	}
	
	if processor.maxSize != maxSize {
		t.Errorf("Expected maxSize %d, got %d", maxSize, processor.maxSize)
	}
	
	// Test that allowed types are set
	if len(processor.allowedTypes) == 0 {
		t.Error("Allowed types should not be empty")
	}
	
	// Test some expected allowed types
	expectedTypes := []string{"application/pdf", "image/jpeg", "image/png", "text/plain"}
	for _, expectedType := range expectedTypes {
		if _, ok := processor.allowedTypes[expectedType]; !ok {
			t.Errorf("Expected type '%s' to be allowed", expectedType)
		}
	}
}