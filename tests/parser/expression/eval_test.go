// tests/parser/expression/eval_test.go
package expression_test

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser/expression"
)

func TestCondition_Evaluate(t *testing.T) {
	data := map[string]string{
		"name":     "Ashutosh",
		"company":  "OpenAI",
		"salary":   "50000",
		"location": "India",
	}

	tests := []struct {
		name     string
		expr     expression.Expression
		expected bool
	}{
		{
			name:     "Equal comparison",
			expr:     expression.Condition{Field: "name", Op: "=", Value: "Ashutosh"},
			expected: true,
		},
		{
			name:     "Not Equal comparison",
			expr:     expression.Condition{Field: "company", Op: "!=", Value: "Google"},
			expected: true,
		},
		{
			name:     "Contains comparison",
			expr:     expression.Condition{Field: "company", Op: "contains", Value: "open"},
			expected: true,
		},
		{
			name:     "StartsWith comparison",
			expr:     expression.Condition{Field: "location", Op: "startsWith", Value: "Ind"},
			expected: true,
		},
		{
			name:     "EndsWith comparison",
			expr:     expression.Condition{Field: "location", Op: "endsWith", Value: "dia"},
			expected: true,
		},
		{
			name:     "Numeric greater than",
			expr:     expression.Condition{Field: "salary", Op: ">", Value: "40000"},
			expected: true,
		},
		{
			name:     "Numeric less than",
			expr:     expression.Condition{Field: "salary", Op: "<", Value: "60000"},
			expected: true,
		},
		{
			name: "AND condition",
			expr: expression.And{
				Left:  expression.Condition{Field: "company", Op: "=", Value: "OpenAI"},
				Right: expression.Condition{Field: "location", Op: "=", Value: "India"},
			},
			expected: true,
		},
		{
			name: "OR condition",
			expr: expression.Or{
				Left:  expression.Condition{Field: "name", Op: "=", Value: "Elon"},
				Right: expression.Condition{Field: "company", Op: "=", Value: "OpenAI"},
			},
			expected: true,
		},
		{
			name: "NOT condition",
			expr: expression.Not{
				Inner: expression.Condition{Field: "location", Op: "=", Value: "USA"},
			},
			expected: true,
		},
		{
			name:     "Missing field",
			expr:     expression.Condition{Field: "department", Op: "=", Value: "Engineering"},
			expected: false,
		},
		{
			name:     "Invalid numeric comparison",
			expr:     expression.Condition{Field: "name", Op: ">", Value: "1000"},
			expected: false,
		},
		{
			name:     "Case insensitive contains",
			expr:     expression.Condition{Field: "company", Op: "contains", Value: "OPEN"},
			expected: true,
		},
		{
			name:     "Numeric equality",
			expr:     expression.Condition{Field: "salary", Op: "=", Value: "50000"},
			expected: true,
		},
		{
			name:     "StartsWith fails",
			expr:     expression.Condition{Field: "location", Op: "startsWith", Value: "USA"},
			expected: false,
		},
		{
			name:     "Alphanumeric mismatch",
			expr:     expression.Condition{Field: "salary", Op: "contains", Value: "USD"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expr.Evaluate(data)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
