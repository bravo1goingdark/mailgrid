package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/cli"
	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/internal/metrics"
	"github.com/bravo1goingdark/mailgrid/parser"
)

func TestFullWorkflow(t *testing.T) {
	// Create test CSV data
	csvContent := `email,name,company
test1@example.com,John Doe,Acme Corp
test2@example.com,Jane Smith,Tech Inc
invalid-email,Bad Entry,No Company`

	// Create temporary CSV file
	csvFile, err := os.CreateTemp("", "test-recipients-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(csvFile.Name())

	if _, err := csvFile.WriteString(csvContent); err != nil {
		t.Fatal(err)
	}
	csvFile.Close()

	// Test CSV parsing
	recipients, err := parser.ParseCSV(csvFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	if len(recipients) != 3 {
		t.Errorf("Expected 3 recipients, got %d", len(recipients))
	}

	// Check first recipient
	if recipients[0].Email != "test1@example.com" {
		t.Errorf("Expected first email to be test1@example.com, got %s", recipients[0].Email)
	}
	if recipients[0].Data["name"] != "John Doe" {
		t.Errorf("Expected first name to be John Doe, got %s", recipients[0].Data["name"])
	}
}

func TestConfigLoadingWithValidation(t *testing.T) {
	configContent := `{
		"smtp": {
			"host": "smtp.example.com",
			"port": 587,
			"username": "test@example.com",
			"password": "password123",
			"from": "test@example.com",
			"use_tls": true
		},
		"rate_limit": 50,
		"burst_limit": 100,
		"max_concurrency": 10,
		"max_batch_size": 50,
		"log": {
			"level": "info",
			"format": "json"
		},
		"metrics": {
			"enabled": true,
			"port": 8091
		}
	}`

	// Create temporary config file
	configFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())

	if _, err := configFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	configFile.Close()

	// Load and validate config
	cfg, err := config.LoadConfig(configFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.SMTP.Host != "smtp.example.com" {
		t.Errorf("Expected SMTP host to be smtp.example.com, got %s", cfg.SMTP.Host)
	}
	if cfg.RateLimit != 50 {
		t.Errorf("Expected rate limit to be 50, got %d", cfg.RateLimit)
	}
	if cfg.MaxConcurrency != 10 {
		t.Errorf("Expected max concurrency to be 10, got %d", cfg.MaxConcurrency)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected log level to be info, got %s", cfg.Log.Level)
	}
	if !cfg.Metrics.Enabled {
		t.Error("Expected metrics to be enabled")
	}
}

func TestEmailTaskPreparation(t *testing.T) {
	// Create test recipients
	recipients := []parser.Recipient{
		{
			Email: "test@example.com",
			Data:  map[string]string{"name": "John", "company": "Acme"},
		},
		{
			Email: "test2@example.com", 
			Data:  map[string]string{"name": "Jane", "company": "Tech"},
		},
	}

	// Create test template
	templateContent := `<html>
<body>
	<h1>Hello {{.name}}!</h1>
	<p>Welcome to {{.company}}!</p>
</body>
</html>`

	templateFile, err := os.CreateTemp("", "test-template-*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(templateFile.Name())

	if _, err := templateFile.WriteString(templateContent); err != nil {
		t.Fatal(err)
	}
	templateFile.Close()

	// Prepare email tasks
	tasks, err := cli.PrepareEmailTasks(recipients, templateFile.Name(), "Welcome {{.name}}!", []string{}, []string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to prepare email tasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Check first task
	task := tasks[0]
	if task.Recipient.Email != "test@example.com" {
		t.Errorf("Expected first task email to be test@example.com, got %s", task.Recipient.Email)
	}
	if task.Subject != "Welcome John!" {
		t.Errorf("Expected subject to be 'Welcome John!', got %s", task.Subject)
	}
	// Check if body rendering worked (basic check)
	if task.Body == "" {
		t.Error("Expected body to be rendered")
	} else {
		t.Logf("Template rendered successfully. Body length: %d", len(task.Body))
	}
}

func TestMetricsIntegration(t *testing.T) {
	m := metrics.GetMetrics()

	// Test metrics recording
	initialSent := m.EmailsSent.Value()
	initialFailed := m.EmailsFailed.Value()
	initialWorkers := m.ActiveWorkers.Value()

	m.RecordEmailSent()
	m.RecordEmailFailed()
	m.RecordWorkerStart()

	if m.EmailsSent.Value() != initialSent+1 {
		t.Errorf("Expected emails sent to increase by 1, got %d -> %d", initialSent, m.EmailsSent.Value())
	}
	if m.EmailsFailed.Value() != initialFailed+1 {
		t.Errorf("Expected emails failed to increase by 1, got %d -> %d", initialFailed, m.EmailsFailed.Value())
	}
	if m.ActiveWorkers.Value() != initialWorkers+1 {
		t.Errorf("Expected active workers to increase by 1, got %d -> %d", initialWorkers, m.ActiveWorkers.Value())
	}

	// Test metrics server functionality
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		err := m.StartMetricsServer(ctx, 0) // Use port 0 for auto-assignment
		if err != nil && err.Error() != "http: Server closed" {
			t.Logf("Metrics server error (expected): %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
}

// TestDispatcherWithMockSMTP is skipped as it requires actual SMTP server for full testing
// The dispatcher functionality is tested in unit tests

func TestCLIArgsParser(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test arguments
	os.Args = []string{
		"mailgrid",
		"--env", "config.json",
		"--csv", "recipients.csv", 
		"--template", "template.html",
		"--subject", "Test Subject",
		"--concurrency", "5",
		"--batch-size", "20",
		"--dry-run",
	}

	args := cli.ParseFlags()

	// Verify parsed arguments
	if args.EnvPath != "config.json" {
		t.Errorf("Expected env path to be config.json, got %s", args.EnvPath)
	}
	if args.CSVPath != "recipients.csv" {
		t.Errorf("Expected CSV path to be recipients.csv, got %s", args.CSVPath)
	}
	if args.TemplatePath != "template.html" {
		t.Errorf("Expected template path to be template.html, got %s", args.TemplatePath)
	}
	if args.Subject != "Test Subject" {
		t.Errorf("Expected subject to be 'Test Subject', got %s", args.Subject)
	}
	if args.Concurrency != 5 {
		t.Errorf("Expected concurrency to be 5, got %d", args.Concurrency)
	}
	if args.BatchSize != 20 {
		t.Errorf("Expected batch size to be 20, got %d", args.BatchSize)
	}
	if !args.DryRun {
		t.Error("Expected dry-run to be true")
	}
}
