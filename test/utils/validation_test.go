package utils_test

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
	"github.com/bravo1goingdark/mailgrid/utils/valid"
	"github.com/stretchr/testify/assert"
)

func TestExtractFieldNames(t *testing.T) {
	tests := []struct {
		name     string
		exprStr  string
		expected []string
	}{
		{
			name:     "simple equality",
			exprStr:  `name == "John"`,
			expected: []string{"name"},
		},
		{
			name:     "multiple fields",
			exprStr:  `name == "John" && company == "Acme"`,
			expected: []string{"name", "company"},
		},
		{
			name:     "contains function",
			exprStr:  `contains(location, "York")`,
			expected: []string{"location"},
		},
		{
			name:     "startsWith function",
			exprStr:  `startsWith(email, "admin")`,
			expected: []string{"email", "admin"},
		},
		{
			name:     "operator style",
			exprStr:  `email contains "test"`,
			expected: []string{"email", "test"},
		},
		{
			name:     "complex expression",
			exprStr:  `(country == "US" && city == "NYC") || status == "active"`,
			expected: []string{"country", "city", "status", "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valid.ExtractFieldNames(tt.exprStr)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestValidateFields(t *testing.T) {
	expr, err := expression.Parse(`name == "John"`)
	assert.NoError(t, err)

	recipients := []parser.Recipient{
		{
			Data: map[string]string{
				"name":    "John",
				"email":   "john@example.com",
				"company": "Acme",
			},
		},
	}

	err = valid.ValidateFields(expr, recipients)
	assert.NoError(t, err)
}

func TestValidateFieldsEmptyRecipients(t *testing.T) {
	expr, err := expression.Parse(`name == "John"`)
	assert.NoError(t, err)

	err = valid.ValidateFields(expr, []parser.Recipient{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients")
}

func TestValidateFieldsWithAllFields(t *testing.T) {
	expr, err := expression.Parse(`name == "John" && age > 18 && city contains "York"`)
	assert.NoError(t, err)

	recipients := []parser.Recipient{
		{
			Data: map[string]string{
				"name":   "John",
				"age":    "25",
				"city":   "New York",
				"email":  "john@example.com",
				"status": "active",
			},
		},
	}

	err = valid.ValidateFields(expr, recipients)
	assert.NoError(t, err)
}
