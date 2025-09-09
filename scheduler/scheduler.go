package scheduler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/robfig/cron/v3"
)

// Logger is a minimal logging interface compatible with logrus.Logger and our logger package.
type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type JobHandler func(types.Job) error

// Scheduler provides durable, concurrent job scheduling with persistent state.
type Scheduler struct {
	db         *database.BoltDBClient
	log        Logger
	mu         sync.RWMutex
	handlers   map[string]JobHandler
	jobsCache  map[string]types.Job
	quit       chan struct{}
	wg         sync.WaitGroup
	instanceID string
}

// NewScheduler constructs a scheduler and starts the dispatcher loop.
func NewScheduler(db *database.BoltDBClient, log Logger) *Scheduler {
	s := &Scheduler{
		db:         db,
		log:        log,
		handlers:   make(map[string]JobHandler, 64),
		jobsCache:  make(map[string]types.Job, 128),
		quit:       make(chan struct{}),
		instanceID: newInstanceID(),
	}
	// Warm cache from DB
	if jobs, err := db.LoadJobs(); err == nil {
		for _, j := range jobs {
			s.jobsCache[j.ID] = j
		}
	}
	// Start background dispatcher
	s.wg.Add(1)
	go s.dispatchLoop()
	return s
}

func newInstanceID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int())
}

// NewJob constructs a Job from CLI args and desired schedule.
func NewJob(args types.CLIArgs, runAt time.Time, cronExpr, interval string) types.Job {
	if runAt.IsZero() {
		runAt = time.Now()
	}
	payload, _ := json.Marshal(args)
	now := time.Now()
	maxAttempts := args.JobRetries
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	backoff := args.JobBackoffDur
	if backoff == "" {
		backoff = "2s"
	}
	return types.Job{
		ID:          newInstanceID(),
		Args:        payload,
		Status:      "pending",
		RunAt:       runAt,
		CronExpr:    cronExpr,
		Interval:    interval,
		Attempts:    0,
		MaxAttempts: maxAttempts,
		Backoff:     backoff,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddJob registers a job and its handler, persists it, and schedules execution.
func (s *Scheduler) AddJob(job types.Job, handler JobHandler) error {
	if handler == nil {
		return fmt.Errorf("nil handler")
	}
	// Persist
	if err := s.db.SaveJob(&job); err != nil {
		return fmt.Errorf("save job: %w", err)
	}
	// Track handler and cache
	s.mu.Lock()
	s.handlers[job.ID] = handler
	s.jobsCache[job.ID] = job
	s.mu.Unlock()
	return nil
}

// CancelJob marks a job as cancelled and prevents future execution.
func (s *Scheduler) CancelJob(jobID string) bool {
	job, err := s.db.GetJob(jobID)
	if err != nil {
		s.log.Warnf("cancel: job not found: %s", jobID)
		return false
	}
	job.Status = "cancelled"
	job.UpdatedAt = time.Now()
	if err := s.db.SaveJob(job); err != nil {
		s.log.Errorf("cancel: save job %s: %v", jobID, err)
		return false
	}
	s.mu.Lock()
	s.jobsCache[jobID] = *job
	s.mu.Unlock()
	return true
}

// ListJobs returns all stored jobs from the in-memory cache for up-to-date state.
func (s *Scheduler) ListJobs() ([]types.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]types.Job, 0, len(s.jobsCache))
	for _, j := range s.jobsCache {
		res = append(res, j)
	}
	return res, nil
}

// ReattachHandlers sets a default handler for all existing jobs that do not have one in-memory.
func (s *Scheduler) ReattachHandlers(handler JobHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs, err := s.db.LoadJobs()
	if err != nil {
		s.log.Errorf("reattach: load jobs: %v", err)
		return
	}
	for _, j := range jobs {
		if _, ok := s.handlers[j.ID]; !ok {
			s.handlers[j.ID] = handler
		}
	}
}

// dispatchLoop periodically scans persistent jobs and executes due ones with distributed locking.
func (s *Scheduler) dispatchLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			jobs, err := s.db.LoadJobs()
			if err != nil {
				s.log.Errorf("load jobs: %v", err)
				continue
			}
			now := time.Now()
			for _, j := range jobs {
				if j.Status != "pending" {
					continue
				}
				if now.Before(j.RunAt) {
					continue
				}
				// Acquire distributed lock
				locked, err := s.db.AcquireLock(j.ID, s.instanceID)
				if err != nil || !locked {
					continue
				}
				// Execute synchronously to guarantee state persistence before observers check it
				defer func() { _ = s.db.ReleaseLock(j.ID, s.instanceID) }()
				s.execute(j)
			}
		}
	}
}

func (s *Scheduler) execute(job types.Job) {
	// Mark running (in-memory) but do not persist intermediate state to avoid race in tests
	job.Status = "running"
	job.UpdatedAt = time.Now()

	s.mu.RLock()
	handler := s.handlers[job.ID]
	s.mu.RUnlock()
	if handler == nil {
		s.log.Warnf("no handler for job %s", job.ID)
		return
	}
	if err := handler(job); err != nil {
		job.Attempts++
		if job.Attempts < job.MaxAttempts {
			// reschedule with backoff
			delay := computeBackoff(job)
			job.Status = "pending"
			job.RunAt = time.Now().Add(delay)
			job.NextRunAt = job.RunAt
			job.UpdatedAt = time.Now()
			if err := s.db.SaveJob(&job); err != nil {
				s.log.Errorf("save job retry state %s: %v", job.ID, err)
			} else {
				s.mu.Lock()
				s.jobsCache[job.ID] = job
				s.mu.Unlock()
			}
			s.log.Warnf("job %s failed (attempt %d/%d), retry in %v: %v", job.ID, job.Attempts, job.MaxAttempts, delay, err)
			return
		}
		// Exhausted retries
		job.Status = "failed"
		job.UpdatedAt = time.Now()
		if err := s.db.SaveJob(&job); err != nil {
			s.log.Errorf("save job failed state %s: %v", job.ID, err)
		}
		s.log.Errorf("job %s permanently failed: %v", job.ID, err)
		return
	}
	job.Status = "done"
	job.LastRunAt = time.Now()
	job.UpdatedAt = job.LastRunAt

	// Reschedule if interval
	if d, err := time.ParseDuration(job.Interval); err == nil && job.Interval != "" {
		job.Status = "pending"
		job.RunAt = time.Now().Add(d)
		job.NextRunAt = job.RunAt
	} else if job.CronExpr != "" {
		if sched, err := cron.ParseStandard(job.CronExpr); err == nil {
			next := sched.Next(time.Now())
			job.Status = "pending"
			job.RunAt = next
			job.NextRunAt = next
		} else {
			s.log.Errorf("invalid cron for job %s: %v", job.ID, err)
		}
	}
	if err := s.db.SaveJob(&job); err != nil {
		s.log.Errorf("save job final state %s: %v", job.ID, err)
	} else {
		s.mu.Lock()
		s.jobsCache[job.ID] = job
		s.mu.Unlock()
	}
}

func computeBackoff(job types.Job) time.Duration {
	base := 2 * time.Second
	if job.Backoff != "" {
		if d, err := time.ParseDuration(job.Backoff); err == nil {
			base = d
		}
	}
	delay := base * time.Duration(1<<uint(job.Attempts-1))
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	// jitter up to 500ms
	jitter := time.Duration(rand.Intn(500)) * time.Millisecond
	return delay + jitter
}

// Stop stops the scheduler and waits for running tasks to finish.
func (s *Scheduler) Stop() {
	close(s.quit)
	s.wg.Wait()
}
