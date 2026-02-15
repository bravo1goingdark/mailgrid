package expression

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple equality",
			input:   `name == "John"`,
			wantErr: false,
		},
		{
			name:    "not equal",
			input:   `name != "John"`,
			wantErr: false,
		},
		{
			name:    "contains operator",
			input:   `email contains "gmail"`,
			wantErr: false,
		},
		{
			name:    "contains function",
			input:   `contains(email, "gmail")`,
			wantErr: false,
		},
		{
			name:    "startsWith operator",
			input:   `name startsWith "Jo"`,
			wantErr: false,
		},
		{
			name:    "startsWith function",
			input:   `startsWith(name, "Jo")`,
			wantErr: false,
		},
		{
			name:    "endsWith operator",
			input:   `email endsWith "@example.com"`,
			wantErr: false,
		},
		{
			name:    "endsWith function",
			input:   `endsWith(email, "@example.com")`,
			wantErr: false,
		},
		{
			name:    "and operator",
			input:   `name == "John" && company == "Acme"`,
			wantErr: false,
		},
		{
			name:    "or operator",
			input:   `name == "John" || name == "Jane"`,
			wantErr: false,
		},
		{
			name:    "not operator",
			input:   `!(name == "John")`,
			wantErr: false,
		},
		{
			name:    "complex expression",
			input:   `(tier == "vip" || tier == "premium") && location == "US"`,
			wantErr: false,
		},
		{
			name:    "greater than",
			input:   `age > 18`,
			wantErr: false,
		},
		{
			name:    "less than",
			input:   `age < 65`,
			wantErr: false,
		},
		{
			name:    "greater or equal",
			input:   `age >= 18`,
			wantErr: false,
		},
		{
			name:    "less or equal",
			input:   `age <= 65`,
			wantErr: false,
		},
		{
			name:    "empty expression",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid syntax",
			input:   `name = = "John"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && expr == nil {
				t.Errorf("Parse(%q) returned nil expression without error", tt.input)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid expression", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParse() panicked on valid expression: %v", r)
			}
		}()

		expr := MustParse(`name == "John"`)
		if expr == nil {
			t.Error("MustParse() returned nil")
		}
	})

	t.Run("invalid expression panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParse() should panic on invalid expression")
			}
		}()

		MustParse("")
	})
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		data     map[string]string
		expected bool
	}{
		{
			name:     "simple equality match",
			input:    `name == "John"`,
			data:     map[string]string{"name": "John"},
			expected: true,
		},
		{
			name:     "simple equality no match",
			input:    `name == "John"`,
			data:     map[string]string{"name": "Jane"},
			expected: false,
		},
		{
			name:     "case insensitive equality",
			input:    `name == "john"`,
			data:     map[string]string{"name": "JOHN"},
			expected: true,
		},
		{
			name:     "not equal match",
			input:    `name != "John"`,
			data:     map[string]string{"name": "Jane"},
			expected: true,
		},
		{
			name:     "contains match",
			input:    `email contains "gmail"`,
			data:     map[string]string{"email": "user@gmail.com"},
			expected: true,
		},
		{
			name:     "contains case insensitive",
			input:    `email contains "GMAIL"`,
			data:     map[string]string{"email": "user@gmail.com"},
			expected: true,
		},
		{
			name:     "contains no match",
			input:    `email contains "yahoo"`,
			data:     map[string]string{"email": "user@gmail.com"},
			expected: false,
		},
		{
			name:     "startsWith match",
			input:    `name startsWith "Jo"`,
			data:     map[string]string{"name": "John"},
			expected: true,
		},
		{
			name:     "startsWith no match",
			input:    `name startsWith "Ja"`,
			data:     map[string]string{"name": "John"},
			expected: false,
		},
		{
			name:     "endsWith match",
			input:    `email endsWith "@example.com"`,
			data:     map[string]string{"email": "user@example.com"},
			expected: true,
		},
		{
			name:     "endsWith no match",
			input:    `email endsWith "@gmail.com"`,
			data:     map[string]string{"email": "user@example.com"},
			expected: false,
		},
		{
			name:     "and operator both true",
			input:    `name == "John" && company == "Acme"`,
			data:     map[string]string{"name": "John", "company": "Acme"},
			expected: true,
		},
		{
			name:     "and operator one false",
			input:    `name == "John" && company == "Acme"`,
			data:     map[string]string{"name": "John", "company": "Other"},
			expected: false,
		},
		{
			name:     "or operator one true",
			input:    `name == "John" || name == "Jane"`,
			data:     map[string]string{"name": "Jane"},
			expected: true,
		},
		{
			name:     "or operator both false",
			input:    `name == "John" || name == "Jane"`,
			data:     map[string]string{"name": "Bob"},
			expected: false,
		},

		{
			name:     "complex expression",
			input:    `(tier == "vip" || tier == "premium") && location == "us"`,
			data:     map[string]string{"tier": "vip", "location": "US"},
			expected: true,
		},
		{
			name:     "complex expression no match",
			input:    `(tier == "vip" || tier == "premium") && location == "us"`,
			data:     map[string]string{"tier": "basic", "location": "US"},
			expected: false,
		},
		{
			name:     "undefined variable",
			input:    `name == "John"`,
			data:     map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse expression %q: %v", tt.input, err)
			}

			result := expr.Evaluate(tt.data)
			if result != tt.expected {
				t.Errorf("Evaluate() = %v, expected %v for expression %q with data %v", result, tt.expected, tt.input, tt.data)
			}
		})
	}
}

func TestTransformExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "contains function",
			input:    `contains(field, "Value")`,
			expected: `field contains "value"`,
		},
		{
			name:     "startsWith function",
			input:    `startsWith(field, "Value")`,
			expected: `field startsWith "value"`,
		},
		{
			name:     "endsWith function",
			input:    `endsWith(field, "Value")`,
			expected: `field endsWith "value"`,
		},
		{
			name:     "equality comparison",
			input:    `field == "Value"`,
			expected: `field == "value"`,
		},
		{
			name:     "not equal comparison",
			input:    `field != "Value"`,
			expected: `field != "value"`,
		},
		{
			name:     "contains operator",
			input:    `field contains "Value"`,
			expected: `field contains "value"`,
		},
		{
			name:     "multiple transformations",
			input:    `name == "John" && contains(email, "GMAIL")`,
			expected: `name == "john" && email contains "gmail"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformExpression(tt.input)
			if result != tt.expected {
				t.Errorf("transformExpression(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
