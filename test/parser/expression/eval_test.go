package parser_test

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
)

func TestExpression_Evaluate(t *testing.T) {
	data := map[string]string{
		"name":     "Ashutosh",
		"company":  "OpenAI",
		"salary":   "50000",
		"location": "India",
	}

	tests := []struct {
		name       string
		expr       string
		data       map[string]string
		expected   bool
		shouldFail bool
	}{
		{
			name:     "Equal comparison",
			expr:     `name == "Ashutosh"`,
			data:     data,
			expected: true,
		},
		{
			name:     "Not Equal comparison",
			expr:     `company != "Google"`,
			data:     data,
			expected: true,
		},
		{
			name:     "Contains comparison",
			expr:     `contains(company, "open")`,
			data:     data,
			expected: true,
		},
		{
			name:     "StartsWith comparison",
			expr:     `startsWith(location, "Ind")`,
			data:     data,
			expected: true,
		},
		{
			name:     "EndsWith comparison",
			expr:     `endsWith(location, "dia")`,
			data:     data,
			expected: true,
		},
		{
			name:     "Numeric greater than",
			expr:     `salary > "40000"`,
			data:     data,
			expected: true,
		},
		{
			name:     "Numeric less than",
			expr:     `salary < "60000"`,
			data:     data,
			expected: true,
		},
		{
			name:     "AND condition",
			expr:     `company == "OpenAI" && location == "India"`,
			data:     data,
			expected: true,
		},
		{
			name:     "OR condition",
			expr:     `name == "Elon" || company == "OpenAI"`,
			data:     data,
			expected: true,
		},
		{
			name:     "NOT condition",
			expr:     `!(location == "USA")`,
			data:     data,
			expected: true,
		},
		{
			name:     "Missing field",
			expr:     `department == "Engineering"`,
			data:     data,
			expected: false,
		},
		{
			name:     "Case insensitive contains",
			expr:     `contains(company, "OPEN")`,
			data:     data,
			expected: true,
		},
		{
			name:     "Numeric equality",
			expr:     `salary == "50000"`,
			data:     data,
			expected: true,
		},
		{
			name:     "StartsWith fails",
			expr:     `startsWith(location, "USA")`,
			data:     data,
			expected: false,
		},
		{
			name:     "Alphanumeric mismatch",
			expr:     `contains(salary, "USD")`,
			data:     data,
			expected: false,
		},
		{
			name:     "Complex nested condition",
			expr:     `(company == "OpenAI" && salary > "40000") || name == "Elon"`,
			data:     data,
			expected: true,
		},
		{
			name:       "Invalid syntax",
			expr:       `name ==`,
			data:       data,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.expr)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("expected error for invalid expression %q", tt.expr)
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to parse expression %q: %v", tt.expr, err)
			}

			result := expr.Evaluate(tt.data)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for expression %q", tt.expected, result, tt.expr)
			}
		})
	}
}

func TestParse_InvalidExpressions(t *testing.T) {
	invalidExprs := []string{
		"",
		"   ",
		"name ==",
		"== value",
	}

	for _, expr := range invalidExprs {
		_, err := parser.ParseExpression(expr)
		if err == nil {
			t.Errorf("expected error for invalid expression %q", expr)
		}
	}
}

func TestExpression_OperatorStyle(t *testing.T) {
	data := map[string]string{
		"email":  "test@example.com",
		"name":   "John Doe",
		"status": "active",
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{
			name:     "operator style contains",
			expr:     `email contains "example"`,
			expected: true,
		},
		{
			name:     "operator style startsWith",
			expr:     `name startsWith "John"`,
			expected: true,
		},
		{
			name:     "operator style endsWith",
			expr:     `email endsWith "com"`,
			expected: true,
		},
		{
			name:     "operator style equals",
			expr:     `status == "active"`,
			expected: true,
		},
		{
			name:     "operator style not equals",
			expr:     `status != "inactive"`,
			expected: true,
		},
		{
			name:     "case insensitive value",
			expr:     `status == "ACTIVE"`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.expr)
			if err != nil {
				t.Fatalf("failed to parse expression %q: %v", tt.expr, err)
			}

			result := expr.Evaluate(data)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for expression %q", tt.expected, result, tt.expr)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	expr := parser.MustParseExpression(`name == "test"`)
	if !expr.Evaluate(map[string]string{"name": "test"}) {
		t.Error("MustParseExpression should not panic for valid expression")
	}
}
