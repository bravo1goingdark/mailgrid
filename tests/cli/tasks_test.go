package cli_test

import (
	"os"
	"testing"

	"mailgrid/cli"
	"mailgrid/parser"
)

func TestHasMissingFields(t *testing.T) {
	r1 := parser.Recipient{Email: "a@b.com", Data: map[string]string{"name": "Alice"}}
	if cli.HasMissingFields(r1) {
		t.Errorf("expected no missing fields")
	}
	r2 := parser.Recipient{Email: "b@b.com", Data: map[string]string{"name": ""}}
	if !cli.HasMissingFields(r2) {
		t.Errorf("expected missing fields")
	}
}

func TestPrepareEmailTasks(t *testing.T) {
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

	recipients := []parser.Recipient{
		{Email: "a@b.com", Data: map[string]string{"name": "Alice"}},
		{Email: "b@b.com", Data: map[string]string{"name": ""}}, // should be skipped
	}

	tasks, err := cli.PrepareEmailTasks(recipients, tmp.Name(), "Hello {{.name }}")
	if err != nil {
		t.Fatalf("prepareEmailTasks error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Subject != "Hello Alice" {
		t.Errorf("unexpected subject: %s", tasks[0].Subject)
	}
}
