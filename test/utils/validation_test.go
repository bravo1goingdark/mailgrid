package valid

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
	"github.com/stretchr/testify/assert"
)

func TestValidateFields_EmptyRecipients(t *testing.T) {
	expr := expression.Condition{Field: "name", Op: "==", Value: "John"}
	err := ValidateFields(expr, []parser.Recipient{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients to validate fields")
}

func TestValidateFields_ValidField(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John", "age": "30"},
		},
	}

	tests := []struct {
		name string
		expr expression.Expression
	}{
		{
			name: "existing field",
			expr: expression.Condition{Field: "name", Op: "==", Value: "John"},
		},
		{
			name: "email field",
			expr: expression.Condition{Field: "email", Op: "==", Value: "user@example.com"},
		},
		{
			name: "case insensitive field",
			expr: expression.Condition{Field: "NAME", Op: "==", Value: "John"},
		},
		{
			name: "case insensitive field lowercase",
			expr: expression.Condition{Field: "age", Op: "==", Value: "30"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFields(tt.expr, recipients)
			assert.NoError(t, err)
		})
	}
}

func TestValidateFields_InvalidField(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John", "age": "30"},
		},
	}

	expr := expression.Condition{Field: "nonexistent", Op: "==", Value: "value"}
	err := ValidateFields(expr, recipients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `field "nonexistent" not found in CSV header`)
}

func TestValidateFields_AndExpression(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John", "age": "30"},
		},
	}

	// Valid AND expression
	expr := expression.And{
		Left:  expression.Condition{Field: "name", Op: "==", Value: "John"},
		Right: expression.Condition{Field: "age", Op: "==", Value: "30"},
	}
	err := ValidateFields(expr, recipients)
	assert.NoError(t, err)

	// AND expression with invalid field on left
	expr = expression.And{
		Left:  expression.Condition{Field: "invalid", Op: "==", Value: "John"},
		Right: expression.Condition{Field: "age", Op: "==", Value: "30"},
	}
	err = ValidateFields(expr, recipients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `field "invalid" not found`)

	// AND expression with invalid field on right
	expr = expression.And{
		Left:  expression.Condition{Field: "name", Op: "==", Value: "John"},
		Right: expression.Condition{Field: "invalid", Op: "==", Value: "30"},
	}
	err = ValidateFields(expr, recipients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `field "invalid" not found`)
}

func TestValidateFields_OrExpression(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John", "age": "30"},
		},
	}

	// Valid OR expression
	expr := expression.Or{
		Left:  expression.Condition{Field: "name", Op: "==", Value: "John"},
		Right: expression.Condition{Field: "age", Op: "==", Value: "30"},
	}
	err := ValidateFields(expr, recipients)
	assert.NoError(t, err)

	// OR expression with invalid field
	expr = expression.Or{
		Left:  expression.Condition{Field: "invalid", Op: "==", Value: "John"},
		Right: expression.Condition{Field: "age", Op: "==", Value: "30"},
	}
	err = ValidateFields(expr, recipients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `field "invalid" not found`)
}

func TestValidateFields_NotExpression(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John", "age": "30"},
		},
	}

	// Valid NOT expression
	expr := expression.Not{
		Inner: expression.Condition{Field: "name", Op: "==", Value: "John"},
	}
	err := ValidateFields(expr, recipients)
	assert.NoError(t, err)

	// NOT expression with invalid field
	expr = expression.Not{
		Inner: expression.Condition{Field: "invalid", Op: "==", Value: "John"},
	}
	err = ValidateFields(expr, recipients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `field "invalid" not found`)
}

func TestValidateFields_ComplexExpression(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John", "age": "30", "city": "NYC"},
		},
	}

	// Complex valid expression: (name == "John" AND age == "30") OR NOT (city == "LA")
	expr := expression.Or{
		Left: expression.And{
			Left:  expression.Condition{Field: "name", Op: "==", Value: "John"},
			Right: expression.Condition{Field: "age", Op: "==", Value: "30"},
		},
		Right: expression.Not{
			Inner: expression.Condition{Field: "city", Op: "==", Value: "LA"},
		},
	}
	err := ValidateFields(expr, recipients)
	assert.NoError(t, err)

	// Complex invalid expression with one invalid field
	expr = expression.Or{
		Left: expression.And{
			Left:  expression.Condition{Field: "name", Op: "==", Value: "John"},
			Right: expression.Condition{Field: "invalid_field", Op: "==", Value: "30"},
		},
		Right: expression.Not{
			Inner: expression.Condition{Field: "city", Op: "==", Value: "LA"},
		},
	}
	err = ValidateFields(expr, recipients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `field "invalid_field" not found`)
}

func TestValidateFields_CaseInsensitivity(t *testing.T) {
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"Name": "John", "AGE": "30", "city": "NYC"},
		},
	}

	tests := []struct {
		name  string
		field string
	}{
		{"lowercase field name", "name"},
		{"uppercase field name", "NAME"},
		{"mixed case field name", "Name"},
		{"existing lowercase", "city"},
		{"existing uppercase lowercase", "age"},
		{"existing uppercase", "AGE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := expression.Condition{Field: tt.field, Op: "==", Value: "test"}
			err := ValidateFields(expr, recipients)
			assert.NoError(t, err)
		})
	}
}

func TestValidateFields_EmailFieldAlwaysValid(t *testing.T) {
	// Test that 'email' field is always valid even if not in CSV data
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data:  map[string]string{"name": "John"}, // No 'email' in data
		},
	}

	expr := expression.Condition{Field: "email", Op: "==", Value: "user@example.com"}
	err := ValidateFields(expr, recipients)
	assert.NoError(t, err)
}