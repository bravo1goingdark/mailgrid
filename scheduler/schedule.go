package scheduler

import (
	"context"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Scheduler manages all scheduled jobs.
// It supports one-time, interval, and cron-based jobs.
// Jobs are persisted to disk (JSON file) and restored across restarts.
type Scheduler struct {
	mu        sync.Mutex
	jobs      map[string]Job                // persistent jobs
	cron      *cron.Cron                    // cron engine for recurring jobs
	file      string                        // persistence file path
	cronIndex map[string]cron.EntryID       // jobID -> cron entry id
	cancelMap map[string]context.CancelFunc // jobID -> cancel func for timers/tickers
}

// NewScheduler creates a scheduler backed by a JSON file.
// Jobs are loaded from disk, but runtime handlers must be reattached
// by calling ReattachHandlers(handler).
func NewScheduler(file string) *Scheduler {
	s := &Scheduler{
		jobs:      make(map[string]Job),
		cron:      cron.New(),
		file:      file,
		cronIndex: make(map[string]cron.EntryID),
		cancelMap: make(map[string]context.CancelFunc),
	}
	s.load()
	s.cron.Start()
	return s
}

// AddJob saves a job, schedules it in memory, and persists it.
// handler is called whenever the job fires.
func (s *Scheduler) AddJob(job Job, handler func(Job)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
	if err := s.persist(); err != nil {
		return err
	}
	s.scheduleRuntime(job, handler)
	return nil
}

// scheduleRuntime creates timers/cron entries for a job.
func (s *Scheduler) scheduleRuntime(job Job, handler func(Job)) {
	if job.Status == "cancelled" || job.Status == "done" {
		return
	}

	// One-time job
	if !job.RunAt.IsZero() {
		ctx, cancel := context.WithCancel(context.Background())
		s.cancelMap[job.ID] = cancel

		go func(j Job) {
			wait := time.Until(j.RunAt)
			if wait < 0 {
				wait = 0
			}
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
				s.executeJob(j, handler)
			}
		}(job)
		return
	}

	// Interval job
	if job.Interval != "" {
		dur, err := time.ParseDuration(job.Interval)
		if err == nil && dur > 0 {
			ctx, cancel := context.WithCancel(context.Background())
			s.cancelMap[job.ID] = cancel

			go func(j Job) {
				ticker := time.NewTicker(dur)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						s.executeJob(j, handler)
					}
				}
			}(job)
		}
	}

	// Cron job
	if job.CronExpr != "" {
		eid, err := s.cron.AddFunc(job.CronExpr, func() { s.executeJob(job, handler) })
		if err == nil {
			s.cronIndex[job.ID] = eid
		}
	}
}

// ReattachHandlers re-registers runtime scheduling for persisted jobs.
// Call this on startup with a handler closure that dispatches jobs.
func (s *Scheduler) ReattachHandlers(handler func(Job)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, job := range s.jobs {
		if job.Status == "pending" || job.Status == "running" {
			go s.scheduleRuntime(job, handler)
		}
	}
}

// executeJob updates job state, runs the handler, and persists state.
func (s *Scheduler) executeJob(job Job, handler func(Job)) {
	// Mark running
	s.mu.Lock()
	j, ok := s.jobs[job.ID]
	if !ok || j.Status == "cancelled" {
		s.mu.Unlock()
		return
	}
	j.Status = "running"
	s.jobs[job.ID] = j
	_ = s.persist()
	s.mu.Unlock()

	// Run a job outside lock
	handler(j)

	// Mark done for one-time jobs
	s.mu.Lock()
	j = s.jobs[job.ID]
	if j.CronExpr == "" && j.Interval == "" {
		j.Status = "done"
	} else {
		j.Status = "pending"
	}
	s.jobs[job.ID] = j
	_ = s.persist()
	s.mu.Unlock()
}

// CancelJob stops a job’s timers/cron entries and marks it canceled.
func (s *Scheduler) CancelJob(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return false
	}

	if cancel, ok := s.cancelMap[id]; ok {
		cancel()
		delete(s.cancelMap, id)
	}
	if eid, ok := s.cronIndex[id]; ok {
		s.cron.Remove(eid)
		delete(s.cronIndex, id)
	}

	job.Status = "cancelled"
	s.jobs[id] = job
	_ = s.persist()
	return true
}

// ListJobs returns all jobs currently tracked.
func (s *Scheduler) ListJobs() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	return jobs
}

// persist writes jobs to the JSON persistence file.
func (s *Scheduler) persist() error {
	dir := filepath.Dir(s.file)
	if dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	data, err := json.MarshalIndent(s.jobs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0o644)
}

// load restores jobs from the JSON persistence file.
func (s *Scheduler) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return
	}
	var jobs map[string]Job
	if err := json.Unmarshal(data, &jobs); err == nil {
		s.jobs = jobs
	}
}
