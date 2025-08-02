package valid

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
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts, nil
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
