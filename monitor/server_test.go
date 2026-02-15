package monitor

import (
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer(9091)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.stats == nil {
		t.Fatal("Server stats is nil")
	}

	if server.clients == nil {
		t.Fatal("Server clients map is nil")
	}

	if server.server == nil {
		t.Fatal("HTTP server is nil")
	}

	// Check server address
	expectedAddr := ":9091"
	if server.server.Addr != expectedAddr {
		t.Errorf("Expected server address '%s', got '%s'", expectedAddr, server.server.Addr)
	}
}

func TestInitializeCampaign(t *testing.T) {
	server := NewServer(9091)

	jobID := "test-job-123"
	config := ConfigSummary{
		CSVFile:           "test.csv",
		TemplateFile:      "template.html",
		ConcurrentWorkers: 5,
		BatchSize:         10,
		RetryLimit:        3,
	}
	totalRecipients := 100

	server.InitializeCampaign(jobID, config, totalRecipients)

	if server.stats.JobID != jobID {
		t.Errorf("Expected JobID '%s', got '%s'", jobID, server.stats.JobID)
	}

	if server.stats.TotalRecipients != totalRecipients {
		t.Errorf("Expected TotalRecipients %d, got %d", totalRecipients, server.stats.TotalRecipients)
	}

	if server.stats.PendingCount != totalRecipients {
		t.Errorf("Expected PendingCount %d, got %d", totalRecipients, server.stats.PendingCount)
	}

	if server.stats.ConfigSummary.CSVFile != config.CSVFile {
		t.Errorf("Expected CSVFile '%s', got '%s'", config.CSVFile, server.stats.ConfigSummary.CSVFile)
	}

	if server.stats.ConfigSummary.ConcurrentWorkers != config.ConcurrentWorkers {
		t.Errorf("Expected ConcurrentWorkers %d, got %d", config.ConcurrentWorkers, server.stats.ConfigSummary.ConcurrentWorkers)
	}
}

func TestUpdateRecipientStatus(t *testing.T) {
	server := NewServer(9091)

	// Initialize campaign first
	server.InitializeCampaign("test-job", ConfigSummary{}, 10)

	email := "test@example.com"
	status := StatusSent
	duration := 500 * time.Millisecond
	errorMsg := ""

	server.UpdateRecipientStatus(email, status, duration, errorMsg)

	// Check if recipient was added
	recipient, exists := server.stats.Recipients[email]
	if !exists {
		t.Fatal("Recipient was not added to stats")
	}

	if recipient.Email != email {
		t.Errorf("Expected recipient email '%s', got '%s'", email, recipient.Email)
	}

	if recipient.Status != status {
		t.Errorf("Expected recipient status '%s', got '%s'", status, recipient.Status)
	}

	if recipient.Duration != duration.Nanoseconds()/1e6 {
		t.Errorf("Expected recipient duration %d ms, got %d ms", duration.Nanoseconds()/1e6, recipient.Duration)
	}

	// Check counts
	if server.stats.SentCount != 1 {
		t.Errorf("Expected SentCount 1, got %d", server.stats.SentCount)
	}

	// PendingCount stays at 10 because the recipient was new (not transitioning from pending)
	if server.stats.PendingCount != 10 {
		t.Errorf("Expected PendingCount 10, got %d", server.stats.PendingCount)
	}
}

func TestUpdateRecipientStatusMultiple(t *testing.T) {
	server := NewServer(9091)
	server.InitializeCampaign("test-job", ConfigSummary{}, 10)

	// Update multiple recipients with different statuses
	testCases := []struct {
		email  string
		status EmailStatus
	}{
		{"user1@example.com", StatusSent},
		{"user2@example.com", StatusSent},
		{"user3@example.com", StatusFailed},
		{"user4@example.com", StatusRetry},
	}

	for _, tc := range testCases {
		server.UpdateRecipientStatus(tc.email, tc.status, 100*time.Millisecond, "")
	}

	// Check final counts
	if server.stats.SentCount != 2 {
		t.Errorf("Expected SentCount 2, got %d", server.stats.SentCount)
	}

	if server.stats.FailedCount != 1 {
		t.Errorf("Expected FailedCount 1, got %d", server.stats.FailedCount)
	}

	if server.stats.RetryCount != 1 {
		t.Errorf("Expected RetryCount 1, got %d", server.stats.RetryCount)
	}

	// PendingCount stays at 10 because recipients were new (not transitioning from pending)
	if server.stats.PendingCount != 10 {
		t.Errorf("Expected PendingCount 10, got %d", server.stats.PendingCount)
	}

	// Check that all recipients were recorded
	if len(server.stats.Recipients) != 4 {
		t.Errorf("Expected 4 recipients, got %d", len(server.stats.Recipients))
	}
}

func TestDomainBreakdownSingleRecipientMultipleTransitions(t *testing.T) {
	server := NewServer(9091)
	server.InitializeCampaign("test-job", ConfigSummary{}, 10)

	email := "user@example.com"
	statuses := []EmailStatus{
		StatusPending,
		StatusSending,
		StatusFailed,
		StatusRetry,
		StatusSent,
	}

	for _, status := range statuses {
		server.UpdateRecipientStatus(email, status, 100*time.Millisecond, "")
	}

	domain := "example.com"
	if count := server.stats.DomainBreakdown[domain]; count != 1 {
		t.Fatalf("expected domain count for %s to remain 1, got %d", domain, count)
	}
}

func TestAddSMTPResponse(t *testing.T) {
	server := NewServer(9091)

	// Add some SMTP response codes
	server.AddSMTPResponse("250")
	server.AddSMTPResponse("250")
	server.AddSMTPResponse("550")
	server.AddSMTPResponse("250")

	// Check counts
	if server.stats.SMTPResponseCodes["250"] != 3 {
		t.Errorf("Expected 3 occurrences of '250', got %d", server.stats.SMTPResponseCodes["250"])
	}

	if server.stats.SMTPResponseCodes["550"] != 1 {
		t.Errorf("Expected 1 occurrence of '550', got %d", server.stats.SMTPResponseCodes["550"])
	}
}

func TestEmailStatusConstants(t *testing.T) {
	// Test that status constants are defined correctly
	expectedStatuses := map[EmailStatus]string{
		StatusPending: "pending",
		StatusSending: "sending",
		StatusSent:    "sent",
		StatusFailed:  "failed",
		StatusRetry:   "retry",
	}

	for status, expected := range expectedStatuses {
		if string(status) != expected {
			t.Errorf("Expected status '%s', got '%s'", expected, string(status))
		}
	}
}
