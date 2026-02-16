package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/offset"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/scheduler"
	"github.com/bravo1goingdark/mailgrid/utils"
	"github.com/bravo1goingdark/mailgrid/webhook"
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
		// Load SMTP config for the scheduler
		smtpConfig, err := config.LoadConfig(args.EnvPath)
		if err != nil {
			return fmt.Errorf("failed to load SMTP config: %w", err)
		}

		// Configure optimized scheduler manager
		config := scheduler.DefaultOptimizedConfig()
		managerConfig := scheduler.ManagerConfig{
			DBPath:          "mailgrid.db",
			SMTPConfig:      smtpConfig.SMTP,
			OptimizedConfig: config,
			ShutdownDelay:   5 * time.Minute,
			AutoShutdown:    true,
		}

		// Initialize global scheduler manager
		scheduler.InitGlobalManager(managerConfig)
		manager := scheduler.GetGlobalManager()

		// Create job handler
		handler := func(job types.Job) error {
			var a types.CLIArgs
			if err := json.Unmarshal(job.Args, &a); err != nil {
				return fmt.Errorf("decode job args: %w", err)
			}

			// Execute the job based on type
			if a.To != "" {
				// Single email
				cliArgs := CLIArgs{
					EnvPath:      a.EnvPath,
					To:           a.To,
					Subject:      a.Subject,
					TemplatePath: a.Template,
					Text:         a.Text,
					Attachments:  a.Attachments,
					Cc:           a.Cc,
					Bcc:          a.Bcc,
					RetryLimit:   a.RetryLimit,
				}
				return SendSingleEmail(cliArgs, smtpConfig.SMTP)
			} else {
				// Bulk email
				cliArgs := CLIArgs{
					EnvPath:      a.EnvPath,
					CSVPath:      a.CSVPath,
					SheetURL:     a.SheetURL,
					TemplatePath: a.Template,
					Subject:      a.Subject,
					Attachments:  a.Attachments,
					Cc:           a.Cc,
					Bcc:          a.Bcc,
					Concurrency:  a.Concurrency,
					RetryLimit:   a.RetryLimit,
					BatchSize:    a.BatchSize,
					Filter:       a.Filter,
				}
				return Run(cliArgs)
			}
		}

		// Parse schedule time
		var runAt time.Time
		if args.ScheduleAt != "" {
			var err error
			runAt, err = time.Parse(time.RFC3339, args.ScheduleAt)
			if err != nil {
				return fmt.Errorf("parse schedule_at: %w", err)
			}
		} else {
			runAt = time.Now()
		}

		// Create job payload
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
			ScheduleAt:  args.ScheduleAt,
			Interval:    args.Interval,
			Cron:        args.Cron,
			JobRetries:  args.JobRetries,
		}

		// Schedule the job (this will auto-start the scheduler)
		if err := manager.ScheduleJob(payload, runAt, args.Cron, args.Interval, handler); err != nil {
			return fmt.Errorf("failed to schedule job: %w", err)
		}

		scheduleInfo := ""
		if args.ScheduleAt != "" {
			scheduleInfo = fmt.Sprintf(" at %s", args.ScheduleAt)
		}
		if args.Interval != "" {
			scheduleInfo += fmt.Sprintf(" every %s", args.Interval)
		}
		if args.Cron != "" {
			scheduleInfo += fmt.Sprintf(" using cron %q", args.Cron)
		}

		fmt.Printf("[SCHEDULE] Job scheduled successfully%s\n", scheduleInfo)
		fmt.Printf("[DATABASE]  Database: mailgrid.db\n")
		fmt.Printf(" Metrics: http://localhost:8090/metrics\n")
		fmt.Printf(" The scheduler will start automatically and run in the background\n")

		return nil
	}
	// Load SMTP configuration from a file
	cfg, err := config.LoadConfig(args.EnvPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if args.To != "" {
		if args.CSVPath != "" || args.SheetURL != "" {
			return fmt.Errorf(" --to is mutually exclusive with --csv and --sheet-url")
		}

		return SendSingleEmail(args, cfg.SMTP)
	}
	if args.CSVPath == "" && args.SheetURL == "" {
		return fmt.Errorf(" You must provide either --csv or --sheet-url")
	}
	if args.CSVPath != "" && args.SheetURL != "" {
		return fmt.Errorf(" Provide only one of --csv or --sheet-url, not both")
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

	// Validate webhook URL if provided
	if err := webhook.ValidateURL(args.WebhookURL); err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Parse CC and BCC addresses from inline or file input
	ccList, err := utils.ParseAddressInput(args.Cc)
	if err != nil {
		return fmt.Errorf("failed to parse CC: %w", err)
	}
	bccList, err := utils.ParseAddressInput(args.Bcc)
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
			if closeErr := stream.Close(); closeErr != nil {
				log.Printf("Warning: Failed to close Google Sheet stream: %v", closeErr)
			}
		}(stream)

		recipients, err = parser.ParseCSVFromReader(stream)
		if err != nil {
			return fmt.Errorf("failed to parse Google Sheet as CSV: %w", err)
		}

		id, gid, _ := parser.ExtractSheetInfo(args.SheetURL)
		fmt.Printf(" Loaded Google Sheet: Spreadsheet ID = %s, GID = %s\n", id, gid)

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

		expr, err := parser.ParseExpression(args.Filter)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}

		if err := parser.ValidateFields(expr, recipients); err != nil {
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
		rendered, err := utils.RenderTemplate(recipients[0], args.TemplatePath)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		return utils.StartPreviewServer(rendered, args.PreviewPort)
	}

	// Render subject & body for each recipient and build email.Task list
	tasks, err := PrepareEmailTasks(recipients, args.TemplatePath, args.Subject, args.Attachments, ccList, bccList)
	if err != nil {
		return err
	}

	// Initialize offset tracker for resumable delivery
	var tracker *offset.Tracker
	var startOffset int

	// Handle offset tracking (only for bulk operations, not single emails)
	if len(tasks) > 1 {
		tracker = offset.NewTracker(".mailgrid.offset")

		// Handle reset-offset flag
		if args.ResetOffset {
			if err := tracker.Reset(); err != nil {
				log.Printf("️ Warning: Failed to reset offset: %v", err)
			} else {
				fmt.Println(" Offset file cleared, starting from beginning")
			}
		}

		// Load existing offset if resume is enabled
		if args.Resume {
			if err := tracker.Load(); err != nil {
				log.Printf("️ Warning: Failed to load offset (starting from beginning): %v", err)
			} else {
				startOffset = tracker.GetOffset()
				if startOffset > 0 {
					if startOffset >= len(tasks) {
						fmt.Printf(" All emails already sent (offset: %d, total: %d)\n", startOffset, len(tasks))
						return nil
					}
					fmt.Printf(" Resuming from offset %d (skipping %d already sent emails)\n", startOffset, startOffset)
					tasks = tasks[startOffset:] // Skip already sent emails
				}
			}
		}

		// Generate unique job ID and set it in tracker
		jobID := fmt.Sprintf("mailgrid-%d", time.Now().Unix())
		if tracker != nil {
			tracker.SetJobID(jobID)
		}
	}

	// If dry-run mode, print emails and skip sending
	if args.DryRun {
		printDryRun(tasks)
		return nil
	}

	// Otherwise, send emails using dispatcher
	start := time.Now()
	email.SetRetryLimit(args.RetryLimit)

	// Use existing job ID from tracker or generate new one
	var jobID string
	if tracker != nil && tracker.GetJobID() != "" {
		jobID = tracker.GetJobID()
	} else {
		jobID = fmt.Sprintf("mailgrid-%d", start.Unix())
	}

	// Initialize monitoring if enabled
	var mon monitor.Monitor = monitor.NewNoOpMonitor()
	var monitorServer *monitor.Server

	if args.Monitor {
		monitorServer = monitor.NewServer(args.MonitorPort)
		mon = monitorServer

		// Start monitoring server in background
		go func() {
			if err := monitorServer.Start(); err != nil && err != http.ErrServerClosed {
				log.Printf("️ Monitor server failed: %v", err)
			}
		}()

		// Initialize campaign tracking
		configSummary := monitor.ConfigSummary{
			CSVFile:           args.CSVPath,
			SheetURL:          args.SheetURL,
			TemplateFile:      args.TemplatePath,
			ConcurrentWorkers: args.Concurrency,
			BatchSize:         args.BatchSize,
			RetryLimit:        args.RetryLimit,
			FilterExpression:  args.Filter,
		}
		mon.InitializeCampaign(jobID, configSummary, len(tasks))

		fmt.Printf("  Monitor dashboard: http://localhost:%d\n", args.MonitorPort)
	}

	// Use offset-aware dispatcher if tracker is available
	opts := &email.DispatchOptions{
		Context:     context.Background(),
		Monitor:     mon,
		Tracker:     tracker,
		StartOffset: startOffset,
	}
	email.StartDispatcher(tasks, cfg.SMTP, args.Concurrency, args.BatchSize, opts)
	// Save final offset after campaign completion
	if tracker != nil {
		if err := tracker.Save(); err != nil {
			log.Printf("️ Warning: Failed to save final offset: %v", err)
		}
	}
	duration := time.Since(start)

	// Cleanup monitoring server if it was started
	if monitorServer != nil {
		go func() {
			// Use context with timeout instead of sleep for cleaner shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			<-ctx.Done() // Give users time to see final results
			if err := monitorServer.Stop(); err != nil {
				log.Printf("️ Failed to stop monitor server: %v", err)
			}
		}()
	}

	fmt.Printf("\u2705 Completed in %s using %d workers\n", duration, args.Concurrency)

	// Send webhook notification if URL is provided
	if args.WebhookURL != "" {
		endTime := time.Now()

		// Get actual delivery statistics from monitor
		var successfulDeliveries, failedDeliveries int
		if monitorServer != nil {
			stats := monitorServer.GetStats()
			if stats != nil {
				successfulDeliveries = stats.SentCount
				failedDeliveries = stats.FailedCount
			}
		}

		// Create webhook payload
		result := webhook.CampaignResult{
			JobID:                jobID,
			Status:               "completed",
			TotalRecipients:      len(tasks),
			SuccessfulDeliveries: successfulDeliveries,
			FailedDeliveries:     failedDeliveries,
			StartTime:            start,
			EndTime:              endTime,
			DurationSeconds:      int(duration.Seconds()),
			ConcurrentWorkers:    args.Concurrency,
		}

		// Set file paths
		if args.CSVPath != "" {
			result.CSVFile = args.CSVPath
		}
		if args.SheetURL != "" {
			result.SheetURL = args.SheetURL
		}
		if args.TemplatePath != "" {
			result.TemplateFile = args.TemplatePath
		}

		// Send webhook notification
		webhookClient := webhook.NewClient()
		if err := webhookClient.SendNotification(args.WebhookURL, result); err != nil {
			fmt.Printf("️ Failed to send webhook notification: %v\n", err)
		} else {
			fmt.Printf(" Webhook notification sent to %s\n", args.WebhookURL)
		}
	}

	return nil
}
