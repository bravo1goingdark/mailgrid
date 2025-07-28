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

// PrepareEmailTasks renders the subject and body templates for each recipient
// and returns a list of email.Task objects ready for sending.
func PrepareEmailTasks(recipients []parser.Recipient, templatePath, subjectTpl string, attachments []string) ([]email.Task, error) {
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
