package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/pflag"
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
		preview      bool
		previewPort  int
	)

	// flag definitions
	pflag.StringVar(&envPath, "env", "example/config.json", "Path to SMTP config JSON (required)")
	pflag.StringVar(&csvPath, "csv", "example/test_contacts.csv", "Path to recipient CSV file (required)")
	pflag.StringVarP(&templatePath, "template", "t", "example/welcome.html", "Path to email HTML template")
	pflag.StringVarP(&subject, "subject", "s", "Test Email from Mailgrid", "Subject line of the email")
	pflag.BoolVar(&dryRun, "dry-run", false, "Render emails to console without sending")
	pflag.BoolVar(&preview, "preview", false, "Start a local server to preview the rendered email in browser")
	pflag.IntVar(&previewPort, "preview-port", 8080, "Port for preview server (default 8080)")

	pflag.Parse()

	// Load config from JSON
	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Parse recipients from CSV
	recipients, err := parser.ParseCSV(csvPath)
	if err != nil {
		log.Fatalf("❌ Failed to parse CSV: %v", err)
	}

	if preview {
		if len(recipients) == 0 {
			log.Fatalf("No recipients found in CSV for preview.")
		}
		first := recipients[0]
		rendered, err := email.RenderTemplate(first, templatePath)
		if err != nil {
			log.Fatalf("Failed to render template for preview: %v", err)
		}
		addr := fmt.Sprintf(":%d", previewPort)
		log.Printf("🌐 Preview server running at http://localhost:%d (Ctrl+C to stop)", previewPort)
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(rendered))
		}
		http.HandleFunc("/", handler)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start preview server: %v", err)
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
