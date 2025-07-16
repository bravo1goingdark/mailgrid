package email

import (
	"bytes"
	"fmt"
	"html/template"
	"mailgrid/parser"
	"os"
)

// RenderTemplate loads an HTML template file and executes it with recipient data.
func RenderTemplate(recipient parser.Recipient, templatePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	// Load and parse the HTML template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Combine top-level Email and dynamic Data
	data := map[string]any{
		"Email": recipient.Email,
		"Data":  recipient.Data,
	}

	// Render to buffer
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return out.String(), nil
}
