package cli

import "github.com/spf13/pflag"

// CLIArgs holds all configurable options passed via the command line.
// It is populated once in ParseFlags() and then passed around the app.
type CLIArgs struct {
	EnvPath      string   // Path to an SMTP config JSON file
	CSVPath      string   // Path to recipient CSV file
	TemplatePath string   // Path to HTML email template
	Subject      string   // Subject line (supports templating with {{ .name }})
	DryRun       bool     // If true, render but do not send emails
	ShowPreview  bool     // If true, serve rendered HTML via localhost
	PreviewPort  int      // Port to run the preview server on
	Concurrency  int      // Number of parallel SMTP workers
	RetryLimit   int      // Max retry attempts for failed sending
	BatchSize    int      // Number of emails sent per SMTP batch
	SheetURL     string   // Optional Google Sheet URL for CSV import
	Filter       string   // Logical filter expression for recipients
	Attachments  []string // File paths to attach to every email
	Cc           string   // Comma-separated emails or file path for CC
	Bcc          string   // Comma-separated emails or file path for BCC
	To           string   // Email address for one-off sending
	Text         string   // Inline plain-text body or path to a text file

	// Scheduling options
	ScheduleAt  string // RFC3339 timestamp for one-time job
	CronExpr    string // Cron expression for recurring job
	Interval    string // Go duration (e.g. "10s", "5m", "24h")
	CancelJobID string // Cancel a scheduled job by ID
	ListJobs    bool   // List scheduled jobs
}

// ParseFlags reads command-line flags into CLIArgs using spf13/pflag.
// Returns a fully populated CLIArgs struct.
func ParseFlags() CLIArgs {
	var args CLIArgs

	pflag.StringVar(&args.EnvPath, "env", "", "Path to SMTP config JSON")
	pflag.StringVar(&args.CSVPath, "csv", "", "Path to recipient CSV file")
	pflag.StringVar(&args.SheetURL, "sheet-url", "", "Public Google Sheet URL (replaces --csv)")
	pflag.StringVarP(&args.TemplatePath, "template", "t", "", "Path to email HTML template")
	pflag.StringVar(&args.Cc, "cc", "", "Comma-separated emails or file path for CC")
	pflag.StringVar(&args.Bcc, "bcc", "", "Comma-separated emails or file path for BCC")
	pflag.StringVarP(&args.Subject, "subject", "s", "Test Email from Mailgrid", "Email subject (templated with {{ .field }})")
	pflag.BoolVar(&args.DryRun, "dry-run", false, "Render emails to console without sending")
	pflag.BoolVarP(&args.ShowPreview, "preview", "p", false, "Start a local preview server to view rendered email")
	pflag.IntVar(&args.PreviewPort, "port", 8080, "Port for preview server")
	pflag.IntVarP(&args.Concurrency, "concurrency", "c", 1, "Number of concurrent SMTP workers")
	pflag.IntVarP(&args.RetryLimit, "retries", "r", 1, "Retry attempts per failed email")
	pflag.IntVar(&args.BatchSize, "batch-size", 1, "Number of emails per SMTP batch")
	pflag.StringVar(&args.Filter, "filter", "", "Logical filter for recipients")
	pflag.StringSliceVar(&args.Attachments, "attach", []string{}, "File attachments (repeat flag to add multiple)")
	pflag.StringVar(&args.To, "to", "", "Email address for single-recipient sending (mutually exclusive with --csv or --sheet-url)")
	pflag.StringVar(&args.Text, "text", "", "Inline plain-text body or path to a .txt file (mutually exclusive with --template)")

	pflag.StringVar(&args.ScheduleAt, "at", "", "Schedule send time (RFC3339 format: 2025-09-10T10:30:00)")
	pflag.StringVar(&args.CronExpr, "cron", "", "Cron expression for recurring sends (e.g. '0 9 * * MON')")
	pflag.StringVar(&args.Interval, "every", "", "Interval for repeated sends (Go duration: '10s', '5m', '24h')")
	pflag.StringVar(&args.CancelJobID, "cancel", "", "Cancel a scheduled job by its ID")
	pflag.BoolVar(&args.ListJobs, "list", false, "List all scheduled jobs")

	pflag.Parse()
	return args
}
