package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetrics_Basic(t *testing.T) {
	m := NewMetrics()

	// Test initial state
	if m.TotalEmailsSent != 0 {
		t.Error("Initial emails sent should be 0")
	}

	// Test recording success
	m.RecordEmailSent(100 * time.Millisecond)
	if m.TotalEmailsSent != 1 {
		t.Error("Expected 1 email sent")
	}

	// Test recording failure
	testErr := errors.New("test error")
	m.RecordEmailFailed(testErr)
	if m.TotalEmailsFailed != 1 {
		t.Error("Expected 1 email failed")
	}

	// Check error was recorded
	if count := m.ErrorCounts["test error"]; count != 1 {
		t.Errorf("Expected error count 1, got %d", count)
	}
}

func TestMetrics_Connections(t *testing.T) {
	m := NewMetrics()

	// Test successful connection
	m.RecordConnection(true)
	if m.ActiveConnections != 1 {
		t.Error("Expected 1 active connection")
	}
	if m.TotalConnections != 1 {
		t.Error("Expected 1 total connection")
	}

	// Test failed connection
	m.RecordConnection(false)
	if m.ConnectionErrors != 1 {
		t.Error("Expected 1 connection error")
	}

	// Test closing connection
	m.RecordConnectionClosed()
	if m.ActiveConnections != 0 {
		t.Error("Expected 0 active connections")
	}
}

func TestMetrics_Batch(t *testing.T) {
	m := NewMetrics()

	// Record some batches
	m.RecordBatch(10, 0.9)
	m.RecordBatch(15, 0.8)

	if m.BatchesProcessed != 2 {
		t.Errorf("Expected 2 batches processed, got %d", m.BatchesProcessed)
	}

	expectedAvg := 12.5 // (10 + 15) / 2
	if m.AvgBatchSize != expectedAvg {
		t.Errorf("Expected avg batch size %.1f, got %.1f", expectedAvg, m.AvgBatchSize)
	}
}

func TestMetrics_TemplateCache(t *testing.T) {
	m := NewMetrics()

	// Test cache hit
	m.RecordTemplateCache(true, 5)
	if m.TemplateCacheHits != 1 {
		t.Error("Expected 1 cache hit")
	}
	if m.TemplateCacheSize != 5 {
		t.Error("Expected cache size 5")
	}

	// Test cache miss
	m.RecordTemplateCache(false, 6)
	if m.TemplateCacheMisses != 1 {
		t.Error("Expected 1 cache miss")
	}
	if m.TemplateCacheSize != 6 {
		t.Error("Expected cache size 6")
	}
}

func TestMetrics_HTTPEndpoint(t *testing.T) {
	m := NewMetrics()
	
	// Add some test data
	m.RecordEmailSent(200 * time.Millisecond)
	m.RecordEmailFailed(errors.New("smtp error"))
	m.RecordBatch(25, 0.95)

	// Test HTTP handler
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "emails_sent") {
		t.Error("Response should contain emails_sent")
	}
	if !strings.Contains(body, "emails_failed") {
		t.Error("Response should contain emails_failed")
	}
}

func TestMetrics_JSONOutput(t *testing.T) {
	m := NewMetrics()
	
	// Add some test data
	m.RecordEmailSent(150 * time.Millisecond)
	
	stats := m.GetStats()
	
	// Should be valid JSON
	if !strings.Contains(stats, "{") || !strings.Contains(stats, "}") {
		t.Error("Output should be valid JSON")
	}
	
	if !strings.Contains(stats, "emails_sent") {
		t.Error("JSON should contain emails_sent field")
	}
}

func TestMetrics_Throttle(t *testing.T) {
	m := NewMetrics()
	
	// Test throttle recording
	m.RecordThrottle(100)
	
	if m.ThrottleEvents != 1 {
		t.Error("Expected 1 throttle event")
	}
	
	if m.CurrentRateLimit != 100 {
		t.Error("Expected rate limit 100")
	}
}