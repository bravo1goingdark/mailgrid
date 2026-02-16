package parser

import (
	"testing"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name       string
		recipients []Recipient
		exprInput  string
		expected   int
	}{
		{
			name: "filter by email",
			recipients: []Recipient{
				{Email: "test1@example.com", Data: map[string]string{"name": "Test1"}},
				{Email: "test2@example.com", Data: map[string]string{"name": "Test2"}},
				{Email: "test3@example.com", Data: map[string]string{"name": "Test3"}},
			},
			exprInput: `email == "test1@example.com"`,
			expected:  1,
		},
		{
			name: "filter by name",
			recipients: []Recipient{
				{Email: "test1@example.com", Data: map[string]string{"name": "Alice"}},
				{Email: "test2@example.com", Data: map[string]string{"name": "Bob"}},
				{Email: "test3@example.com", Data: map[string]string{"name": "Charlie"}},
			},
			exprInput: `name == "Bob"`,
			expected:  1,
		},
		{
			name: "filter with contains",
			recipients: []Recipient{
				{Email: "test1@gmail.com", Data: map[string]string{"name": "Alice"}},
				{Email: "test2@yahoo.com", Data: map[string]string{"name": "Bob"}},
				{Email: "test3@gmail.com", Data: map[string]string{"name": "Charlie"}},
			},
			exprInput: `email contains "gmail"`,
			expected:  2,
		},
		{
			name: "filter returns all",
			recipients: []Recipient{
				{Email: "test1@example.com", Data: map[string]string{"name": "Alice"}},
				{Email: "test2@example.com", Data: map[string]string{"name": "Bob"}},
				{Email: "test3@example.com", Data: map[string]string{"name": "Charlie"}},
			},
			exprInput: `name != "Nobody"`,
			expected:  3,
		},
		{
			name: "filter returns none",
			recipients: []Recipient{
				{Email: "test1@example.com", Data: map[string]string{"name": "Alice"}},
				{Email: "test2@example.com", Data: map[string]string{"name": "Bob"}},
				{Email: "test3@example.com", Data: map[string]string{"name": "Charlie"}},
			},
			exprInput: `name == "Nobody"`,
			expected:  0,
		},
		{
			name:       "empty recipients",
			recipients: []Recipient{},
			exprInput:  `name == "Alice"`,
			expected:   0,
		},
		{
			name: "case insensitive field matching",
			recipients: []Recipient{
				{Email: "test@example.com", Data: map[string]string{"NAME": "Alice", "Company": "Acme"}},
			},
			exprInput: `name == "Alice"`,
			expected:  1,
		},
		{
			name: "filter with startsWith",
			recipients: []Recipient{
				{Email: "alice@example.com", Data: map[string]string{"name": "Alice"}},
				{Email: "bob@example.com", Data: map[string]string{"name": "Bob"}},
				{Email: "alice@company.com", Data: map[string]string{"name": "Alice2"}},
			},
			exprInput: `email startsWith "alice"`,
			expected:  2,
		},
		{
			name: "filter with endsWith",
			recipients: []Recipient{
				{Email: "user1@example.com", Data: map[string]string{"name": "User1"}},
				{Email: "user2@example.com", Data: map[string]string{"name": "User2"}},
				{Email: "user3@company.com", Data: map[string]string{"name": "User3"}},
			},
			exprInput: `email endsWith "example.com"`,
			expected:  2,
		},
		{
			name: "complex logical expression",
			recipients: []Recipient{
				{Email: "vip1@example.com", Data: map[string]string{"tier": "vip", "location": "US"}},
				{Email: "vip2@example.com", Data: map[string]string{"tier": "vip", "location": "UK"}},
				{Email: "basic@example.com", Data: map[string]string{"tier": "basic", "location": "US"}},
				{Email: "premium@example.com", Data: map[string]string{"tier": "premium", "location": "US"}},
			},
			exprInput: `(tier == "vip" || tier == "premium") && location == "us"`,
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseExpression(tt.exprInput)
			if err != nil {
				t.Fatalf("Failed to parse expression %q: %v", tt.exprInput, err)
			}

			result := Filter(tt.recipients, expr)
			if len(result) != tt.expected {
				t.Errorf("Filter() returned %d recipients, expected %d", len(result), tt.expected)
			}
		})
	}
}

func TestFilterPreservesEmail(t *testing.T) {
	recipients := []Recipient{
		{Email: "test@example.com", Data: map[string]string{"name": "Test"}},
	}

	expr, err := ParseExpression(`email == "test@example.com"`)
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	result := Filter(recipients, expr)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result[0].Email != "test@example.com" {
		t.Errorf("Email not preserved: got %s, expected test@example.com", result[0].Email)
	}

	if result[0].Data["name"] != "Test" {
		t.Errorf("Data not preserved: got %s, expected Test", result[0].Data["name"])
	}
}

func TestFilterCaseInsensitivity(t *testing.T) {
	recipients := []Recipient{
		{Email: "TEST@EXAMPLE.COM", Data: map[string]string{"name": "Test"}},
	}

	expr, err := ParseExpression(`email == "test@example.com"`)
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	result := Filter(recipients, expr)

	if len(result) != 1 {
		t.Errorf("Expected 1 result for case-insensitive match, got %d", len(result))
	}
}
