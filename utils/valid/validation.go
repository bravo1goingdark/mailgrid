package valid

import (
	"fmt"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
	"strings"
)

// ValidateFields ensures all fields used in the expression exist in the CSV data
func ValidateFields(expr expression.Expression, recipients []parser.Recipient) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients to validate fields")
	}

	// Use first recipient's fields as schema
	validFields := map[string]struct{}{
		"email": {}, // email is always allowed
	}
	for k := range recipients[0].Data {
		validFields[strings.ToLower(k)] = struct{}{}
	}

	// Recursively visit all Condition nodes
	var check func(expression.Expression) error
	check = func(e expression.Expression) error {
		switch v := e.(type) {
		case expression.Condition:
			field := strings.ToLower(v.Field)
			if _, ok := validFields[field]; !ok {
				return fmt.Errorf("field %q not found in CSV header", v.Field)
			}
		case expression.And:
			if err := check(v.Left); err != nil {
				return err
			}
			if err := check(v.Right); err != nil {
				return err
			}
		case expression.Or:
			if err := check(v.Left); err != nil {
				return err
			}
			if err := check(v.Right); err != nil {
				return err
			}
		case expression.Not:
			if err := check(v.Inner); err != nil {
				return err
			}
		}
		return nil
	}

	return check(expr)
}
