package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/bravo1goingdark/mailgrid/parser"
)

// ParseAddressInput to parse --cc / --bcc from inline or file input.
// If the input is an existing file path, it reads emails from the file.
// Otherwise, it treats the input as comma-separated emails.
func ParseAddressInput(input string) ([]string, error) {
	if input == "" {
		return nil, nil
	}
	// If the file exists, treat as file mode
	if info, err := os.Stat(input); err == nil && !info.IsDir() {
		content, readErr := os.ReadFile(input)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read address file: %w", readErr)
		}
		var emails []string
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && parser.IsValidEmail(line) {
				emails = append(emails, line)
			}
		}
		return emails, nil
	}

	// Inline mode: split on commas
	parts := strings.Split(input, ",")
	emails := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && parser.IsValidEmail(p) {
			emails = append(emails, p)
		}
	}
	return emails, nil
}
