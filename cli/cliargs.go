package cli

import "github.com/spf13/pflag"

// CLIArgs holds all configurable options passed via the command line.
// This struct is used throughout the Mailgrid CLI flow.
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
	WebhookURL   string   // HTTP URL to send completion notification

	// Monitoring options
	Monitor     bool // Enable real-time monitoring dashboard
	MonitorPort int  // Port for monitoring dashboard

	// Scheduling options (if any of these are set, we schedule instead of immediate sending)
	ScheduleAt string // RFC3339 timestamp
	Interval   string // Go duration, e.g. "1h", "30m"
	Cron       string // Cron expression (5-field)

	// Scheduler job-level retry/backoff (separate it from SMTP retries)
	JobRetries int
	JobBackoff string // duration

	// Job management
	ListJobs     bool
	CancelJobID  string
	SchedulerRun bool // Run dispatcher in foreground

	SchedulerDB string // Path to BoltDB file for persisted schedules

	// Version flag
	ShowVersion bool // Show version information and exit

	// Offset tracking for resumable delivery
	Resume      bool   // Resume from last saved offset
	ResetOffset bool   // Clear offset file and start from beginning
	OffsetFile  string // Custom path to offset file
}

// ParseFlags reads command-line flags using spf13/pflag and returns a filled CLIArgs struct.
func ParseFlags() CLIArgs {
	var args CLIArgs

	pflag.StringVarP(&args.EnvPath, "env", "e", "", "Path to SMTP config JSON")
	pflag.StringVarP(&args.CSVPath, "csv", "f", "", "Path to recipient CSV file")
	pflag.StringVarP(&args.SheetURL, "sheet-url", "u", "", "Public Google Sheet URL (replaces --csv)")
	pflag.StringVarP(&args.TemplatePath, "template", "t", "", "Path to email HTML template")
	pflag.StringVar(&args.Cc, "cc", "", "Comma-separated emails or file path for CC")
	pflag.StringVar(&args.Bcc, "bcc", "", "Comma-separated emails or file path for BCC")
	pflag.StringVarP(&args.Subject, "subject", "s", "Test Email from Mailgrid", "Email subject (templated with {{ .field }})")
	pflag.BoolVarP(&args.DryRun, "dry-run", "d", false, "Render emails to console without sending")
	pflag.BoolVarP(&args.ShowPreview, "preview", "p", false, "Start a local preview server to view rendered email")
	pflag.IntVar(&args.PreviewPort, "port", 8080, "Port for preview server")
	pflag.IntVarP(&args.Concurrency, "concurrency", "c", 1, "Number of concurrent SMTP workers")
	pflag.IntVarP(&args.RetryLimit, "retries", "r", 1, "Retry attempts per failed email")
	pflag.IntVarP(&args.BatchSize, "batch-size", "b", 1, "Number of emails per SMTP batch")
	pflag.StringVarP(&args.Filter, "filter", "F", "", "Logical filter for recipients")
	pflag.StringSliceVarP(&args.Attachments, "attach", "a", []string{}, "File attachments (repeat flag to add multiple)")
	pflag.StringVar(&args.To, "to", "", "Email address for single-recipient sending (mutually exclusive with --csv or --sheet-url)")
	pflag.StringVar(&args.Text, "text", "", "Inline plain-text body or path to a .txt file (mutually exclusive with --template)")
	pflag.StringVarP(&args.WebhookURL, "webhook", "w", "", "HTTP URL to send POST request with campaign results")

	// Monitoring flags
	pflag.BoolVarP(&args.Monitor, "monitor", "m", false, "Enable real-time monitoring dashboard")
	pflag.IntVar(&args.MonitorPort, "monitor-port", 9091, "Port for monitoring dashboard")

	// Scheduling flags (single-letter shorthands)
	pflag.StringVarP(&args.ScheduleAt, "schedule-at", "A", "", "Schedule time in RFC3339 (e.g., 2025-09-08T09:00:00Z)")
	pflag.StringVarP(&args.Interval, "interval", "i", "", "Repeat interval as Go duration (e.g., 1h, 30m)")
	pflag.StringVarP(&args.Cron, "cron", "C", "", "Cron expression (5-field) for recurring schedules")
	pflag.IntVarP(&args.JobRetries, "job-retries", "J", 3, "Scheduler-level retry attempts on handler failure")
	pflag.StringVarP(&args.JobBackoff, "job-backoff", "B", "2s", "Base backoff for scheduler retries (Go duration)")
	pflag.BoolVarP(&args.ListJobs, "jobs-list", "L", false, "List scheduled jobs")
	pflag.StringVarP(&args.CancelJobID, "jobs-cancel", "X", "", "Cancel job by ID")
	pflag.BoolVarP(&args.SchedulerRun, "scheduler-run", "R", false, "Run the scheduler dispatcher in the foreground")
	pflag.StringVarP(&args.SchedulerDB, "scheduler-db", "D", "mailgrid.db", "Path to BoltDB file for schedules")

	// Version flag
	pflag.BoolVar(&args.ShowVersion, "version", false, "Show version information and exit")

	// Offset tracking flags
	pflag.BoolVar(&args.Resume, "resume", false, "Resume sending from last saved offset")
	pflag.BoolVar(&args.ResetOffset, "reset-offset", false, "Clear offset file and start from beginning")
	pflag.StringVar(&args.OffsetFile, "offset-file", ".mailgrid.offset", "Custom path to offset file")

	pflag.Parse()

	return args
}
