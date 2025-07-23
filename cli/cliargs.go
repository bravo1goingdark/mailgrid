package cli

import "github.com/spf13/pflag"

// CLIArgs holds all configurable options passed via the command line.
// This struct is used throughout the Mailgrid CLI flow.
type CLIArgs struct {
	EnvPath      string // Path to an SMTP config JSON file
	CSVPath      string // Path to recipient CSV file
	TemplatePath string // Path to HTML email template
	Subject      string // Subject line (supports templating with {{ .name }})
	DryRun       bool   // If true, render but do not send emails
	ShowPreview  bool   // If true, serve rendered HTML via localhost
	PreviewPort  int    // Port to run the preview server on
	Concurrency  int    // Number of parallel SMTP workers
	RetryLimit   int    // Max retry attempts for failed sending
	BatchSize    int    // Number of emails sent per SMTP batch
	Resume       bool   // Resume sending from saved offset
	ResetOffset  bool   // Ignore offset and start fresh
}

// ParseFlags reads command-line flags using spf13/pflag and returns a filled CLIArgs struct.
func ParseFlags() CLIArgs {
	var args CLIArgs

	pflag.StringVar(&args.EnvPath, "env", "example/config.json", "Path to SMTP config JSON")
	pflag.StringVar(&args.CSVPath, "csv", "example/test_contacts.csv", "Path to recipient CSV file")
	pflag.StringVarP(&args.TemplatePath, "template", "t", "example/welcome.html", "Path to email HTML template")
	pflag.StringVarP(&args.Subject, "subject", "s", "Test Email from Mailgrid", "Email subject (templated with {{ .field }})")
	pflag.BoolVar(&args.DryRun, "dry-run", false, "Render emails to console without sending")
	pflag.BoolVarP(&args.ShowPreview, "preview", "p", false, "Start a local preview server to view rendered email")
	pflag.IntVar(&args.PreviewPort, "port", 8080, "Port for preview server")
	pflag.IntVarP(&args.Concurrency, "concurrency", "c", 1, "Number of concurrent SMTP workers")
	pflag.IntVarP(&args.RetryLimit, "retries", "r", 1, "Retry attempts per failed email")
	pflag.IntVar(&args.BatchSize, "batch-size", 1, "Number of emails per SMTP batch")
	pflag.BoolVar(&args.Resume, "resume", false, "Resume sending from saved offset")
	pflag.BoolVar(&args.ResetOffset, "reset-offset", false, "Ignore offset and start from beginning")
	pflag.Parse()

	return args
}
