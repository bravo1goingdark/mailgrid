package utils

import (
	"fmt"
	"os"
	"strings"
)

// ParseAddressInput to parse --cc / --bcc from inline or file input
func ParseAddressInput(input string) ([]string, error) {
	if input == "" {
		return nil, nil
	}
	if strings.Contains(input, "@") {
		// Inline mode
		parts := strings.Split(input, ",")
		emails := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				emails = append(emails, p)
			}
		}
		return emails, nil
	}

	// File mode
	content, err := os.ReadFile(input)
	if err != nil {
		return nil, fmt.Errorf("failed to read address file: %w", err)
	}

	var emails []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			emails = append(emails, line)
		}
	}
	return emails, nil
}
