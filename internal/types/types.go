package types

import (
	"encoding/json"
	"time"
)

// CLIArgs is the payload used for scheduled jobs. It mirrors key CLI fields.
type CLIArgs struct {
	EnvPath     string   `json:"env,omitempty"`
	To          string   `json:"to,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Text        string   `json:"text,omitempty"`
	Template    string   `json:"template,omitempty"`
	CSVPath     string   `json:"csv,omitempty"`
	SheetURL    string   `json:"sheet_url,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
	Cc          string   `json:"cc,omitempty"`
	Bcc         string   `json:"bcc,omitempty"`
	Concurrency int      `json:"concurrency,omitempty"`
	RetryLimit  int      `json:"retries,omitempty"`
	BatchSize   int      `json:"batch_size,omitempty"`
	Filter      string   `json:"filter,omitempty"`

	ScheduleAt    string `json:"schedule_at,omitempty"`
	Interval      string `json:"interval,omitempty"`
	Cron          string `json:"cron,omitempty"`
	JobRetries    int    `json:"job_retries,omitempty"`
	JobBackoffDur string `json:"job_backoff,omitempty"` // Go duration, base backoff
}

// Job represents a scheduled unit of work persisted by the scheduler.
// Args contains a JSON-encoded CLIArgs payload.
type Job struct {
	ID        string          `json:"id"`
	Args      json.RawMessage `json:"args"`
	Status    string          `json:"status"` // pending, running, done, cancelled, failed
	RunAt     time.Time       `json:"run_at"`
	CronExpr  string          `json:"cron_expr,omitempty"`
	Interval  string          `json:"interval,omitempty"`

	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
	Backoff     string    `json:"backoff,omitempty"` // base backoff duration

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastRunAt time.Time `json:"last_run_at,omitempty"`
	NextRunAt time.Time `json:"next_run_at,omitempty"`
}

