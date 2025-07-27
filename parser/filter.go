// Package parser provides CSV recipient filtering using logical expressions.
//
// This file implements the high-level Filter() function that:
// - Applies a logical expression to each recipient
// - Normalizes field names to lowercase
// - Supports email fallback field
package parser

import (
	"strings"

	"mailgrid/parser/expression"
)

// Filter applies the provided logical EXPRESSION to a slice of recipients.
// It returns only those recipients for whom the expression evaluates to true.
//
// The expression is evaluated case-insensitively, and field names are normalized to lowercase.
func Filter(recipients []Recipient, exp expression.Expression) []Recipient {
	var out []Recipient

	for _, r := range recipients {
		// Normalize field names for filtering
		data := make(map[string]string)
		data["email"] = r.Email
		for k, v := range r.Data {
			data[strings.ToLower(k)] = v
		}

		if exp.Evaluate(data) {
			out = append(out, r)
		}
	}

	return out
}
