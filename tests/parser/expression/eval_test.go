package expression

import (
	"mailgrid/parser"
	"testing"
)

func TestParseAndFilter(t *testing.T) {
	expr, err := parser.Parse("name == Alice AND score > 5")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	r1 := parser.Recipient{Email: "a@b.com", Data: map[string]string{"name": "Alice", "score": "6"}}
	r2 := parser.Recipient{Email: "b@b.com", Data: map[string]string{"name": "Bob", "score": "7"}}
	out := parser.Filter([]parser.Recipient{r1, r2}, expr)
	if len(out) != 1 || out[0].Email != "a@b.com" {
		t.Fatalf("unexpected filter result: %+v", out)
	}
}

func TestParseOrContains(t *testing.T) {
	expr, err := parser.Parse("email contains example.com OR name == Bob")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	r1 := parser.Recipient{Email: "a@example.com", Data: map[string]string{"name": "Ann"}}
	r2 := parser.Recipient{Email: "b@test.com", Data: map[string]string{"name": "Bob"}}
	out := parser.Filter([]parser.Recipient{r1, r2}, expr)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}
