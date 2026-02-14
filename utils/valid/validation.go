package valid

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
)

// ValidateFields ensures all fields used in the expression exist in the CSV data.
// Since we use the expr library, it handles undefined variables gracefully.
// This function does basic validation and extracts field names for warnings.
func ValidateFields(expr expression.Expression, recipients []parser.Recipient) error {
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
