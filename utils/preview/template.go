package preview

import (
	"bytes"
	"fmt"
	"html/template"
	"github.com/bravo1goingdark/mailgrid/parser"
	"os"
	"sync"
)

var templateCache sync.Map // Cache parsed templates for reuse

// LoadTemplate parses and caches an HTML template file by its path.
//
// If the template has been parsed before, it returns the cached version.
// Otherwise, it loads and parses the template and stores it in memory.
func LoadTemplate(path string) (*template.Template, error) {
	if tmpl, ok := templateCache.Load(path); ok {
		return tmpl.(*template.Template), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", path)
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	templateCache.Store(path, tmpl)
	return tmpl, nil
}

// RenderTemplate renders an HTML template with the given recipient's data.
//
// It uses caching internally for performance. Recipient values can be accessed in the template via:
//   - {{ .Email }} for the top-level email
//   - {{ .Data.name }}, {{ .Data.age }}, etc. for CSV fields
func RenderTemplate(recipient parser.Recipient, templatePath string) (string, error) {
	tmpl, err := LoadTemplate(templatePath)
	if err != nil {
		return "", err
	}

	// Template context: top-level Email and nested Data map
	data := map[string]any{
		"Email": recipient.Email,
		"Data":  recipient.Data,
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return out.String(), nil
}
