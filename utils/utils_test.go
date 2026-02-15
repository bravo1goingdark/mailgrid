package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single email",
			input:    "test@example.com",
			expected: []string{"test@example.com"},
		},
		{
			name:     "multiple emails",
			input:    "test1@example.com,test2@example.com,test3@example.com",
			expected: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
		},
		{
			name:     "emails with whitespace",
			input:    "  test1@example.com  ,  test2@example.com  ,  test3@example.com  ",
			expected: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
		},
		{
			name:     "emails with empty entries",
			input:    "test1@example.com,,test2@example.com",
			expected: []string{"test1@example.com", "test2@example.com"},
		},
		{
			name:     "all whitespace entries",
			input:    "  ,  ,  ",
			expected: []string{},
		},
		{
			name:     "mixed content",
			input:    "  , test1@example.com ,  , test2@example.com , ",
			expected: []string{"test1@example.com", "test2@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitAndTrim(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("SplitAndTrim(%q) returned %d items, expected %d", tt.input, len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("SplitAndTrim(%q)[%d] = %q, expected %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSplitAndTrimDoesNotReturnNil(t *testing.T) {
	result := SplitAndTrim("")
	if result == nil {
		t.Error("SplitAndTrim(\"\") should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Errorf("SplitAndTrim(\"\") returned %d items, expected 0", len(result))
	}
}

func TestReadTextInput(t *testing.T) {
	t.Run("inline text", func(t *testing.T) {
		input := "Hello, this is a test message"
		result, err := ReadTextInput(input)
		if err != nil {
			t.Errorf("ReadTextInput(%q) returned error: %v", input, err)
		}
		if result != input {
			t.Errorf("ReadTextInput(%q) = %q, expected %q", input, result, input)
		}
	})

	t.Run("text file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "Hello from file!"

		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadTextInput(testFile)
		if err != nil {
			t.Errorf("ReadTextInput(%q) returned error: %v", testFile, err)
		}
		if result != content {
			t.Errorf("ReadTextInput(%q) = %q, expected %q", testFile, result, content)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		nonExistentFile := "/path/that/does/not/exist.txt"
		_, err := ReadTextInput(nonExistentFile)
		if err == nil {
			t.Error("ReadTextInput() should return error for non-existent file")
		}
	})

	t.Run("file with .txt suffix in name but inline", func(t *testing.T) {
		input := "This is not a file.txt just a string"
		result, err := ReadTextInput(input)
		if err != nil {
			t.Errorf("ReadTextInput(%q) returned error: %v", input, err)
		}
		if result != input {
			t.Errorf("ReadTextInput(%q) = %q, expected %q", input, result, input)
		}
	})
}
