package cli

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/utils"
	"github.com/bravo1goingdark/mailgrid/webhook"
)

// PrepareEmailTasks renders the subject and body templates for each recipient
// and returns a list of email.Task objects ready for sending.
//
// plainText is optional plain-text content (inline string or file path resolved
// by the caller). When both templatePath (HTML) and plainText are provided, the
// task carries both bodies so the sender can build a multipart/alternative message.
func PrepareEmailTasks(recipients []parser.Recipient, templatePath, plainText, subjectTpl string, attachments []string, ccList []string, bccList []string) ([]email.Task, error) {
	tmpl, err := template.New("subject").Option("missingkey=error").Parse(subjectTpl)
	if err != nil {
		return nil, fmt.Errorf("invalid subject template: %w", err)
	}

	var tasks []email.Task
	skipped := 0
	for i, r := range recipients {
		// Skip rows with missing fields
		if HasMissingFields(r) {
			log.Printf("️ Skipping %s: missing CSV fields", r.Email)
			skipped++
			continue
		}

		var body string
		if templatePath != "" {
			var err error
			body, err = utils.RenderTemplate(r, templatePath)
			if err != nil {
				log.Printf("️ Skipping %s: template rendering failed (%v)", r.Email, err)
				skipped++
				continue
			}
		}

		// Render personalized subject line
		var sb bytes.Buffer
		if err := tmpl.Execute(&sb, r.Data); err != nil {
			log.Printf("️ Skipping %s: subject template failed (%v)", r.Email, err)
			skipped++
			continue
		}

		tasks = append(tasks, email.Task{
			Recipient:   r,
			Subject:     sb.String(),
			Body:        body,
			PlainText:   plainText,
			Attachments: attachments,
			CC:          ccList,
			BCC:         bccList,
			Retries:     0,
			Index:       i, // Add index for offset tracking
		})
	}

	if skipped > 0 {
		log.Printf("Prepared %d tasks. Skipped %d recipient(s) (missing fields or render errors).", len(tasks), skipped)
	}

	return tasks, nil
}

// HasMissingFields returns true if the recipient email is empty.
// Other data fields may be optional depending on the template, so only
// the email field is required.
func HasMissingFields(r parser.Recipient) bool {
	return r.Email == ""
}

// printDryRun logs rendered email content to the console instead of sending.
func printDryRun(tasks []email.Task) {
	for i, t := range tasks {
		fmt.Printf(" Email #%d → %s\nSubject: %s\n", i+1, t.Recipient.Email, t.Subject)
		if len(t.Attachments) > 0 {
			fmt.Printf("Attachments: %v\n", t.Attachments)
		}
		switch {
		case t.Body != "" && t.PlainText != "":
			fmt.Printf("\n[plain text]\n%s\n\n[html]\n%s\n\n", t.PlainText, t.Body)
		case t.Body != "":
			fmt.Printf("\n%s\n\n", t.Body)
		case t.PlainText != "":
			fmt.Printf("\n%s\n\n", t.PlainText)
		default:
			fmt.Printf("\n(no body)\n\n")
		}
	}
	fmt.Printf(" Dry-run complete: %d emails rendered\n", len(tasks))
}

// SendSingleEmail handles one-off email sending using --to.
//
// Accepted combinations:
//   - --template only: HTML email
//   - --text only:     plain-text email
//   - --template + --text: multipart/alternative (HTML + plain text)
func SendSingleEmail(args CLIArgs, cfg config.SMTPConfig) error {
	if args.To == "" {
		return fmt.Errorf("--to flag is required for single email sending")
	}
	if args.TemplatePath == "" && args.Text == "" {
		return fmt.Errorf("either --template or --text must be provided")
	}

	// Build a single recipient with minimal substitution map
	recipient := parser.Recipient{
		Email: args.To,
		Data: map[string]string{
			"email": args.To,
		},
	}

	// Resolve plain-text input (inline string or file path)
	var plainText string
	var err error
	if args.Text != "" {
		plainText, err = utils.ReadTextInput(args.Text)
		if err != nil {
			return fmt.Errorf("failed to read plain-text body: %w", err)
		}
	}

	ccList := utils.SplitAndTrim(args.Cc)
	bccList := utils.SplitAndTrim(args.Bcc)

	tasks, err := PrepareEmailTasks(
		[]parser.Recipient{recipient},
		args.TemplatePath,
		plainText,
		args.Subject,
		args.Attachments,
		ccList,
		bccList,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare task: %w", err)
	}
	if len(tasks) == 0 {
		return fmt.Errorf("no task generated (maybe due to template/rendering failure)")
	}

	if args.DryRun {
		printDryRun(tasks)
		return nil
	}

	start := time.Now()
	jobID := fmt.Sprintf("mailgrid-single-%d", start.Unix())

	email.SetRetryLimit(args.RetryLimit)

	// Initialize monitoring for single email if enabled
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
			TemplateFile:      args.TemplatePath,
			ConcurrentWorkers: 1,
			BatchSize:         1,
			RetryLimit:        args.RetryLimit,
		}
		mon.InitializeCampaign(jobID, configSummary, 1)

		fmt.Printf("  Monitor dashboard: http://localhost:%d\n", args.MonitorPort)

		// Cleanup monitor after completion
		defer func() {
			go func() {
				if err := monitorServer.Stop(); err != nil {
					log.Printf("Failed to stop monitor server: %v", err)
				}
			}()
		}()
	}

	dispatchResult := email.StartDispatcher(tasks, cfg, 1, 1, &email.DispatchOptions{
		Context: context.Background(),
		Monitor: mon,
	})
	duration := time.Since(start)

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

		// Create webhook payload for single email
		result := webhook.CampaignResult{
			JobID:                jobID,
			Status:               "completed",
			TotalRecipients:      1,
			SuccessfulDeliveries: successfulDeliveries,
			FailedDeliveries:     failedDeliveries,
			StartTime:            start,
			EndTime:              endTime,
			DurationSeconds:      int(duration.Seconds()),
			ConcurrentWorkers:    1,
		}

		// Set template file if provided
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
