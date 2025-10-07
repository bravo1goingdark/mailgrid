package cli

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/utils"
	"github.com/bravo1goingdark/mailgrid/utils/preview"
	"github.com/bravo1goingdark/mailgrid/webhook"
)

// PrepareEmailTasks renders the subject and body templates for each recipient
// and returns a list of email.Task objects ready for sending.
func PrepareEmailTasks(recipients []parser.Recipient, templatePath, subjectTpl string, attachments []string, ccList []string, bccList []string) ([]email.Task, error) {
	tmpl, err := template.New("subject").Parse(subjectTpl)
	if err != nil {
		return nil, fmt.Errorf("invalid subject template: %w", err)
	}

	var tasks []email.Task
	for _, r := range recipients {
		// Skip rows with missing fields
		if HasMissingFields(r) {
			log.Printf("⚠️ Skipping %s: missing CSV fields", r.Email)
			continue
		}

		var body string
		if templatePath != "" {
			var err error
			body, err = preview.RenderTemplate(r, templatePath)
			if err != nil {
				log.Printf("⚠️ Skipping %s: template rendering failed (%v)", r.Email, err)
				continue
			}
		}

		// Render personalized subject line
		var sb bytes.Buffer
		if err := tmpl.Execute(&sb, r.Data); err != nil {
			log.Printf("⚠️ Skipping %s: subject template failed (%v)", r.Email, err)
			continue
		}

		tasks = append(tasks, email.Task{
			Recipient:   r,
			Subject:     sb.String(),
			Body:        body,
			Attachments: attachments,
			CC:          ccList,
			BCC:         bccList,
			Retries:     0,
		})
	}
	return tasks, nil
}

// HasMissingFields returns true if any field in recipient data is empty.
func HasMissingFields(r parser.Recipient) bool {
	for _, val := range r.Data {
		if val == "" {
			return true
		}
	}
	return false
}

// printDryRun logs rendered email content to the console instead of sending.
func printDryRun(tasks []email.Task) {
	for i, t := range tasks {
		fmt.Printf(" Email #%d → %s\nSubject: %s\n", i+1, t.Recipient.Email, t.Subject)
		if len(t.Attachments) > 0 {
			fmt.Printf("Attachments: %v\n", t.Attachments)
		}
		if t.Body != "" {
			fmt.Printf("\n%s\n\n", t.Body)
		} else {
			fmt.Printf("\n(no body)\n\n")
		}
	}
	fmt.Printf(" Dry-run complete: %d emails rendered\n", len(tasks))
}

// SendSingleEmail handles one-off email sending using --to along with either --template or --text (mutually exclusive).
func SendSingleEmail(args CLIArgs, cfg config.SMTPConfig) error {
	if args.To == "" {
		return fmt.Errorf("--to flag is required for single email sending")
	}
	if (args.TemplatePath == "" && args.Text == "") || (args.TemplatePath != "" && args.Text != "") {
		return fmt.Errorf("either --template or --text must be provided, but not both")
	}

	// Build a single recipient with minimal substitution map
	recipient := parser.Recipient{
		Email: args.To,
		Data: map[string]string{
			"email": args.To, // Can be expanded with more CLI-provided fields in future
		},
	}

	var templatePath string
	var body string
	var err error

	if args.TemplatePath != "" {
		templatePath = args.TemplatePath
	} else {
		body, err = utils.ReadTextInput(args.Text)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}
	}

	ccList := utils.SplitAndTrim(args.Cc);
	bccList := utils.SplitAndTrim(args.Bcc);

	// Use existing logic to render subject and body
	tasks, err := PrepareEmailTasks(
		[]parser.Recipient{recipient},
		templatePath,
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

	// If --text is used, override body (PrepareEmailTasks would leave it empty)
	if args.TemplatePath == "" {
		tasks[0].Body = body
	}

	if args.DryRun {
		printDryRun(tasks)
		return nil
	}

	start := time.Now()
	jobID := fmt.Sprintf("mailgrid-single-%d", start.Unix())

	email.SetRetryLimit(args.RetryLimit)
	email.StartDispatcher(tasks, cfg, 1, 1)
	duration := time.Since(start)

	// Send webhook notification if URL is provided
	if args.WebhookURL != "" {
		endTime := time.Now()

		// Create webhook payload for single email
		result := webhook.CampaignResult{
			JobID:                jobID,
			Status:               "completed",
			TotalRecipients:      1,
			SuccessfulDeliveries: 1, // TODO: Get actual success count
			FailedDeliveries:     0, // TODO: Get actual failure count
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
			fmt.Printf("⚠️ Failed to send webhook notification: %v\n", err)
		} else {
			fmt.Printf("🔔 Webhook notification sent to %s\n", args.WebhookURL)
		}
	}

	return nil
}
