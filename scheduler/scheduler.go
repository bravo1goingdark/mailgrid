package scheduler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	mrand "math/rand"
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
	if jobs, err := db.LoadJobs(); err != nil {
		s.log.Warnf("Failed to load jobs from database: %v", err)
	} else {
		for _, j := range jobs {
			s.jobsCache[j.ID] = j
		}
	}
	// Start background dispatcher
	s.wg.Add(1)
	go s.dispatchLoop()
	return s
}

// newInstanceID generates a unique instance ID using UUID-like format
// This provides better uniqueness than timestamp + random number
func newInstanceID() string {
	// Generate 8 random bytes (64 bits) for collision resistance
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
	}

	// Format as hex string with timestamp prefix for sortability
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%x-%x", timestamp, b)
}

// NewJob constructs a Job from CLI args and desired schedule.
func NewJob(args types.CLIArgs, runAt time.Time, cronExpr, interval string) (types.Job, error) {
	if runAt.IsZero() {
		runAt = time.Now()
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return types.Job{}, fmt.Errorf("marshal job args: %w", err)
	}
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
	}, nil
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

// ListJobs returns all stored jobs, always reading from the DB so external
// modifications (e.g. from another process or test) are reflected immediately.
func (s *Scheduler) ListJobs() ([]types.Job, error) {
	return s.db.LoadJobs()
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
// Jitter (±500ms) is added to the ticker so multiple instances don't thunder-herd.
func (s *Scheduler) dispatchLoop() {
	defer s.wg.Done()

	// ±200ms jitter prevents thundering herd when multiple instances run.
	// 1s base interval is responsive while remaining cheap (BoltDB reads are fast).
	jitterMs := time.Duration(mrand.Int63n(400)) * time.Millisecond
	ticker := time.NewTicker(1*time.Second + jitterMs)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			// Always query the DB for the authoritative job list (never the stale in-memory cache).
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
				// Acquire distributed lock before execution so only one instance runs a job.
				locked, err := s.db.AcquireLock(j.ID, s.instanceID)
				if err != nil || !locked {
					continue
				}
				// Release lock unconditionally (even on panic) via defer inside execute wrapper.
				s.executeWithLock(j)
			}
		}
	}
}

// executeWithLock runs the job and ensures the distributed lock is always released,
// even if the handler panics. This prevents jobs from being permanently stuck in
// a "running" state when the process crashes mid-execution.
func (s *Scheduler) executeWithLock(job types.Job) {
	defer func() {
		if err := s.db.ReleaseLock(job.ID, s.instanceID); err != nil {
			s.log.Errorf("release lock for job %s: %v", job.ID, err)
		}
	}()
	s.execute(job)
}

func (s *Scheduler) execute(job types.Job) {
	// Recover from handler panics so a bad job doesn't crash the entire dispatch loop.
	defer func() {
		if r := recover(); r != nil {
			s.log.Errorf("job %s panicked: %v", job.ID, r)
			job.Status = "failed"
			job.UpdatedAt = time.Now()
			if err := s.db.SaveJob(&job); err != nil {
				s.log.Errorf("save panicked job state %s: %v", job.ID, err)
			}
			s.mu.Lock()
			s.jobsCache[job.ID] = job
			s.mu.Unlock()
		}
	}()

	job.Status = "running"
	job.UpdatedAt = time.Now()

	s.mu.RLock()
	handler := s.handlers[job.ID]
	s.mu.RUnlock()
	if handler == nil {
		s.log.Warnf("no handler for job %s — marking failed", job.ID)
		job.Status = "failed"
		job.UpdatedAt = time.Now()
		if err := s.db.SaveJob(&job); err != nil {
			s.log.Errorf("save no-handler job state %s: %v", job.ID, err)
		}
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
	// Safeguard: Attempts should always be >= 1 when this is called,
	// but protect against underflow just in case.
	shift := 0
	if job.Attempts > 0 {
		shift = job.Attempts - 1
	}
	delay := base * time.Duration(1<<uint(shift))
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	// jitter up to 500ms
	jitterMs, err := rand.Int(rand.Reader, big.NewInt(500))
	if err != nil {
		return delay // fallback to no jitter if crypto/rand fails
	}
	jitter := time.Duration(jitterMs.Int64()) * time.Millisecond
	return delay + jitter
}

// Stop stops the scheduler and waits for running tasks to finish.
func (s *Scheduler) Stop() {
	close(s.quit)
	s.wg.Wait()
	// Close database to ensure all data is persisted
	if s.db != nil {
		s.db.Close()
	}
}
