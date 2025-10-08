package cli_test

import (
	"os"
	"testing"

	"github.com/bravo1goingdark/mailgrid/cli"
	"github.com/bravo1goingdark/mailgrid/parser"
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

	a, err := os.CreateTemp(t.TempDir(), "a*.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.WriteString("file"); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}

	tasks, err := cli.PrepareEmailTasks(recipients, tmp.Name(), "Hello {{.name }}", []string{a.Name()}, []string{}, []string{})
	if err != nil {
		t.Fatalf("prepareEmailTasks error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Subject != "Hello Alice" {
		t.Errorf("unexpected subject: %s", tasks[0].Subject)
	}
	if len(tasks[0].Attachments) != 1 || tasks[0].Attachments[0] != a.Name() {
		t.Errorf("attachments not set correctly")
	}
}

func TestPrepareEmailTasks_AttachOnly(t *testing.T) {
	recipients := []parser.Recipient{{Email: "a@b.com", Data: map[string]string{"name": "A"}}}

	a, err := os.CreateTemp(t.TempDir(), "a*.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.WriteString("file"); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}

	tasks, err := cli.PrepareEmailTasks(recipients, "", "Hi", []string{a.Name()}, []string{}, []string{})
	if err != nil {
		t.Fatalf("prepareEmailTasks error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Body != "" {
		t.Errorf("expected empty body")
	}
}

func TestPrepareEmailTasks_CC_BCC(t *testing.T) {
	recipients := []parser.Recipient{
		{Email: "jacob@example.com", Data: map[string]string{"name": "Jacob"}},
	}

	tasks, err := cli.PrepareEmailTasks(
		recipients,
		"",
		"Test Subject",
		[]string{},
		[]string{"cc1@example.com"},
		[]string{"bcc1@example.com"},
	)
	if err != nil {
		t.Fatalf("prepareEmailTasks error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]

	if len(task.CC) != 1 || task.CC[0] != "cc1@example.com" {
		t.Errorf("expected cc to be [cc1@example.com], got %v", task.CC)
	}

	if len(task.BCC) != 1 || task.BCC[0] != "bcc1@example.com" {
		t.Errorf("expected bcc to be [bcc1@example.com], got %v", task.BCC)
	}
}
