package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer_StatusAPI(t *testing.T) {
	server := NewServer(9091)

	// Initialize campaign
	config := ConfigSummary{
		CSVFile:           "test.csv",
		ConcurrentWorkers: 2,
		BatchSize:         10,
	}
	server.InitializeCampaign("test-job", config, 100)

	// Update some recipient statuses
	server.UpdateRecipientStatus("test1@example.com", StatusSent, 100*time.Millisecond, "")
	server.UpdateRecipientStatus("test2@example.com", StatusFailed, 200*time.Millisecond, "SMTP error")

	// Create test request
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	server.handleStatusAPI(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Parse response
	var stats CampaignStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify stats
	if stats.JobID != "test-job" {
		t.Errorf("Expected JobID 'test-job', got %s", stats.JobID)
	}
	if stats.TotalRecipients != 100 {
		t.Errorf("Expected TotalRecipients 100, got %d", stats.TotalRecipients)
	}
	if stats.SentCount != 1 {
		t.Errorf("Expected SentCount 1, got %d", stats.SentCount)
	}
	if stats.FailedCount != 1 {
		t.Errorf("Expected FailedCount 1, got %d", stats.FailedCount)
	}

	// Check recipients
	if len(stats.Recipients) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(stats.Recipients))
	}

	recipient1, exists := stats.Recipients["test1@example.com"]
	if !exists {
		t.Error("Expected test1@example.com to exist in recipients")
	} else if recipient1.Status != StatusSent {
		t.Errorf("Expected test1@example.com status to be sent, got %s", recipient1.Status)
	}

	recipient2, exists := stats.Recipients["test2@example.com"]
	if !exists {
		t.Error("Expected test2@example.com to exist in recipients")
	} else if recipient2.Status != StatusFailed {
		t.Errorf("Expected test2@example.com status to be failed, got %s", recipient2.Status)
	}
}

func TestServer_UpdateRecipientStatus(t *testing.T) {
	server := NewServer(9091)

	// Initialize campaign
	config := ConfigSummary{CSVFile: "test.csv"}
	server.InitializeCampaign("test-job", config, 1)

	// Test updating recipient status
	email := "test@example.com"
	duration := 150 * time.Millisecond
	errorMsg := "test error"

	server.UpdateRecipientStatus(email, StatusFailed, duration, errorMsg)

	// Check the recipient was updated
	server.mu.RLock()
	recipient, exists := server.stats.Recipients[email]
	server.mu.RUnlock()

	if !exists {
		t.Fatalf("Expected recipient %s to exist", email)
	}

	if recipient.Status != StatusFailed {
		t.Errorf("Expected status failed, got %s", recipient.Status)
	}

	if recipient.Duration != duration.Nanoseconds()/1e6 {
		t.Errorf("Expected duration %d ms, got %d ms", duration.Nanoseconds()/1e6, recipient.Duration)
	}

	if recipient.Error != errorMsg {
		t.Errorf("Expected error '%s', got '%s'", errorMsg, recipient.Error)
	}

	// Check counts were updated
	if server.stats.FailedCount != 1 {
		t.Errorf("Expected FailedCount 1, got %d", server.stats.FailedCount)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"test@example.com", "example.com"},
		{"user@gmail.com", "gmail.com"},
		{"admin@subdomain.example.org", "subdomain.example.org"},
		{"invalid-email", ""},
		{"@domain.com", "domain.com"},
		{"user@", ""},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := extractDomain(tt.email)
			if result != tt.expected {
				t.Errorf("extractDomain(%s) = %s, expected %s", tt.email, result, tt.expected)
			}
		})
	}
}

func TestNoOpMonitor(t *testing.T) {
	monitor := NewNoOpMonitor()

	// These should not panic
	monitor.InitializeCampaign("test", ConfigSummary{}, 100)
	monitor.UpdateRecipientStatus("test@example.com", StatusSent, time.Millisecond, "")
	monitor.AddSMTPResponse("250")
	monitor.AddLogEntry("INFO", "test message", "test@example.com")
}