package email

import (
	"testing"
)

func TestExtractSMTPCode(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "error message with 550 code at start",
			errMsg:   "550 5.1.1 User unknown",
			expected: "550",
		},
		{
			name:     "error message with 421 code",
			errMsg:   "421 4.7.0 Try again later",
			expected: "421",
		},
		{
			name:     "error message with 450 code",
			errMsg:   "450 4.1.8 Domain not found",
			expected: "450",
		},
		{
			name:     "error message with 451 code",
			errMsg:   "451 4.4.2 Timeout",
			expected: "451",
		},
		{
			name:     "error message with 452 code",
			errMsg:   "452 4.2.2 Over quota",
			expected: "452",
		},
		{
			name:     "error message with 551 code",
			errMsg:   "551 5.1.1 User not local",
			expected: "551",
		},
		{
			name:     "error message with 552 code",
			errMsg:   "552 5.2.2 Message too large",
			expected: "552",
		},
		{
			name:     "error message with 553 code",
			errMsg:   "553 5.3.0 Invalid address",
			expected: "553",
		},
		{
			name:     "error message with 554 code",
			errMsg:   "554 5.7.1 Message rejected",
			expected: "554",
		},
		{
			name:     "error message with code embedded",
			errMsg:   "Server said: 550 5.1.1 User unknown",
			expected: "550",
		},
		{
			name:     "error message with no code",
			errMsg:   "Some random error message",
			expected: "",
		},
		{
			name:     "empty error message",
			errMsg:   "",
			expected: "",
		},
		{
			name:     "short error message",
			errMsg:   "OK",
			expected: "",
		},
		{
			name:     "error with multiple codes",
			errMsg:   "550 error, then 421 later",
			expected: "550",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSMTPCode(tt.errMsg)
			if result != tt.expected {
				t.Errorf("extractSMTPCode(%q) = %q, expected %q", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestSetRetryLimit(t *testing.T) {
	// Save original value
	originalLimit := retryLimit
	defer func() {
		retryLimit = originalLimit
	}()

	tests := []struct {
		name  string
		limit int
	}{
		{
			name:  "set to 0",
			limit: 0,
		},
		{
			name:  "set to 1",
			limit: 1,
		},
		{
			name:  "set to 5",
			limit: 5,
		},
		{
			name:  "set to 10",
			limit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetRetryLimit(tt.limit)
			if retryLimit != tt.limit {
				t.Errorf("SetRetryLimit(%d) did not set retryLimit to %d, got %d", tt.limit, tt.limit, retryLimit)
			}
		})
	}
}
