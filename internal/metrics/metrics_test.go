package metrics

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestMetricsSingleton(t *testing.T) {
	// Reset singleton for test
	once = sync.Once{}
	instance = nil

	m1 := GetMetrics()
	m2 := GetMetrics()

	if m1 != m2 {
		t.Error("GetMetrics should return the same instance")
	}
}

func TestMetricsIncrement(t *testing.T) {
	// Use existing singleton instead of resetting
	// (expvar doesn't allow resetting published variables)

	m := GetMetrics()

	// Test email metrics
	initial := m.EmailsSent.Value()
	m.RecordEmailSent()
	m.RecordEmailSent()
	if m.EmailsSent.Value() != initial+2 {
		t.Errorf("Expected emails sent to be %d, got %d", initial+2, m.EmailsSent.Value())
	}

	// Test failed emails
	initialFailed := m.EmailsFailed.Value()
	m.RecordEmailFailed()
	if m.EmailsFailed.Value() != initialFailed+1 {
		t.Errorf("Expected emails failed to be %d, got %d", initialFailed+1, m.EmailsFailed.Value())
	}

	// Test retries
	initialRetries := m.EmailsRetried.Value()
	m.RecordEmailRetried()
	if m.EmailsRetried.Value() != initialRetries+1 {
		t.Errorf("Expected emails retried to be %d, got %d", initialRetries+1, m.EmailsRetried.Value())
	}
}

func TestWorkerMetrics(t *testing.T) {
	// Use existing singleton

	m := GetMetrics()

	initial := m.ActiveWorkers.Value()
	
	m.RecordWorkerStart()
	m.RecordWorkerStart()
	if m.ActiveWorkers.Value() != initial+2 {
		t.Errorf("Expected active workers to be %d, got %d", initial+2, m.ActiveWorkers.Value())
	}

	m.RecordWorkerStop()
	if m.ActiveWorkers.Value() != initial+1 {
		t.Errorf("Expected active workers to be %d, got %d", initial+1, m.ActiveWorkers.Value())
	}
}

func TestSMTPConnectionMetrics(t *testing.T) {
	m := GetMetrics()

	initial := m.SMTPConnections.Value()
	
	m.RecordSMTPConnection()
	if m.SMTPConnections.Value() != initial+1 {
		t.Errorf("Expected SMTP connections to be %d, got %d", initial+1, m.SMTPConnections.Value())
	}

	m.RecordSMTPDisconnection()
	if m.SMTPConnections.Value() != initial {
		t.Errorf("Expected SMTP connections to be %d, got %d", initial, m.SMTPConnections.Value())
	}
}

func TestJobMetrics(t *testing.T) {
	m := GetMetrics()

	initialScheduled := m.JobsScheduled.Value()
	initialCompleted := m.JobsCompleted.Value()
	initialFailed := m.JobsFailed.Value()

	m.RecordJobScheduled()
	if m.JobsScheduled.Value() != initialScheduled+1 {
		t.Errorf("Expected jobs scheduled to be %d, got %d", initialScheduled+1, m.JobsScheduled.Value())
	}

	m.RecordJobCompleted()
	if m.JobsCompleted.Value() != initialCompleted+1 {
		t.Errorf("Expected jobs completed to be %d, got %d", initialCompleted+1, m.JobsCompleted.Value())
	}

	m.RecordJobFailed()
	if m.JobsFailed.Value() != initialFailed+1 {
		t.Errorf("Expected jobs failed to be %d, got %d", initialFailed+1, m.JobsFailed.Value())
	}
}

func TestResponseTimeMetrics(t *testing.T) {
	m := GetMetrics()

	duration := 100 * time.Millisecond
	m.RecordResponseTime("test_operation", duration)

	// Check if the response time was recorded (expvar.Map doesn't expose values easily for testing)
	// This test ensures the method doesn't panic
}

func TestErrorMetrics(t *testing.T) {
	m := GetMetrics()

	m.RecordError("smtp_error")
	m.RecordError("timeout_error")
	m.RecordError("smtp_error") // Same error type again

	// Check if errors were recorded (expvar.Map doesn't expose values easily for testing)
	// This test ensures the method doesn't panic
}

func TestMetricsServer(t *testing.T) {
	m := GetMetrics()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Start server on a different port to avoid conflicts
		err := m.StartMetricsServer(ctx, 0) // Port 0 lets OS choose
		if err != nil && err != http.ErrServerClosed {
			t.Logf("Metrics server error (expected): %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop server
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestHealthHandler(t *testing.T) {
	m := GetMetrics()
	
	// Test health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := &testResponseWriter{}
	m.healthHandler(rr, req)

	if rr.statusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.statusCode)
	}

	if rr.header.Get("Content-Type") != "application/json" {
		t.Error("Expected JSON content type")
	}

	expectedSubstring := `"status":"healthy"`
	if !contains(string(rr.body), expectedSubstring) {
		t.Errorf("Expected response to contain %q, got %q", expectedSubstring, string(rr.body))
	}
}

func TestReadinessHandler(t *testing.T) {
	m := GetMetrics()

	req, err := http.NewRequest("GET", "/ready", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Get current worker count (might be > 0 from other tests)
	currentWorkers := m.ActiveWorkers.Value()

	// Test with current workers
	rr := &testResponseWriter{}
	m.readinessHandler(rr, req)

	if currentWorkers > 0 {
		if rr.statusCode != http.StatusOK {
			t.Errorf("Expected status %d with %d workers, got %d", http.StatusOK, currentWorkers, rr.statusCode)
		}
	} else {
		if rr.statusCode != http.StatusServiceUnavailable {
			t.Errorf("Expected status %d with 0 workers, got %d", http.StatusServiceUnavailable, rr.statusCode)
		}
	}

	// Test with active workers
	m.RecordWorkerStart()
	rr2 := &testResponseWriter{}
	m.readinessHandler(rr2, req)

	if rr2.statusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr2.statusCode)
	}

	expectedSubstring := `"status":"ready"`
	if !contains(string(rr2.body), expectedSubstring) {
		t.Errorf("Expected response to contain %q, got %q", expectedSubstring, string(rr2.body))
	}
}

// Helper types for testing HTTP handlers
type testResponseWriter struct {
	header     http.Header
	body       []byte
	statusCode int
}

func (rw *testResponseWriter) Header() http.Header {
	if rw.header == nil {
		rw.header = make(http.Header)
	}
	return rw.header
}

func (rw *testResponseWriter) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return len(data), nil
}

func (rw *testResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
