package cli

import (
	"fmt"
	"mailgrid/config"
	"mailgrid/email"
	"mailgrid/parser"
	"mailgrid/parser/expression"
	"mailgrid/utils"
	"mailgrid/utils/preview"
	"mailgrid/utils/valid"
	"time"
)

// Run is the main orchestration function. It controls the full Mailgrid lifecycle:
// 1. Load config
// 2. Parse CSV or Google Sheet
// 3. Apply optional filter
// 4. Preview or send emails
func Run(args CLIArgs) error {
	// Load SMTP configuration from a file
	cfg, err := config.LoadConfig(args.EnvPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if args.CSVPath == "" && args.SheetURL == "" {
		return fmt.Errorf("❌ You must provide either --csv or --sheet-url")
	}
	if args.CSVPath != "" && args.SheetURL != "" {
		return fmt.Errorf("❌ Provide only one of --csv or --sheet-url, not both")
	}

	// Parse Recipients
	var recipients []parser.Recipient

	if args.SheetURL != "" {
		stream, err := parser.GetSheetCSVStream(args.SheetURL)
		if err != nil {
			return fmt.Errorf("failed to fetch Google Sheet: %w", err)
		}
		defer stream.Close()

		recipients, err = parser.ParseCSVFromReader(stream)
		if err != nil {
			return fmt.Errorf("failed to parse Google Sheet as CSV: %w", err)
		}

		id, gid, _ := utils.ExtractSheetInfo(args.SheetURL)
		fmt.Printf("📄 Loaded Google Sheet: Spreadsheet ID = %s, GID = %s\n", id, gid)

	} else {
		recipients, err = parser.ParseCSV(args.CSVPath)
		if err != nil {
			return fmt.Errorf("failed to parse CSV: %w", err)
		}
	}

	// Optional logical filtering
	if args.Filter != "" {
		if len(recipients) == 0 {
			return fmt.Errorf("no recipients found in CSV for filtering")
		}

		expr, err := expression.Parse(args.Filter)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}

		if err := valid.ValidateFields(expr, recipients); err != nil {
			return fmt.Errorf("invalid filter field: %w", err)
		}

		recipients = parser.Filter(recipients, expr)

		if len(recipients) == 0 {
			return fmt.Errorf("no recipients matched the filter: %q", args.Filter)
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
	tasks, err := PrepareEmailTasks(recipients, args.TemplatePath, args.Subject)
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
