package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
	"github.com/bravo1goingdark/mailgrid/scheduler"
	"github.com/bravo1goingdark/mailgrid/utils"
	"github.com/bravo1goingdark/mailgrid/utils/preview"
	"github.com/bravo1goingdark/mailgrid/utils/valid"
)

const maxAttachSize = 10 << 20 // 10 MB

// Run orchestrates the CLI workflow. Depending on flags, it will:
//   - Schedule a job (--at, --cron, --every)
//   - Cancel a scheduled job (--cancel)
//   - List scheduled jobs (--list)
//   - Or run an immediate email sent
func Run(args CLIArgs) error {
	// Initialize scheduler (persistent JSON-backed store)
	sched := scheduler.NewScheduler("scheduler/jobs.json")

	// Attach job execution handler
	handler := func(j scheduler.Job) {
		fmt.Printf("🔔 Running scheduled job %s (created %s)\n", j.ID, j.CreatedAt.Format(time.RFC3339))
		if err := runOnce(j.Args); err != nil {
			fmt.Printf("❌ Job %s failed: %v\n", j.ID, err)
		}
	}

	// Restore jobs after restart
	sched.ReattachHandlers(handler)

	// Cancel a job if requested
	if args.CancelJobID != "" {
		if sched.CancelJob(args.CancelJobID) {
			fmt.Printf("✅ Cancelled job %s\n", args.CancelJobID)
		} else {
			fmt.Printf("❌ No job found with ID %s\n", args.CancelJobID)
		}
		return nil
	}

	// List jobs
	if args.ListJobs {
		jobs := sched.ListJobs()
		if len(jobs) == 0 {
			fmt.Println("No scheduled jobs.")
			return nil
		}
		for _, j := range jobs {
			fmt.Printf("- ID: %s | Status: %s | Created: %s\n", j.ID, j.Status, j.CreatedAt.Format(time.RFC3339))
			if !j.RunAt.IsZero() {
				fmt.Printf("    RunAt: %s\n", j.RunAt.Format(time.RFC3339))
			}
			if j.CronExpr != "" {
				fmt.Printf("    Cron: %s\n", j.CronExpr)
			}
			if j.Interval != "" {
				fmt.Printf("    Every: %s\n", j.Interval)
			}
		}
		return nil
	}

	// Schedule job if requested
	if args.ScheduleAt != "" || args.CronExpr != "" || args.Interval != "" {
		var runAt time.Time
		if args.ScheduleAt != "" {
			parsed, err := time.Parse(time.RFC3339, args.ScheduleAt)
			if err != nil {
				return fmt.Errorf("invalid --at time: %w", err)
			}
			runAt = parsed
		}
		job := scheduler.NewJob(args, runAt, args.CronExpr, args.Interval)
		if err := sched.AddJob(job, handler); err != nil {
			return fmt.Errorf("failed to schedule job: %w", err)
		}
		fmt.Printf("✅ Scheduled job %s\n", job.ID)
		return nil
	}

	// Default: run immediately
	return runOnce(args)
}

// runOnce executes the full Mailgrid pipeline once (used by both immediate and scheduled runs).
func runOnce(args CLIArgs) error {
	// Load SMTP config
	cfg, err := config.LoadConfig(args.EnvPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Mutually exclusive: --to vs --csv/--sheet-url
	if args.To != "" {
		if args.CSVPath != "" || args.SheetURL != "" {
			return fmt.Errorf("❌ --to is mutually exclusive with --csv and --sheet-url")
		}
		return SendSingleEmail(args, cfg.SMTP)
	}
	if args.CSVPath == "" && args.SheetURL == "" {
		return fmt.Errorf("❌ You must provide either --csv or --sheet-url")
	}
	if args.CSVPath != "" && args.SheetURL != "" {
		return fmt.Errorf("❌ Provide only one of --csv or --sheet-url, not both")
	}

	// Validate attachments
	for _, f := range args.Attachments {
		info, err := os.Stat(f)
		if err != nil {
			return fmt.Errorf("attachment not found: %s", f)
		}
		if info.Size() > maxAttachSize {
			return fmt.Errorf("attachment too large (>%d bytes): %s", maxAttachSize, f)
		}
	}
	if args.TemplatePath == "" && len(args.Attachments) == 0 {
		return fmt.Errorf("provide --template, --attach, or both")
	}

	// Parse CC/BCC
	ccList, err := valid.ParseAddressInput(args.Cc)
	if err != nil {
		return fmt.Errorf("failed to parse CC: %w", err)
	}
	bccList, err := valid.ParseAddressInput(args.Bcc)
	if err != nil {
		return fmt.Errorf("failed to parse BCC: %w", err)
	}

	// Load recipients
	var recipients []parser.Recipient
	if args.SheetURL != "" {
		stream, err := parser.GetSheetCSVStream(args.SheetURL)
		if err != nil {
			return fmt.Errorf("failed to fetch Google Sheet: %w", err)
		}
		defer func(stream io.ReadCloser) {
			err := stream.Close()
			if err != nil {
				return
			}
		}(stream)

		recipients, err = parser.ParseCSVFromReader(stream)
		if err != nil {
			return fmt.Errorf("failed to parse Google Sheet: %w", err)
		}
		id, gid, _ := utils.ExtractSheetInfo(args.SheetURL)
		fmt.Printf("📄 Loaded Google Sheet: Spreadsheet ID=%s, GID=%s\n", id, gid)
	} else {
		recipients, err = parser.ParseCSV(args.CSVPath)
		if err != nil {
			return fmt.Errorf("failed to parse CSV: %w", err)
		}
	}

	// Apply optional filter
	if args.Filter != "" {
		expr, err := expression.Parse(args.Filter)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
		if err := valid.ValidateFields(expr, recipients); err != nil {
			return fmt.Errorf("invalid filter field: %w", err)
		}
		recipients = parser.Filter(recipients, expr)
		if len(recipients) == 0 {
			return fmt.Errorf("no recipients matched filter: %q", args.Filter)
		}
	}

	// Preview mode
	if args.ShowPreview {
		if args.TemplatePath == "" {
			return fmt.Errorf("cannot preview without --template")
		}
		if len(recipients) == 0 {
			return fmt.Errorf("no recipients found in CSV for preview")
		}
		rendered, err := preview.RenderTemplate(recipients[0], args.TemplatePath)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		return preview.StartServer(rendered, args.PreviewPort)
	}

	// Build tasks
	tasks, err := PrepareEmailTasks(recipients, args.TemplatePath, args.Subject, args.Attachments, ccList, bccList)
	if err != nil {
		return err
	}

	// Dry-run mode
	if args.DryRun {
		printDryRun(tasks)
		return nil
	}

	// Dispatch emails
	start := time.Now()
	email.SetRetryLimit(args.RetryLimit)
	email.StartDispatcher(tasks, cfg.SMTP, args.Concurrency, args.BatchSize)
	fmt.Printf("✅ Completed in %s using %d workers\n", time.Since(start), args.Concurrency)
	return nil
}
