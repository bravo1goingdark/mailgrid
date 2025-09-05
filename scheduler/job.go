package scheduler

import (
	"github.com/bravo1goingdark/mailgrid/cli"
	"github.com/google/uuid"
	"time"
)

// Job represents a scheduled email sent.
// It is serializable to JSON, so it can be persisted and restored after restart.
type Job struct {
	ID        string      `json:"id"`               // Unique job ID
	Args      cli.CLIArgs `json:"args"`             // CLI arguments used for the job
	RunAt     time.Time   `json:"run_at,omitempty"` // One-time execution time
	CronExpr  string      `json:"cron_expr,omitempty"`
	Interval  string      `json:"interval,omitempty"` // Interval duration string
	CreatedAt time.Time   `json:"created_at"`
	Status    string      `json:"status"` // pending, running, done, canceled
}

// NewJob creates a new Job with a generated UUID and "pending" status.
func NewJob(args cli.CLIArgs, runAt time.Time, cronExpr string, interval string) Job {
	return Job{
		ID:        uuid.NewString(),
		Args:      args,
		RunAt:     runAt,
		CronExpr:  cronExpr,
		Interval:  interval,
		CreatedAt: time.Now(),
		Status:    "pending",
	}
}
