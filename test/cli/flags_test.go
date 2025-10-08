package cli_test

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/bravo1goingdark/mailgrid/cli"
)

func TestParseFlags_AllOptions(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"mailgrid", "--env", "cfg.json", "--csv", "rec.csv", "--template", "body.html", "--subject", "Hello", "--dry-run", "--preview", "--port", "9090", "--concurrency", "3", "--retries", "4", "--batch-size", "2", "--filter", "tier = \"pro\"", "--attach", "a.pdf", "--attach", "b.pdf"}
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	args := cli.ParseFlags()

	if args.EnvPath != "cfg.json" || args.CSVPath != "rec.csv" || args.TemplatePath != "body.html" || args.Subject != "Hello" {
		t.Fatalf("string args mismatch: %+v", args)
	}
	if !args.DryRun || !args.ShowPreview || args.PreviewPort != 9090 || args.Concurrency != 3 || args.RetryLimit != 4 || args.BatchSize != 2 {
		t.Fatalf("bool/int args mismatch: %+v", args)
	}
	if args.Filter != "tier = \"pro\"" {
		t.Fatalf("filter mismatch: %q", args.Filter)
	}
	if len(args.Attachments) != 2 || args.Attachments[0] != "a.pdf" || args.Attachments[1] != "b.pdf" {
		t.Fatalf("attachments mismatch: %+v", args.Attachments)
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"mailgrid"}
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	args := cli.ParseFlags()

	if args.Subject != "Test Email from Mailgrid" || args.PreviewPort != 8080 || args.Concurrency != 1 || args.RetryLimit != 1 || args.BatchSize != 1 {
		t.Fatalf("defaults mismatch: %+v", args)
	}
	if len(args.Attachments) != 0 {
		t.Fatalf("expected no attachments by default")
	}
}
