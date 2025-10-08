package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL is valid",
			url:     "",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "valid HTTPS URL",
			url:     "https://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "invalid protocol",
			url:     "ftp://example.com/webhook",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
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

func TestClient_SendNotificationSync(t *testing.T) {
	// Create test server
	var receivedPayload CampaignResult
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		userAgent := r.Header.Get("User-Agent")
		if userAgent != "Mailgrid-Webhook/1.0" {
			t.Errorf("Expected User-Agent Mailgrid-Webhook/1.0, got %s", userAgent)
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode JSON: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()

	testTime := time.Now()
	result := CampaignResult{
		JobID:                "test-job-123",
		Status:               "completed",
		TotalRecipients:      100,
		SuccessfulDeliveries: 95,
		FailedDeliveries:     5,
		StartTime:            testTime,
		EndTime:              testTime.Add(5 * time.Minute),
		DurationSeconds:      300,
		CSVFile:              "test.csv",
		TemplateFile:         "template.html",
		ConcurrentWorkers:    5,
	}

	err := client.SendNotificationSync(server.URL, result)
	if err != nil {
		t.Errorf("SendNotificationSync() error = %v", err)
	}

	// Verify received payload
	if receivedPayload.JobID != result.JobID {
		t.Errorf("Expected JobID %s, got %s", result.JobID, receivedPayload.JobID)
	}
	if receivedPayload.Status != result.Status {
		t.Errorf("Expected Status %s, got %s", result.Status, receivedPayload.Status)
	}
	if receivedPayload.TotalRecipients != result.TotalRecipients {
		t.Errorf("Expected TotalRecipients %d, got %d", result.TotalRecipients, receivedPayload.TotalRecipients)
	}
}

func TestClient_SendNotificationSyncServerError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()

	result := CampaignResult{
		JobID:  "test-job-123",
		Status: "completed",
	}

	err := client.SendNotificationSync(server.URL, result)
	if err == nil {
		t.Errorf("Expected error for server error response, got nil")
	}
}

func TestClient_SendNotificationEmptyURL(t *testing.T) {
	client := NewClient()

	result := CampaignResult{
		JobID:  "test-job-123",
		Status: "completed",
	}

	err := client.SendNotificationSync("", result)
	if err != nil {
		t.Errorf("Expected no error for empty URL, got %v", err)
	}
}