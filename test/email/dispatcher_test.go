package email

import (
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/parser"
)

func TestNewPooledDispatcher(t *testing.T) {
	// Skip test if no SMTP server available
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	dispatcher, err := NewPooledDispatcher(cfg, 2, 10)
	if err != nil {
		t.Skipf("Skipping test - no SMTP server available: %v", err)
		return
	}
	defer dispatcher.Close()

	if dispatcher == nil {
		t.Fatal("Expected non-nil dispatcher")
	}

	if dispatcher.pool == nil {
		t.Error("Expected pool to be initialized")
	}

	if dispatcher.processor == nil {
		t.Error("Expected processor to be initialized")
	}

	if dispatcher.webhook == nil {
		t.Error("Expected webhook client to be initialized")
	}

	if dispatcher.monitor == nil {
		t.Error("Expected monitor to be initialized")
	}
}

func TestNewPooledDispatcherWithCampaign(t *testing.T) {
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	campaign := CampaignConfig{
		JobID:             "test-job-123",
		TotalRecipients:   100,
		CSVFile:           "test.csv",
		TemplateFile:      "template.html",
		ConcurrentWorkers: 5,
		WebhookURL:        "https://example.com/webhook",
		Monitor:           monitor.NewNoOpMonitor(),
	}

	dispatcher, err := NewPooledDispatcherWithCampaign(cfg, 2, 10, campaign)
	if err != nil {
		t.Skipf("Skipping test - no SMTP server available: %v", err)
		return
	}
	defer dispatcher.Close()

	// Check campaign metrics were set correctly
	metrics := dispatcher.GetMetrics()
	if metrics.jobID != "test-job-123" {
		t.Errorf("Expected jobID 'test-job-123', got %s", metrics.jobID)
	}

	if metrics.totalRecipients != 100 {
		t.Errorf("Expected totalRecipients 100, got %d", metrics.totalRecipients)
	}

	if metrics.csvFile != "test.csv" {
		t.Errorf("Expected csvFile 'test.csv', got %s", metrics.csvFile)
	}

	if metrics.templateFile != "template.html" {
		t.Errorf("Expected templateFile 'template.html', got %s", metrics.templateFile)
	}

	if metrics.webhookURL != "https://example.com/webhook" {
		t.Errorf("Expected webhookURL 'https://example.com/webhook', got %s", metrics.webhookURL)
	}
}

func TestCampaignMetrics_RecordSuccess(t *testing.T) {
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	dispatcher, err := NewPooledDispatcher(cfg, 1, 1)
	if err != nil {
		t.Skipf("Skipping test - no SMTP server available: %v", err)
		return
	}
	defer dispatcher.Close()

	// Record some successes
	dispatcher.RecordSuccess()
	dispatcher.RecordSuccess()
	dispatcher.RecordSuccess()

	metrics := dispatcher.GetMetrics()
	if metrics.successfulDeliveries != 3 {
		t.Errorf("Expected 3 successful deliveries, got %d", metrics.successfulDeliveries)
	}

	if metrics.failedDeliveries != 0 {
		t.Errorf("Expected 0 failed deliveries, got %d", metrics.failedDeliveries)
	}
}

func TestCampaignMetrics_RecordFailure(t *testing.T) {
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	dispatcher, err := NewPooledDispatcher(cfg, 1, 1)
	if err != nil {
		t.Skipf("Skipping test - no SMTP server available: %v", err)
		return
	}
	defer dispatcher.Close()

	// Record some failures
	dispatcher.RecordFailure()
	dispatcher.RecordFailure()

	metrics := dispatcher.GetMetrics()
	if metrics.failedDeliveries != 2 {
		t.Errorf("Expected 2 failed deliveries, got %d", metrics.failedDeliveries)
	}

	if metrics.successfulDeliveries != 0 {
		t.Errorf("Expected 0 successful deliveries, got %d", metrics.successfulDeliveries)
	}
}

func TestCampaignMetrics_RecordWithDetails(t *testing.T) {
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	dispatcher, err := NewPooledDispatcher(cfg, 1, 1)
	if err != nil {
		t.Skipf("Skipping test - no SMTP server available: %v", err)
		return
	}
	defer dispatcher.Close()

	// Record success with details
	dispatcher.RecordSuccessWithDetails("test@example.com", 100*time.Millisecond)

	// Record failure with details
	dispatcher.RecordFailureWithDetails("fail@example.com", 200*time.Millisecond, "SMTP error")

	// Record retry with details
	dispatcher.RecordRetryWithDetails("retry@example.com", 150*time.Millisecond, "Temporary failure")

	metrics := dispatcher.GetMetrics()
	if metrics.successfulDeliveries != 1 {
		t.Errorf("Expected 1 successful delivery, got %d", metrics.successfulDeliveries)
	}

	if metrics.failedDeliveries != 1 {
		t.Errorf("Expected 1 failed delivery, got %d", metrics.failedDeliveries)
	}
}

func TestTaskCreation(t *testing.T) {
	recipient := parser.Recipient{
		Email: "test@example.com",
		Data:  map[string]string{"name": "Test User"},
	}

	task := Task{
		Recipient:   recipient,
		Subject:     "Test Subject",
		Body:        "Test Body",
		Retries:     3,
		Attachments: []string{"file1.txt", "file2.pdf"},
		CC:          []string{"cc@example.com"},
		BCC:         []string{"bcc@example.com"},
	}

	if task.Recipient.Email != "test@example.com" {
		t.Errorf("Expected recipient email 'test@example.com', got %s", task.Recipient.Email)
	}

	if task.Subject != "Test Subject" {
		t.Errorf("Expected subject 'Test Subject', got %s", task.Subject)
	}

	if task.Body != "Test Body" {
		t.Errorf("Expected body 'Test Body', got %s", task.Body)
	}

	if task.Retries != 3 {
		t.Errorf("Expected retries 3, got %d", task.Retries)
	}

	if len(task.Attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(task.Attachments))
	}

	if len(task.CC) != 1 || task.CC[0] != "cc@example.com" {
		t.Errorf("Expected CC ['cc@example.com'], got %v", task.CC)
	}

	if len(task.BCC) != 1 || task.BCC[0] != "bcc@example.com" {
		t.Errorf("Expected BCC ['bcc@example.com'], got %v", task.BCC)
	}
}

func TestCampaignConfig(t *testing.T) {
	config := CampaignConfig{
		JobID:             "test-job-456",
		TotalRecipients:   250,
		CSVFile:           "contacts.csv",
		SheetURL:          "https://sheets.google.com/...",
		TemplateFile:      "email-template.html",
		ConcurrentWorkers: 10,
		WebhookURL:        "https://api.example.com/webhook",
		Monitor:           monitor.NewNoOpMonitor(),
	}

	if config.JobID != "test-job-456" {
		t.Errorf("Expected JobID 'test-job-456', got %s", config.JobID)
	}

	if config.TotalRecipients != 250 {
		t.Errorf("Expected TotalRecipients 250, got %d", config.TotalRecipients)
	}

	if config.CSVFile != "contacts.csv" {
		t.Errorf("Expected CSVFile 'contacts.csv', got %s", config.CSVFile)
	}

	if config.SheetURL != "https://sheets.google.com/..." {
		t.Errorf("Expected SheetURL 'https://sheets.google.com/...', got %s", config.SheetURL)
	}

	if config.ConcurrentWorkers != 10 {
		t.Errorf("Expected ConcurrentWorkers 10, got %d", config.ConcurrentWorkers)
	}

	if config.Monitor == nil {
		t.Error("Expected Monitor to be set")
	}
}

// Mock implementation for testing
type mockMetricsRecorder struct {
	successCount      int
	failureCount      int
	successWithDetail []string
	failureWithDetail []string
	retryWithDetail   []string
}

func (m *mockMetricsRecorder) RecordSuccess() {
	m.successCount++
}

func (m *mockMetricsRecorder) RecordFailure() {
	m.failureCount++
}

func (m *mockMetricsRecorder) RecordSuccessWithDetails(email string, duration time.Duration) {
	m.successWithDetail = append(m.successWithDetail, email)
}

func (m *mockMetricsRecorder) RecordFailureWithDetails(email string, duration time.Duration, errorMsg string) {
	m.failureWithDetail = append(m.failureWithDetail, email)
}

func (m *mockMetricsRecorder) RecordRetryWithDetails(email string, duration time.Duration, errorMsg string) {
	m.retryWithDetail = append(m.retryWithDetail, email)
}

func TestMetricsRecorderInterface(t *testing.T) {
	recorder := &mockMetricsRecorder{}

	// Test basic recording
	recorder.RecordSuccess()
	recorder.RecordFailure()

	if recorder.successCount != 1 {
		t.Errorf("Expected 1 success, got %d", recorder.successCount)
	}

	if recorder.failureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", recorder.failureCount)
	}

	// Test detailed recording
	recorder.RecordSuccessWithDetails("success@example.com", time.Millisecond)
	recorder.RecordFailureWithDetails("failure@example.com", time.Millisecond, "error")
	recorder.RecordRetryWithDetails("retry@example.com", time.Millisecond, "retry error")

	if len(recorder.successWithDetail) != 1 || recorder.successWithDetail[0] != "success@example.com" {
		t.Errorf("Expected success detail ['success@example.com'], got %v", recorder.successWithDetail)
	}

	if len(recorder.failureWithDetail) != 1 || recorder.failureWithDetail[0] != "failure@example.com" {
		t.Errorf("Expected failure detail ['failure@example.com'], got %v", recorder.failureWithDetail)
	}

	if len(recorder.retryWithDetail) != 1 || recorder.retryWithDetail[0] != "retry@example.com" {
		t.Errorf("Expected retry detail ['retry@example.com'], got %v", recorder.retryWithDetail)
	}
}