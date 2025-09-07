package scheduler

import (
	"time"

	"github.com/bravo1goingdark/mailgrid/internal/types"
)

// EmailScheduler embeds Scheduler and handles email-specific jobs.
type EmailScheduler struct {
	*Scheduler
}

// NewEmailScheduler creates a new scheduler for email jobs.
func NewEmailScheduler(s *Scheduler) *EmailScheduler {
	return &EmailScheduler{Scheduler: s}
}

// AddEmailJob creates a new job and adds it to the scheduler.
func (es *EmailScheduler) AddEmailJob(args types.CLIArgs, runAt time.Time, cronExpr, interval string, handler func(types.Job) error) error {
	job := NewJob(args, runAt, cronExpr, interval)
	return es.AddJob(job, handler)
}

// ReattachHandlers rebinds handlers for persisted jobs after process restarts.
func (es *EmailScheduler) ReattachHandlers(handler func(types.Job) error) {
	es.Scheduler.ReattachHandlers(handler)
}
