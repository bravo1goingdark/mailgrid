package cli

import (
	"fmt"
	"mailgrid/config"
	"mailgrid/email"
	"mailgrid/parser"
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
	// Validate mutual exclusivity of CSV and SheetURL
	if args.SheetURL != "" && args.CSVPath != "example/fake_batch_1.csv" {
		return fmt.Errorf("must provide either CSV or SheetURL, not both")
	}

	// Parse Recipients
	var recipients []parser.Recipient

	if args.SheetURL != "" {
		// Fetch and stream public Google Sheet
		stream, err := parser.GetSheetCSVStream(args.SheetURL)
		if err != nil {
			return fmt.Errorf("failed to fetch Google Sheet: %w", err)
		}
		defer stream.Close()

		recipients, err = parser.ParseCSV(stream)
		if err != nil {
			return fmt.Errorf("failed to parse Google Sheet as CSV: %w", err)
		}

		// Log extracted ID and GID for transparency
		id, gid, _ := parser.ExtractSheetInfo(args.SheetURL)
		fmt.Printf("📄 Loaded Google Sheet: Spreadsheet ID = %s, GID = %s\n", id, gid)

	} else {
		// Default: Parse local CSV
		recipients, err = parser.ParseCSV(args.CSVPath)
		if err != nil {
			return fmt.Errorf("failed to parse CSV: %w", err)
		}
	}

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
	email.StartDispatcher(tasks, cfg.SMTP, args.Concurrency, args.BatchSize)

	fmt.Printf("\u2705 Completed in %s using %d workers\n", time.Since(start), args.Concurrency)
	return nil
}
