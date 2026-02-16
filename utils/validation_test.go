package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
)

func TestParseAddressInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "single email",
			input:    "test@example.com",
			expected: []string{"test@example.com"},
			wantErr:  false,
		},
		{
			name:     "multiple emails",
			input:    "test1@example.com,test2@example.com,test3@example.com",
			expected: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
			wantErr:  false,
		},
		{
			name:     "emails with whitespace",
			input:    "  test1@example.com  ,  test2@example.com  ",
			expected: []string{"test1@example.com", "test2@example.com"},
			wantErr:  false,
		},
		{
			name:     "emails with empty entries",
			input:    "test1@example.com,,test2@example.com",
			expected: []string{"test1@example.com", "test2@example.com"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAddressInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddressInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expected == nil {
				if result != nil {
					t.Errorf("ParseAddressInput() = %v, expected nil", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ParseAddressInput() returned %d items, expected %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ParseAddressInput()[%d] = %q, expected %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseAddressInputFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "emails.txt")

	content := "user1@example.com\nuser2@example.com\n\nuser3@example.com\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ParseAddressInput(testFile)
	if err != nil {
		t.Errorf("ParseAddressInput() from file returned error: %v", err)
	}

	expected := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	if len(result) != len(expected) {
		t.Errorf("ParseAddressInput() from file returned %d items, expected %d", len(result), len(expected))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("ParseAddressInput() from file [%d] = %q, expected %q", i, result[i], expected[i])
		}
	}
}

func TestParseAddressInputFromNonExistentFile(t *testing.T) {
	nonExistentFile := "/path/that/does/not/exist.txt"
	_, err := ParseAddressInput(nonExistentFile)
	if err == nil {
		t.Error("ParseAddressInput() should return error for non-existent file")
	}
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name       string
		recipients []parser.Recipient
		wantErr    bool
	}{
		{
			name:       "empty recipients",
			recipients: []parser.Recipient{},
			wantErr:    true,
		},
		{
			name: "valid recipients",
			recipients: []parser.Recipient{
				{Email: "test@example.com", Data: map[string]string{"name": "Test", "company": "Acme"}},
			},
			wantErr: false,
		},
		{
			name: "multiple recipients",
			recipients: []parser.Recipient{
				{Email: "test1@example.com", Data: map[string]string{"name": "Test1"}},
				{Email: "test2@example.com", Data: map[string]string{"name": "Test2"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple expression
			expr, _ := parser.ParseExpression(`name == "Test"`)

			err := parser.ValidateFields(expr, tt.recipients)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractFieldNames(t *testing.T) {
	tests := []struct {
		name             string
		exprStr          string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "single field",
			exprStr:          `name == "John"`,
			shouldContain:    []string{"name"},
			shouldNotContain: []string{},
		},
		{
			name:             "multiple fields",
			exprStr:          `name == "John" && company == "Acme"`,
			shouldContain:    []string{"name", "company"},
			shouldNotContain: []string{},
		},
		{
			name:             "field with contains",
			exprStr:          `email contains "gmail"`,
			shouldContain:    []string{"email"},
			shouldNotContain: []string{},
		},
		{
			name:             "field with startsWith",
			exprStr:          `name startsWith "Jo"`,
			shouldContain:    []string{"name"},
			shouldNotContain: []string{},
		},
		{
			name:             "field with endsWith",
			exprStr:          `email endsWith "@example.com"`,
			shouldContain:    []string{"email"},
			shouldNotContain: []string{},
		},
		{
			name:             "complex expression",
			exprStr:          `(tier == "vip" || tier == "premium") && location == "US"`,
			shouldContain:    []string{"tier", "location"},
			shouldNotContain: []string{},
		},
		{
			name:             "no fields",
			exprStr:          `"test"`,
			shouldContain:    []string{},
			shouldNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractFieldNames(tt.exprStr)

			// Check that all expected fields are present
			for _, field := range tt.shouldContain {
				found := false
				for _, r := range result {
					if r == field {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ExtractFieldNames(%q) should contain %q, got %v", tt.exprStr, field, result)
				}
			}

			// Check that unwanted fields are not present
			for _, field := range tt.shouldNotContain {
				for _, r := range result {
					if r == field {
						t.Errorf("ExtractFieldNames(%q) should not contain %q", tt.exprStr, field)
					}
				}
			}
		})
	}
}

func TestExtractFieldNamesFiltersKeywords(t *testing.T) {
	keywords := []string{"contains", "startsWith", "endsWith", "and", "or", "not", "true", "false"}

	for _, keyword := range keywords {
		result := parser.ExtractFieldNames(keyword)
		for _, r := range result {
			if r == keyword {
				t.Errorf("ExtractFieldNames(%q) should filter out keyword %q", keyword, keyword)
			}
		}
	}
}
