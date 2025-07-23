package cli

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"mailgrid/email"
	"mailgrid/parser"
	"mailgrid/utils/preview"
)

// prepareEmailTasks renders the subject and body templates for each recipient
// and returns a list of email.Task objects ready for sending.
func prepareEmailTasks(recipients []parser.Recipient, templatePath, subjectTpl string) ([]email.Task, error) {
	tmpl, err := template.New("subject").Parse(subjectTpl)
	if err != nil {
		return nil, fmt.Errorf("invalid subject template: %w", err)
	}

	var tasks []email.Task
	for _, r := range recipients {
		// Skip rows with missing fields
		if hasMissingFields(r) {
			log.Printf("⚠️ Skipping %s: missing CSV fields", r.Email)
			continue
		}

		// Render HTML body for recipient
		body, err := preview.RenderTemplate(r, templatePath)
		if err != nil {
			log.Printf("⚠️ Skipping %s: template rendering failed (%v)", r.Email, err)
			continue
		}

		// Render personalized subject line
		var sb bytes.Buffer
		if err := tmpl.Execute(&sb, r.Data); err != nil {
			log.Printf("⚠️ Skipping %s: subject template failed (%v)", r.Email, err)
			continue
		}

		tasks = append(tasks, email.Task{
			Recipient: r,
			Subject:   sb.String(),
			Body:      body,
			Retries:   0,
		})
	}
	return tasks, nil
}

// hasMissingFields returns true if any field in recipient data is empty.
func hasMissingFields(r parser.Recipient) bool {
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
		fmt.Printf(" Email #%d → %s\nSubject: %s\n\n%s\n\n", i+1, t.Recipient.Email, t.Subject, t.Body)
	}
	fmt.Printf(" Dry-run complete: %d emails rendered\n", len(tasks))
}
