package valid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAddressInput_Empty(t *testing.T) {
	result, err := ParseAddressInput("")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestParseAddressInput_InlineMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single email",
			input:    "user@example.com",
			expected: []string{"user@example.com"},
		},
		{
			name:     "multiple emails",
			input:    "user1@example.com,user2@example.com,user3@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:     "emails with spaces",
			input:    "user1@example.com, user2@example.com , user3@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:     "emails with empty parts",
			input:    "user1@example.com,,user2@example.com, ,user3@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:     "single email with whitespace",
			input:    "  user@example.com  ",
			expected: []string{"user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAddressInput(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAddressInput_FileMode(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		fileContent string
		expected    []string
	}{
		{
			name:        "single email in file",
			fileContent: "user@example.com",
			expected:    []string{"user@example.com"},
		},
		{
			name: "multiple emails in file",
			fileContent: `user1@example.com
user2@example.com
user3@example.com`,
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name: "emails with empty lines",
			fileContent: `user1@example.com

user2@example.com

user3@example.com
`,
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name: "emails with whitespace",
			fileContent: `  user1@example.com
  user2@example.com
  user3@example.com  `,
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:        "empty file",
			fileContent: "",
			expected:    nil, // Function returns nil for empty
		},
		{
			name:        "file with only whitespace",
			fileContent: "   \n  \n   ",
			expected:    nil, // Function returns nil for empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, "test_addresses.txt")
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			require.NoError(t, err)

			result, err := ParseAddressInput(filePath)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAddressInput_FileNotFound(t *testing.T) {
	result, err := ParseAddressInput("/nonexistent/file.txt")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to read address file")
}

func TestParseAddressInput_InlineModeDetection(t *testing.T) {
	// Test that the function correctly detects inline mode vs file mode

	// Should be detected as inline mode (contains @)
	result, err := ParseAddressInput("test@example.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"test@example.com"}, result)

	// Should be detected as file mode (no @)
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "addresses")
	err = os.WriteFile(filePath, []byte("user1@example.com\nuser2@example.com"), 0644)
	require.NoError(t, err)

	result, err = ParseAddressInput(filePath)
	require.NoError(t, err)
	assert.Equal(t, []string{"user1@example.com", "user2@example.com"}, result)
}

func TestParseAddressInput_EdgeCases(t *testing.T) {
	// Test with just commas
	result, err := ParseAddressInput("test@example.com,,,")
	require.NoError(t, err)
	assert.Equal(t, []string{"test@example.com"}, result)

	// Test with just spaces and commas
	result, err = ParseAddressInput("test@example.com, , , ")
	require.NoError(t, err)
	assert.Equal(t, []string{"test@example.com"}, result)
}