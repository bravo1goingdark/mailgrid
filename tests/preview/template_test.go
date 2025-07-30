package preview

import (
	"os"
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/utils/preview"
)

func TestRenderTemplate(t *testing.T) {
	tmpl := "<p>Hello {{ .Data.name }}</p>"
	tmp, err := os.CreateTemp(t.TempDir(), "tmpl*.html")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString(tmpl); err != nil {
		t.Fatal(err)
	}
	err = tmp.Close()
	if err != nil {
		return
	}

	rec := parser.Recipient{Email: "a@b.com", Data: map[string]string{"name": "Alice"}}

	out, err := preview.RenderTemplate(rec, tmp.Name())
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}
	if out != "<p>Hello Alice</p>" {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestRenderTemplate_FileMissing(t *testing.T) {
	rec := parser.Recipient{}
	if _, err := preview.RenderTemplate(rec, "nope.html"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
