package email

import (
	"testing"
)

func TestMaxInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a greater than b",
			a:        10,
			b:        5,
			expected: 10,
		},
		{
			name:     "b greater than a",
			a:        5,
			b:        10,
			expected: 10,
		},
		{
			name:     "a equals b",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative numbers",
			a:        -10,
			b:        -5,
			expected: -5,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        5,
			expected: 5,
		},
		{
			name:     "both zero",
			a:        0,
			b:        0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("maxInt(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestTaskStruct(t *testing.T) {
	task := Task{
		Recipient: struct {
			Email string
			Data  map[string]string
		}{Email: "test@example.com", Data: map[string]string{"name": "Test"}},
		Subject:     "Test Subject",
		Body:        "Test Body",
		Retries:     0,
		Attachments: []string{"file1.pdf", "file2.pdf"},
		CC:          []string{"cc@example.com"},
		BCC:         []string{"bcc@example.com"},
		Index:       1,
	}

	if task.Subject != "Test Subject" {
		t.Errorf("Expected Subject 'Test Subject', got '%s'", task.Subject)
	}

	if task.Index != 1 {
		t.Errorf("Expected Index 1, got %d", task.Index)
	}

	if len(task.Attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(task.Attachments))
	}
}

func TestDispatchOptionsDefaults(t *testing.T) {
	opts := &DispatchOptions{}

	// Test that fields can be set
	opts.StartOffset = 10
	if opts.StartOffset != 10 {
		t.Errorf("Expected StartOffset 10, got %d", opts.StartOffset)
	}
}
