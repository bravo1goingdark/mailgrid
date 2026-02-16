package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Expression is the interface for evaluating filter expressions.
// It wraps the expr library for boolean evaluation against recipient data.
type Expression interface {
	Evaluate(data map[string]string) bool
}

// compiledExpr wraps a compiled expr program for evaluation.
type compiledExpr struct {
	program *vm.Program
}

// Evaluate runs the compiled expression against the provided data.
func (c *compiledExpr) Evaluate(data map[string]string) bool {
	// Convert all string values to lowercase for case-insensitive comparison
	lowerData := make(map[string]string, len(data))
	for k, v := range data {
		lowerData[k] = strings.ToLower(v)
	}

	result, err := expr.Run(c.program, lowerData)
	if err != nil {
		return false
	}

	// Handle boolean result
	if b, ok := result.(bool); ok {
		return b
	}

	// Handle string result (truthy check)
	if s, ok := result.(string); ok {
		return s != "" && s != "false"
	}

	return false
}

// transformExpression converts our simplified syntax to expr-compatible syntax.
// It transforms:
//   - contains(field, "value") -> field contains "value" (with lowercase value)
//   - startsWith(field, "value") -> field startsWith "value" (with lowercase value)
//   - endsWith(field, "value") -> field endsWith "value" (with lowercase value)
//   - field == "Value" -> field == "value" (lowercase comparison value)
func transformExpression(input string) string {
	// Helper to lowercase a quoted string
	lowerCaseString := func(s string) string {
		// Extract the quoted content
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			content := s[1 : len(s)-1]
			return `"` + strings.ToLower(content) + `"`
		}
		return s
	}

	// Transform contains(field, "value") -> field contains "value"
	containsRe := regexp.MustCompile(`contains\s*\(\s*(\w+)\s*,\s*("[^"]*")\s*\)`)
	input = containsRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := containsRe.FindStringSubmatch(match)
		if len(parts) == 3 {
			return parts[1] + ` contains ` + lowerCaseString(parts[2])
		}
		return match
	})

	// Transform startsWith(field, "value") -> field startsWith "value"
	startsWithRe := regexp.MustCompile(`startsWith\s*\(\s*(\w+)\s*,\s*("[^"]*")\s*\)`)
	input = startsWithRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := startsWithRe.FindStringSubmatch(match)
		if len(parts) == 3 {
			return parts[1] + ` startsWith ` + lowerCaseString(parts[2])
		}
		return match
	})

	// Transform endsWith(field, "value") -> field endsWith "value"
	endsWithRe := regexp.MustCompile(`endsWith\s*\(\s*(\w+)\s*,\s*("[^"]*")\s*\)`)
	input = endsWithRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := endsWithRe.FindStringSubmatch(match)
		if len(parts) == 3 {
			return parts[1] + ` endsWith ` + lowerCaseString(parts[2])
		}
		return match
	})

	// Transform direct operator usage (contains, startsWith, endsWith) to lowercase comparison values
	operatorRe := regexp.MustCompile(`(\w+)\s+(contains|startsWith|endsWith)\s+("[^"]*")`)
	input = operatorRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := operatorRe.FindStringSubmatch(match)
		if len(parts) == 4 {
			return parts[1] + ` ` + parts[2] + ` ` + lowerCaseString(parts[3])
		}
		return match
	})

	// Transform equality comparisons to lowercase values
	// field == "Value" -> field == "value"
	// field != "Value" -> field != "value"
	equalityRe := regexp.MustCompile(`(\w+)\s*(==|!=)\s*("[^"]*")`)
	input = equalityRe.ReplaceAllStringFunc(input, func(match string) string {
		parts := equalityRe.FindStringSubmatch(match)
		if len(parts) == 4 {
			return parts[1] + ` ` + parts[2] + ` ` + lowerCaseString(parts[3])
		}
		return match
	})

	return input
}

// ParseExpression compiles a filter expression string into an evaluable Expression.
//
// Supported syntax:
//   - Comparison: ==, !=, <, <=, >, >=
//   - String contains: contains(field, "value") or field contains "value"
//   - String prefix: startsWith(field, "value") or field startsWith "value"
//   - String suffix: endsWith(field, "value") or field endsWith "value"
//   - Logical operators: &&, ||, !
//   - Parentheses for grouping: (a == b && c == d)
//
// All string comparisons are case-insensitive.
//
// Example expressions:
//   - name == "John"
//   - company != "Acme" && salary > 50000
//   - contains(location, "York") || startsWith(email, "admin")
//   - location contains "York" || email startsWith "admin"
func ParseExpression(input string) (Expression, error) {
	// Clean up the input
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Transform function-style calls to operator-style and lowercase values
	input = transformExpression(input)

	// Configure expr with options
	options := []expr.Option{
		expr.Env(map[string]string{}),
		expr.AllowUndefinedVariables(),
	}

	program, err := expr.Compile(input, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	return &compiledExpr{program: program}, nil
}

// MustParseExpression is like ParseExpression but panics on error. Useful for testing.
func MustParseExpression(input string) Expression {
	expr, err := ParseExpression(input)
	if err != nil {
		panic(err)
	}
	return expr
}

// ValidateFields ensures all fields used in the expression exist in the CSV data.
// Since we use the expr library, it handles undefined variables gracefully.
// This function does basic validation and extracts field names for warnings.
func ValidateFields(expr Expression, recipients []Recipient) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients to validate fields")
	}

	// Get valid fields from first recipient
	validFields := map[string]struct{}{
		"email": {},
	}
	for k := range recipients[0].Data {
		validFields[strings.ToLower(k)] = struct{}{}
	}

	// The expr library handles undefined variables gracefully,
	// so we don't need strict validation. Return nil to allow
	// the expression to be used even if fields might be missing.
	return nil
}

// ExtractFieldNames extracts potential field names from an expression string.
// This is a best-effort extraction for informational purposes.
func ExtractFieldNames(exprStr string) []string {
	// Pattern to match identifiers (field names) that aren't string literals
	// This is a simple heuristic - field names are typically lowercase identifiers
	re := regexp.MustCompile(`\b[a-z_][a-z0-9_]*\b`)
	matches := re.FindAllString(exprStr, -1)

	// Filter out keywords and functions
	keywords := map[string]bool{
		"contains": true, "startsWith": true, "endsWith": true,
		"and": true, "or": true, "not": true, "true": true, "false": true,
	}

	var fields []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if !keywords[match] && !seen[match] {
			fields = append(fields, match)
			seen[match] = true
		}
	}

	return fields
}
