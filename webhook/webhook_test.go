package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.httpClient == nil {
		t.Fatal("HTTP client is nil")
	}

	// Check timeout is set
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Empty URL should be valid",
			url:     "",
			wantErr: false,
		},
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL",
			url:     "https://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "Invalid scheme",
			url:     "ftp://example.com/webhook",
			wantErr: true,
		},
		{
			name:    "Invalid URL format",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendNotificationSync(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		expectedContentType := "application/json"
		if r.Header.Get("Content-Type") != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, r.Header.Get("Content-Type"))
		}

		// Verify user agent
		expectedUserAgent := "Mailgrid-Webhook/1.0"
		if r.Header.Get("User-Agent") != expectedUserAgent {
			t.Errorf("Expected User-Agent %s, got %s", expectedUserAgent, r.Header.Get("User-Agent"))
		}

		// Parse the request body
		var result CampaignResult
		err := json.NewDecoder(r.Body).Decode(&result)
		if err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify some fields
		if result.JobID != "test-job-123" {
			t.Errorf("Expected JobID 'test-job-123', got '%s'", result.JobID)
		}

		if result.Status != "completed" {
			t.Errorf("Expected Status 'completed', got '%s'", result.Status)
		}

		// Send successful response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client and test data
	client := NewClient()
	result := CampaignResult{
		JobID:                "test-job-123",
		Status:               "completed",
		TotalRecipients:      100,
		SuccessfulDeliveries: 95,
		FailedDeliveries:     5,
		StartTime:            time.Now().Add(-1 * time.Hour),
		EndTime:              time.Now(),
		DurationSeconds:      3600,
	}

	// Send notification
	err := client.SendNotificationSync(server.URL, result)
	if err != nil {
		t.Fatalf("SendNotificationSync failed: %v", err)
	}
}

func TestSendNotificationSync_EmptyURL(t *testing.T) {
	client := NewClient()
	result := CampaignResult{
		JobID:  "test-job",
		Status: "completed",
	}

	// Should return nil for empty URL (no webhook configured)
	err := client.SendNotificationSync("", result)
	if err != nil {
		t.Errorf("Expected nil error for empty URL, got: %v", err)
	}
}

func TestSendNotificationSync_ServerError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	result := CampaignResult{
		JobID:  "test-job",
		Status: "completed",
	}

	// Should return error for server error
	err := client.SendNotificationSync(server.URL, result)
	if err == nil {
		t.Error("Expected error for server error response")
	}
}

func TestSendNotification(t *testing.T) {
	// Test that the async version doesn't return an error immediately
	client := NewClient()
	result := CampaignResult{
		JobID:  "test-job",
		Status: "completed",
	}

	// Should return nil immediately (async call)
	err := client.SendNotification("http://example.com/webhook", result)
	if err != nil {
		t.Errorf("SendNotification should return nil immediately, got: %v", err)
	}

	// For empty URL
	err = client.SendNotification("", result)
	if err != nil {
		t.Errorf("SendNotification should return nil for empty URL, got: %v", err)
	}
}
