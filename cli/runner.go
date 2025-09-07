package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
	"github.com/bravo1goingdark/mailgrid/scheduler"
	"github.com/bravo1goingdark/mailgrid/utils"
	"github.com/bravo1goingdark/mailgrid/utils/preview"
	"github.com/bravo1goingdark/mailgrid/utils/valid"
)

const maxAttachSize = 10 << 20 // 10 MB

// Run is the main orchestration function. It controls the full Mailgrid lifecycle:
// 1. Load config
// 2. Parse CSV or Google Sheet
// 3. Apply optional filter
// 4. Preview or send emails
func Run(args CLIArgs) error {
	// Run scheduler dispatcher in foreground
	if args.SchedulerRun {
		db, err := database.NewDB(args.SchedulerDB)
		if err != nil {
			return fmt.Errorf("open scheduler db: %w", err)
		}
		log := logger.New("cli-scheduler-run")
		s := scheduler.NewScheduler(db, log)
		es := scheduler.NewEmailScheduler(s)
		r := NewRunner(es)
		es.ReattachHandlers(r.EmailJobHandler)
		fmt.Printf("⏱️  Scheduler running with DB %s. Press Ctrl+C to stop.\n", args.SchedulerDB)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		<-ctx.Done()
		fmt.Println("\nShutting down scheduler...")
		s.Stop()
		_ = db.Close()
		fmt.Println("Scheduler stopped.")
		return nil
	}

	// Jobs admin: list/cancel
	if args.ListJobs || args.CancelJobID != "" {
		db, err := database.NewDB(args.SchedulerDB)
		if err != nil {
			return fmt.Errorf("open scheduler db: %w", err)
		}
		log := logger.New("cli-jobs")
		s := scheduler.NewScheduler(db, log)
		defer func() {
			s.Stop()
			_ = db.Close()
		}()

		if args.CancelJobID != "" {
			ok := s.CancelJob(args.CancelJobID)
			if !ok {
				return fmt.Errorf("no such job: %s", args.CancelJobID)
			}
			fmt.Printf("🛑 Cancelled job %s\n", args.CancelJobID)
		}
		if args.ListJobs {
			jobs, err := s.ListJobs()
			if err != nil {
				return err
			}
			if len(jobs) == 0 {
				fmt.Println("(no jobs)")
				return nil
			}
			for _, j := range jobs {
				next := "-"
				if !j.NextRunAt.IsZero() {
					next = j.NextRunAt.Format(time.RFC3339)
				}
				fmt.Printf("%s\t%s\trunAt=%s\tnext=%s\tattempts=%d/%d\n", j.ID, j.Status, j.RunAt.Format(time.RFC3339), next, j.Attempts, j.MaxAttempts)
			}
		}
		return nil
	}

	// If scheduling flags are set, schedule a job and return
	if args.ScheduleAt != "" || args.Interval != "" || args.Cron != "" {
		if args.To == "" && args.CSVPath == "" && args.SheetURL == "" {
			return fmt.Errorf("❌ scheduling requires --to or --csv or --sheet-url")
		}
		db, err := database.NewDB(args.SchedulerDB)
		if err != nil {
			return fmt.Errorf("open scheduler db: %w", err)
		}
		log := logger.New("cli-scheduler")
		s := scheduler.NewScheduler(db, log)
		es := scheduler.NewEmailScheduler(s)
		runner := NewRunner(es)
		payload := types.CLIArgs{
			EnvPath:     args.EnvPath,
			To:          args.To,
			Subject:     args.Subject,
			Text:        args.Text,
			Template:    args.TemplatePath,
			CSVPath:     args.CSVPath,
			SheetURL:    args.SheetURL,
			Attachments: args.Attachments,
			Cc:          args.Cc,
			Bcc:         args.Bcc,
			Concurrency: args.Concurrency,
			RetryLimit:  args.RetryLimit,
			BatchSize:   args.BatchSize,
			Filter:      args.Filter,
			ScheduleAt:    args.ScheduleAt,
			Interval:      args.Interval,
			Cron:          args.Cron,
			JobRetries:    args.JobRetries,
			JobBackoffDur: args.JobBackoff,
		}
		if err := runner.Run(context.Background(), payload); err != nil {
			return err
		}
		fmt.Printf("📅 Scheduled job for %s (interval=%q cron=%q) in %s\n", args.ScheduleAt, args.Interval, args.Cron, args.SchedulerDB)
		return nil
	}
	// Load SMTP configuration from a file
	cfg, err := config.LoadConfig(args.EnvPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
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

	// Parse CC and BCC addresses from inline or file input
	ccList, err := valid.ParseAddressInput(args.Cc)
	if err != nil {
		return fmt.Errorf("failed to parse CC: %w", err)
	}
	bccList, err := valid.ParseAddressInput(args.Bcc)
	if err != nil {
		return fmt.Errorf("failed to parse BCC: %w", err)
	}

	// Parse Recipients
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

	// Render subject & body for each recipient and build email.Task list
	tasks, err := PrepareEmailTasks(recipients, args.TemplatePath, args.Subject, args.Attachments, ccList, bccList)
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
