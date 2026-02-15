package cli

import (
	"fmt"
	"github.com/spf13/pflag"
)

// CLIArgs holds all configurable options passed via the command line.
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

	// Monitoring
	Monitor     bool // Enable real-time monitoring dashboard
	MonitorPort int  // Port for monitoring dashboard (includes metrics)

	// Scheduling
	ScheduleAt string // RFC3339 timestamp
	Interval   string // Go duration, e.g. "1h", "30m"
	Cron       string // Cron expression (5-field)
	JobRetries int    // Scheduler retry attempts

	// Job management
	ListJobs     bool   // List scheduled jobs
	CancelJobID  string // Cancel job by ID
	SchedulerRun bool   // Run dispatcher in foreground

	// Version
	ShowVersion bool // Show version information and exit

	// Help
	ShowHelp bool // Show help message

	// Offset tracking
	Resume      bool // Resume from last saved offset
	ResetOffset bool // Clear offset file and start from beginning
}

// printHelp prints a custom formatted help message with grouped flags
func printHelp() {
	fmt.Println("MailGrid - A production-ready email orchestrator for bulk email campaigns")
	fmt.Println()
	fmt.Println("Usage: mailgrid [flags]")
	fmt.Println()
	fmt.Println("RECIPIENT SOURCE (provide one):")
	fmt.Println("  -f, --csv              string   Path to recipient CSV file")
	fmt.Println("  -u, --sheet-url        string   Public Google Sheet URL (replaces --csv)")
	fmt.Println("      --to               string   Email address for single-recipient sending")
	fmt.Println()
	fmt.Println("EMAIL CONTENT:")
	fmt.Println("  -t, --template         string   Path to email HTML template")
	fmt.Println("      --text             string   Inline plain-text body or path to a .txt file")
	fmt.Println("  -s, --subject          string   Email subject (templated with {{ .field }})")
	fmt.Println("  -a, --attach           strings  File attachments (repeat flag to add multiple)")
	fmt.Println("      --cc               string   Comma-separated emails or file path for CC")
	fmt.Println("      --bcc              string   Comma-separated emails or file path for BCC")
	fmt.Println()
	fmt.Println("RECIPIENT FILTERING:")
	fmt.Println("  -F, --filter           string   Logical filter for recipients")
	fmt.Println()
	fmt.Println("SMTP CONFIGURATION:")
	fmt.Println("  -e, --env              string   Path to SMTP config JSON")
	fmt.Println("  -c, --concurrency      int      Number of concurrent SMTP workers")
	fmt.Println("  -b, --batch-size       int      Number of emails per SMTP batch")
	fmt.Println("  -r, --retries          int      Retry attempts per failed email")
	fmt.Println()
	fmt.Println("RESUMABLE SENDING:")
	fmt.Println("      --resume                    Resume sending from last saved offset")
	fmt.Println("      --reset-offset              Clear offset file and start from beginning")
	fmt.Println()
	fmt.Println("SCHEDULING:")
	fmt.Println("  -R, --scheduler-run             Run the scheduler dispatcher in the foreground")
	fmt.Println("  -A, --schedule-at      string   Schedule time in RFC3339 format")
	fmt.Println("  -i, --interval         string   Repeat interval as Go duration (e.g., 1h, 30m)")
	fmt.Println("  -C, --cron             string   Cron expression (5-field) for recurring schedules")
	fmt.Println("  -J, --job-retries      int      Scheduler-level retry attempts on handler failure")
	fmt.Println("  -L, --jobs-list               List scheduled jobs")
	fmt.Println("  -X, --jobs-cancel      string   Cancel job by ID")
	fmt.Println()
	fmt.Println("MONITORING:")
	fmt.Println("  -m, --monitor                   Enable real-time monitoring dashboard")
	fmt.Println("      --monitor-port     int      Port for monitoring dashboard and metrics")
	fmt.Println()
	fmt.Println("TESTING & DEBUG:")
	fmt.Println("  -d, --dry-run                   Render emails to console without sending")
	fmt.Println("  -p, --preview                   Start a local preview server to view rendered email")
	fmt.Println("      --port             int      Port for preview server")
	fmt.Println()
	fmt.Println("NOTIFICATIONS:")
	fmt.Println("  -w, --webhook          string   HTTP URL to send POST request with campaign results")
	fmt.Println()
	fmt.Println("INFO:")
	fmt.Println("      --version                   Show version information and exit")
	fmt.Println("  -h, --help                      Show this help message")
	fmt.Println()
	fmt.Println("DEFAULTS:")
	fmt.Println("  --concurrency 1, --batch-size 1, --retries 1, --monitor-port 9091, --port 8080")
	fmt.Println("  --subject \"Test Email from Mailgrid\", --job-retries 3")
}

func ParseFlags() CLIArgs {
	var args CLIArgs
	var showHelp bool

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

	pflag.BoolVarP(&args.Monitor, "monitor", "m", false, "Enable real-time monitoring dashboard")
	pflag.IntVar(&args.MonitorPort, "monitor-port", 9091, "Port for monitoring dashboard and metrics")

	pflag.StringVarP(&args.ScheduleAt, "schedule-at", "A", "", "Schedule time in RFC3339 (e.g., 2025-09-08T09:00:00Z)")
	pflag.StringVarP(&args.Interval, "interval", "i", "", "Repeat interval as Go duration (e.g., 1h, 30m)")
	pflag.StringVarP(&args.Cron, "cron", "C", "", "Cron expression (5-field) for recurring schedules")
	pflag.IntVarP(&args.JobRetries, "job-retries", "J", 3, "Scheduler-level retry attempts on handler failure")
	pflag.BoolVarP(&args.ListJobs, "jobs-list", "L", false, "List scheduled jobs")
	pflag.StringVarP(&args.CancelJobID, "jobs-cancel", "X", "", "Cancel job by ID")
	pflag.BoolVarP(&args.SchedulerRun, "scheduler-run", "R", false, "Run the scheduler dispatcher in the foreground")

	pflag.BoolVar(&args.ShowVersion, "version", false, "Show version information and exit")

	pflag.BoolVar(&args.Resume, "resume", false, "Resume sending from last saved offset")
	pflag.BoolVar(&args.ResetOffset, "reset-offset", false, "Clear offset file and start from beginning")

	// Add help flag manually to control behavior
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show this help message")

	pflag.Parse()

	if showHelp {
		printHelp()
		args.ShowHelp = true
		return args
	}

	return args
}
