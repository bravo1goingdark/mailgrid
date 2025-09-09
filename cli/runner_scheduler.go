package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/scheduler"
)

// Runner wires the EmailScheduler with concrete email-sending logic.
type Runner struct {
	es *scheduler.EmailScheduler
}

func NewRunner(es *scheduler.EmailScheduler) *Runner {
	return &Runner{es: es}
}

// EmailJobHandler executes a scheduled email job by decoding the payload and sending a single email.
func (r *Runner) EmailJobHandler(job types.Job) error {
	var a types.CLIArgs
	if err := json.Unmarshal(job.Args, &a); err != nil {
		return fmt.Errorf("decode job args: %w", err)
	}
	// If To is set, send one-off; if CSV or Sheet, send bulk via standard CLI pipeline.
	if a.To != "" {
		cfg, err := config.LoadConfig(a.EnvPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		args := CLIArgs{
			EnvPath:      a.EnvPath,
			To:           a.To,
			Subject:      a.Subject,
			TemplatePath: a.Template,
			Text:         a.Text,
			Attachments:  a.Attachments,
			Cc:           a.Cc,
			Bcc:          a.Bcc,
			RetryLimit:   a.RetryLimit,
		}
		return SendSingleEmail(args, cfg.SMTP)
	}
	// Bulk path: call Run with a CLIArgs built from payload
	mapped := CLIArgs{
		EnvPath:      a.EnvPath,
		CSVPath:      a.CSVPath,
		SheetURL:     a.SheetURL,
		TemplatePath: a.Template,
		Subject:      a.Subject,
		Attachments:  a.Attachments,
		Cc:           a.Cc,
		Bcc:          a.Bcc,
		Concurrency:  a.Concurrency,
		RetryLimit:   a.RetryLimit,
		BatchSize:    a.BatchSize,
		Filter:       a.Filter,
	}
	return Run(mapped)
}

// Run schedules (or immediately executes) an email job based on the provided args.
func (r *Runner) Run(_ context.Context, a types.CLIArgs) error {
	var runAt time.Time
	if a.ScheduleAt != "" {
		var err error
		runAt, err = time.Parse(time.RFC3339, a.ScheduleAt)
		if err != nil {
			return fmt.Errorf("parse schedule_at: %w", err)
		}
	} else {
		runAt = time.Now()
	}
	return r.es.AddEmailJob(a, runAt, a.Cron, a.Interval, r.EmailJobHandler)
}

