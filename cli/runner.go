package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
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
	// Create context with signal cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on Ctrl+C / SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	var wg sync.WaitGroup
	defer wg.Wait()

	go func() {
		select {
		case <-sigChan:
			fmt.Println("\n Received interrupt signal, shutting down gracefully...")
			cancel()
		case <-ctx.Done():
		}
	}()

	// Handle --jobs-list flag
	if args.ListJobs {
		return listScheduledJobs(args.DBPath, args.EnvPath)
	}

	// Handle --jobs-cancel flag
	if args.CancelJobID != "" {
		return cancelScheduledJob(args.DBPath, args.EnvPath, args.CancelJobID)
	}

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
			DBPath:          args.DBPath,
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
			if runAt.Before(time.Now()) {
				log.Printf("Warning: --schedule-at is in the past; job will run immediately")
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
		fmt.Printf("[DATABASE]  Database: %s\n", args.DBPath)
		fmt.Printf(" The scheduler will start automatically and run in the background\n")

		return nil
	}
	if args.EnvPath == "" {
		return fmt.Errorf("SMTP config required: set --env path/to/config.json")
	}
	// Load SMTP configuration from a file
	cfg, err := config.LoadConfig(args.EnvPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if err := config.Validate(cfg.SMTP); err != nil {
		return fmt.Errorf("invalid SMTP config: %w", err)
	}
	// Wire configurable SMTP dial timeout.
	if args.SMTPTimeout > 0 {
		cfg.SMTP.DialTimeout = time.Duration(args.SMTPTimeout) * time.Second
	}
	if args.Concurrency < 1 {
		return fmt.Errorf("--concurrency must be at least 1")
	}
	if args.BatchSize < 1 {
		args.BatchSize = 1
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
		file, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("attachment not readable: %s", f)
		}
		file.Close()
	}

	if args.TemplatePath == "" && args.Text == "" && len(args.Attachments) == 0 {
		return fmt.Errorf("provide --template, --text, or --attach (at least one is required)")
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

	if len(recipients) == 0 && !args.DryRun {
		return fmt.Errorf("no recipients found (CSV/Sheet is empty or all rows were skipped)")
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

	// Resolve optional plain-text body for multipart/alternative sending.
	var plainText string
	if args.Text != "" && args.TemplatePath != "" {
		// Both provided: HTML template + plain-text → multipart/alternative
		var terr error
		plainText, terr = utils.ReadTextInput(args.Text)
		if terr != nil {
			return fmt.Errorf("failed to read plain-text body: %w", terr)
		}
	}

	// Render subject & body for each recipient and build email.Task list
	tasks, err := PrepareEmailTasks(recipients, args.TemplatePath, plainText, args.Subject, args.Attachments, ccList, bccList)
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
				log.Printf("⚠️ Warning: Failed to reset offset: %v", err)
			} else {
				fmt.Println(" Offset file cleared, starting from beginning")
			}
		}

		// Load existing offset if resume is enabled
		if args.Resume {
			if err := tracker.Load(); err != nil {
				log.Printf("⚠️ Warning: Failed to load offset (starting from beginning): %v", err)
			} else {
				startOffset = tracker.GetOffset()
				if startOffset > 0 {
					if startOffset >= len(tasks) {
						fmt.Printf(" All emails already sent (offset: %d, total: %d)\n", startOffset, len(tasks))
						return nil
					}
					fmt.Printf(" Resuming from offset %d (skipping %d already sent emails)\n", startOffset, startOffset)
					tasks = tasks[startOffset:]
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
		monitorServer = monitor.NewServer(args.MonitorPort, time.Duration(args.MonitorClientTimeout)*time.Second)
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
		Context:     ctx,
		Monitor:     mon,
		Tracker:     tracker,
		StartOffset: startOffset,
	}
	dispatchResult := email.StartDispatcher(tasks, cfg.SMTP, args.Concurrency, args.BatchSize, opts)
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
			if err := monitorServer.Stop(); err != nil {
				log.Printf("Failed to stop monitor server: %v", err)
			}
		}()
	}

	fmt.Printf("\u2705 Completed in %s using %d workers\n", duration, args.Concurrency)

	// Send webhook notification if URL is provided
	if args.WebhookURL != "" {
		endTime := time.Now()

		// Use dispatch result for stats (works with or without monitor)
		successfulDeliveries := dispatchResult.Sent
		failedDeliveries := dispatchResult.Failed

		// If monitor is enabled, prefer its counts (includes retries)
		if monitorServer != nil {
			stats := monitorServer.GetStats()
			if stats.SentCount > 0 || stats.FailedCount > 0 {
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

		// Send webhook notification (signed with HMAC if secret is provided)
		var webhookClient *webhook.Client
		if args.WebhookSecret != "" {
			webhookClient = webhook.NewClientWithSecret(args.WebhookSecret)
		} else {
			webhookClient = webhook.NewClient()
		}
		if err := webhookClient.SendNotification(args.WebhookURL, result); err != nil {
			fmt.Printf("️ Failed to send webhook notification: %v\n", err)
		} else {
			fmt.Printf(" Webhook notification sent to %s\n", args.WebhookURL)
		}
	}

	return nil
}

// listScheduledJobs lists all scheduled jobs from the database
func listScheduledJobs(dbPath, envPath string) error {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	log := logger.New("scheduler")
	sched := scheduler.NewScheduler(db, log)

	jobs, err := sched.ListJobs()
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	if len(jobs) == 0 {
		fmt.Println("No scheduled jobs found.")
		return nil
	}

	fmt.Printf("Found %d job(s):\n\n", len(jobs))
	fmt.Printf("%-30s %-15s %-20s %s\n", "JOB ID", "STATUS", "RUN AT", "CREATED AT")
	fmt.Println("--------------------------------------------------------------------")

	for _, job := range jobs {
		runAt := job.RunAt.Format("2006-01-02 15:04:05")
		createdAt := job.CreatedAt.Format("2006-01-02 15:04:05")
		fmt.Printf("%-30s %-15s %-20s %s\n", job.ID, job.Status, runAt, createdAt)
	}

	return nil
}

// cancelScheduledJob cancels a scheduled job by ID
func cancelScheduledJob(dbPath, envPath, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("job ID required (--jobs-cancel)")
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	log := logger.New("scheduler")
	sched := scheduler.NewScheduler(db, log)

	success := sched.CancelJob(jobID)
	if !success {
		return fmt.Errorf("job not found or could not be cancelled: %s", jobID)
	}

	fmt.Printf("Job %s cancelled successfully.\n", jobID)
	return nil
}
