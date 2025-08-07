package utils

import "strings"

// SplitAndTrim splits comma-separated string and trims whitespace from each email.
func SplitAndTrim(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	// Preallocate the result slice to avoid repeated allocations when
	// trimming a large list of comma-separated values.
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		email := strings.TrimSpace(part)
		if email != "" {
			result = append(result, email)
		}
	}
	return result
}
