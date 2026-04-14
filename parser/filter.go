// Package parser provides CSV recipient filtering using logical expressions.
//
// This file implements the high-level Filter() function that:
// - Applies a logical expression to each recipient
// - Normalizes field names to lowercase
// - Supports email fallback field
package parser

import "strings"

// Filter applies the provided logical EXPRESSION to a slice of recipients.
// It returns only those recipients for whom the expression evaluates to true.
//
// The expression is evaluated case-insensitively, and field names are normalized to lowercase.
func Filter(recipients []Recipient, exp Expression) []Recipient {
	if len(recipients) == 0 {
		return nil
	}

	var out []Recipient

	// Preallocate once with capacity hint; clear and reuse for each recipient
	// to avoid allocating a new map per call.
	data := make(map[string]string, len(recipients[0].Data)+1)

	for _, r := range recipients {
		// Clear previous iteration's data without re-allocating.
		for k := range data {
			delete(data, k)
		}

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
