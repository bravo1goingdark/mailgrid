package parser_test

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
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
			name:     "complex expression",
			exprStr:  `name == "John" && age > 18 && city contains "York"`,
			expected: []string{"name", "age", "city"},
		},
		{
			name:     "no fields (string only)",
			exprStr:  `"hello"`,
			expected: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractFieldNames(tt.exprStr)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestValidateFields(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "test@example.com",
			Data: map[string]string{
				"name": "John",
				"age":  "25",
			},
		},
	}

	tests := []struct {
		name      string
		exprInput string
		wantErr   bool
	}{
		{
			name:      "valid field",
			exprInput: `name == "John"`,
			wantErr:   false,
		},
		{
			name:      "another valid field",
			exprInput: `age > "18"`,
			wantErr:   false,
		},
		{
			name:      "field not in csv but allowed",
			exprInput: `department == "Engineering"`,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.exprInput)
			assert.NoError(t, err)

			err = parser.ValidateFields(expr, recipients)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFields_EmptyRecipients(t *testing.T) {
	expr, err := parser.ParseExpression(`name == "John"`)
	assert.NoError(t, err)

	err = parser.ValidateFields(expr, []parser.Recipient{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients")
}

func TestValidateFields_MultipleFields(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "test@example.com",
			Data: map[string]string{
				"name":    "John",
				"company": "Acme",
				"city":    "NYC",
			},
		},
	}

	expr, err := parser.ParseExpression(`name == "John" && age > 18 && city contains "York"`)
	assert.NoError(t, err)

	// Should not error even with 'age' field not present (expr handles undefined)
	err = parser.ValidateFields(expr, recipients)
	assert.NoError(t, err)
}
