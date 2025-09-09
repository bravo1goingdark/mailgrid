package utils

import (
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/parser/expression"
	"github.com/bravo1goingdark/mailgrid/utils/valid"
	"testing"
)

func TestValidateFields(t *testing.T) {
	// Mock recipients
	recipients := []parser.Recipient{
		{
			Email: "user@example.com",
			Data: map[string]string{
				"Name":    "Alice",
				"Company": "Gadgetry",
				"Age":     "30",
			},
		},
	}

	tests := []struct {
		name      string
		expr      expression.Expression
		shouldErr bool
	}{
		{
			name: "valid single field",
			expr: expression.Condition{
				Field: "Name",
				Op:    "contains",
				Value: "Al",
			},
			shouldErr: false,
		},
		{
			name: "valid and condition",
			expr: expression.And{
				Left: expression.Condition{
					Field: "Company",
					Op:    "=",
					Value: "Gadgetry",
				},
				Right: expression.Condition{
					Field: "Age",
					Op:    ">",
					Value: "25",
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid field",
			expr: expression.Condition{
				Field: "UnknownField",
				Op:    "=",
				Value: "X",
			},
			shouldErr: true,
		},
		{
			name: "case-insensitive match",
			expr: expression.Condition{
				Field: "name",
				Op:    "contains",
				Value: "ali",
			},
			shouldErr: false,
		},
		{
			name: "nested NOT and OR with valid/invalid fields",
			expr: expression.Not{
				Inner: expression.Or{
					Left: expression.Condition{
						Field: "Company",
						Op:    "=",
						Value: "X",
					},
					Right: expression.Condition{
						Field: "badfield",
						Op:    "=",
						Value: "Y",
					},
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := valid.ValidateFields(tt.expr, recipients)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
