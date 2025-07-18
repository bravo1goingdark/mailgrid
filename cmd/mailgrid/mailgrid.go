package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	preview "mailgrid/cmd/preview"
	"mailgrid/config"
	"mailgrid/email"
	"mailgrid/parser"
)

func main() {
	var (
		envPath      string
		csvPath      string
		templatePath string
		subject      string
		dryRun       bool
		showPreview  bool
		previewPort  int
	)

	// flag definitions
	pflag.StringVar(&envPath, "env", "example/config.json", "Path to SMTP config JSON (required)")
	pflag.StringVar(&csvPath, "csv", "example/test_contacts.csv", "Path to recipient CSV file (required)")
	pflag.StringVarP(&templatePath, "template", "t", "example/welcome.html", "Path to email HTML template")
	pflag.StringVarP(&subject, "subject", "s", "Test Email from Mailgrid", "Subject line of the email")
	pflag.BoolVar(&dryRun, "dry-run", false, "Render emails to console without sending")
	pflag.BoolVar(&showPreview, "preview", false, "Start a local server to preview the rendered email in browser")
	pflag.IntVar(&previewPort, "preview-port", 8080, "Port for preview server (default 8080)")

	pflag.Parse()

	// Load config from JSON
	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Parse recipients from CSV
	recipients, err := parser.ParseCSV(csvPath)
	if err != nil {
		fmt.Printf("❌ Failed to parse CSV: %v\n", err)
		os.Exit(1)
	}

	if showPreview {
		if len(recipients) == 0 {
			fmt.Println("No recipients found in CSV for preview.")
			os.Exit(1)
		}
		first := recipients[0]
		rendered, err := email.RenderTemplate(first, templatePath)
		if err != nil {
			fmt.Printf("Failed to render template for preview: %v\n", err)
			os.Exit(1)
		}
		if err := preview.StartServer(rendered, previewPort); err != nil {
			fmt.Printf("❌ Preview server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Summary counters
	sentCount := 0
	failCount := 0
	skippedCount := 0

	// Iterate through recipients and process emails
	for i, r := range recipients {
		// Render the email template with dynamic fields
		rendered, err := email.RenderTemplate(r, templatePath)
		if err != nil {
			log.Printf("❌ Failed to render email for %s: %v", r.Email, err)
			failCount++
			continue
		}

		// ⚠️ Check for any missing fields in the CSV
		var missingFields []string
		for key, val := range r.Data {
			if val == "" {
				missingFields = append(missingFields, key)
			}
		}
		if len(missingFields) > 0 {
			log.Printf("⚠️ Missing fields [%v] in CSV for %s", missingFields, r.Email)
			skippedCount++
			continue
		}

		// If dry-run, just print the email to console
		if dryRun {
			fmt.Printf("📩 Email #%d for %s (dry-run):\n%s\n\n", i+1, r.Email, rendered)
			sentCount++
			continue
		}

		// Send the email via SMTP
		err = email.SendEmail(cfg.SMTP, r.Email, subject, rendered)
		if err != nil {
			log.Printf("❌ Failed to send email to %s: %v", r.Email, err)
			failCount++
		} else {
			log.Printf("✅ Sent email to %s", r.Email)
			sentCount++
		}
	}

	// final summary
	fmt.Println()
	fmt.Printf("📊 Summary: Sent ✅ %d | Failed ❌ %d | Skipped ⚠️ %d\n", sentCount, failCount, skippedCount)
}
