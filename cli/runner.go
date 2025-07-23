package cli

import (
	"fmt"
	"mailgrid/config"
	"mailgrid/email"
	"mailgrid/parser"
	"mailgrid/utils"
	"mailgrid/utils/preview"
	"time"
)

// Run is the main orchestration function. It controls the full Mailgrid lifecycle:
// 1. Load config
// 2. Parse CSV
// 3. Render and preview/send emails
func Run(args CLIArgs) error {
	// Load SMTP configuration from file
	cfg, err := config.LoadConfig(args.EnvPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse CSV recipients into []Recipient
	recipients, err := parser.ParseCSV(args.CSVPath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}
	// Handle offset logic
	if args.ResetOffset { // <-- If reset is requested
		utils.ResetOffset() // <-- Delete any saved offset
	}
	startFrom := 0
	if args.Resume { // <-- If resume is requested
		startFrom = utils.LoadOffset() // <-- Load saved offset
	}
	// Edge case: make sure offset doesn't exceed available recipients
	if startFrom > len(recipients) {
		return fmt.Errorf("resume offset exceeds number of recipients")
	}
	recipients = recipients[startFrom:] // <-- Trim recipient list to offset

	// If preview mode is enabled, serve one rendered email via localhost
	if args.ShowPreview {
		if len(recipients) == 0 {
			return fmt.Errorf("no recipients found in CSV for preview")
		}
		rendered, err := preview.RenderTemplate(recipients[0], args.TemplatePath)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		return preview.StartServer(rendered, args.PreviewPort)
	}

	// Render subject & body for each recipient and build email.Task list
	tasks, err := prepareEmailTasks(recipients, args.TemplatePath, args.Subject)
	if err != nil {
		return err
	}

	// If dry-run mode, print emails and skip sending
	if args.DryRun {
		printDryRun(tasks)
		return nil
	}

	// Otherwise, send emails using dispatcher
	start := time.Now()
	email.SetRetryLimit(args.RetryLimit)

	// Send emails and track how many succeeded
	setCount := email.StartDispatcher(tasks, cfg.SMTP, args.Concurrency, args.BatchSize, startFrom)
	utils.SaveOffset(startFrom + setCount) // Save new offset after sending

	fmt.Printf("\u2705 Completed in %s using %d workers\n", time.Since(start), args.Concurrency)
	return nil
}
