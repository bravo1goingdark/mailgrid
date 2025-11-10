package scheduler

import (
	"context"
	"fmt"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/metrics"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

// OptimizedScheduler provides high-performance job scheduling with:
// - Concurrent job execution
// - Adaptive polling intervals
// - Performance metrics
// - Template caching for email rendering
// - Circuit breaking and resilience
type OptimizedScheduler struct {
	db         *database.BoltDBClient
	log        Logger
	mu         sync.RWMutex
	handlers   map[string]JobHandler
	jobsCache  map[string]types.Job
	quit       chan struct{}
	wg         sync.WaitGroup
	instanceID string

	// Performance components
	smtpPool       *email.SMTPPool
	batchProcessor *email.BatchProcessor
	metrics        *metrics.Metrics
	resilience     *email.ResilienceManager
	templateCache  *email.TemplateCache

	// Optimization settings
	maxConcurrency  int
	adaptivePolling bool
	baseInterval    time.Duration
	maxInterval     time.Duration

	// Job execution pool
	jobQueue   chan types.Job
	workerPool chan struct{}
}

// OptimizedConfig provides configuration for the optimized scheduler
type OptimizedConfig struct {
	MaxConcurrency   int
	AdaptivePolling  bool
	BaseInterval     time.Duration
	MaxInterval      time.Duration
	PoolConfig       email.PoolConfig
	BatchConfig      email.BatchConfig
	MetricsPort      int
	EnableResilience bool
}

// DefaultOptimizedConfig returns sensible defaults
func DefaultOptimizedConfig() OptimizedConfig {
	return OptimizedConfig{
		MaxConcurrency:  10,
		AdaptivePolling: true,
		BaseInterval:    100 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		PoolConfig: email.PoolConfig{
			InitialSize:         5,
			MaxSize:             20,
			MaxIdleTime:         5 * time.Minute,
			MaxWaitTime:         30 * time.Second,
			HealthCheckInterval: 30 * time.Second,
		},
		BatchConfig: email.BatchConfig{
			MinBatchSize:     10,
			MaxBatchSize:     100,
			TargetLatency:    500 * time.Millisecond,
			AdaptationPeriod: 1 * time.Minute,
		},
		MetricsPort:      8090,
		EnableResilience: true,
	}
}

// NewOptimizedScheduler creates a high-performance scheduler
func NewOptimizedScheduler(db *database.BoltDBClient, log Logger, smtpCfg any, config OptimizedConfig) (*OptimizedScheduler, error) {
	s := &OptimizedScheduler{
		db:              db,
		log:             log,
		handlers:        make(map[string]JobHandler, 64),
		jobsCache:       make(map[string]types.Job, 128),
		quit:            make(chan struct{}),
		instanceID:      newInstanceID(),
		maxConcurrency:  config.MaxConcurrency,
		adaptivePolling: config.AdaptivePolling,
		baseInterval:    config.BaseInterval,
		maxInterval:     config.MaxInterval,
		jobQueue:        make(chan types.Job, config.MaxConcurrency*2),
		workerPool:      make(chan struct{}, config.MaxConcurrency),
	}

	// Initialize performance components
	s.metrics = metrics.NewMetrics()
	s.templateCache = email.NewTemplateCache(1*time.Hour, 100)

	// SMTP pool initialization is disabled - requires proper type assertion
	// This can be enabled when smtpCfg type is properly defined

	// Initialize resilience manager if enabled
	if config.EnableResilience {
		s.resilience = email.NewResilienceManager(5, 1*time.Minute, nil)
	}

	// Start metrics server
	if config.MetricsPort > 0 {
		metricsServer := metrics.NewServer(s.metrics, config.MetricsPort)
		go func() {
			if err := metricsServer.Start(); err != nil {
				s.log.Warnf("Failed to start metrics server: %v", err)
			}
		}()
	}

	// Warm cache from DB
	if jobs, err := db.LoadJobs(); err == nil {
		for _, j := range jobs {
			s.jobsCache[j.ID] = j
		}
	}

	// Initialize worker pool
	for i := 0; i < config.MaxConcurrency; i++ {
		s.workerPool <- struct{}{}
	}

	// Start background processes
	s.wg.Add(3)
	go s.optimizedDispatchLoop()
	go s.jobWorkerPool()
	go s.metricsCollector()

	return s, nil
}

// optimizedDispatchLoop uses adaptive polling and concurrent job processing
func (s *OptimizedScheduler) optimizedDispatchLoop() {
	defer s.wg.Done()

	currentInterval := s.baseInterval
	lastJobCount := 0

	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			jobs, err := s.db.LoadJobs()
			if err != nil {
				s.log.Errorf("load jobs: %v", err)
				s.metrics.RecordEmailFailed(err)
				continue
			}

			dueJobs := s.findDueJobs(jobs)

			// Adaptive polling: adjust interval based on workload
			if s.adaptivePolling {
				currentInterval = s.adaptPollingInterval(len(dueJobs), lastJobCount)
				ticker.Reset(currentInterval)
				lastJobCount = len(dueJobs)
			}

			// Queue due jobs for concurrent processing
			for _, job := range dueJobs {
				select {
				case s.jobQueue <- job:
					s.metrics.RecordBatch(1, 1.0) // Record job queued
				case <-time.After(100 * time.Millisecond):
					s.log.Warnf("Job queue full, skipping job %s", job.ID)
				}
			}
		}
	}
}

// findDueJobs finds all jobs that are due for execution
func (s *OptimizedScheduler) findDueJobs(jobs []types.Job) []types.Job {
	var dueJobs []types.Job
	now := time.Now()

	for _, j := range jobs {
		if j.Status != "pending" {
			continue
		}
		if now.Before(j.RunAt) {
			continue
		}

		// Try to acquire distributed lock
		locked, err := s.db.AcquireLock(j.ID, s.instanceID)
		if err != nil || !locked {
			continue
		}

		dueJobs = append(dueJobs, j)
	}

	return dueJobs
}

// adaptPollingInterval adjusts polling frequency based on workload
func (s *OptimizedScheduler) adaptPollingInterval(currentJobs, lastJobs int) time.Duration {
	if currentJobs > 0 || lastJobs > 0 {
		// High activity: poll more frequently
		return s.baseInterval
	}

	// Low activity: poll less frequently to save resources
	newInterval := s.baseInterval * 2
	if newInterval > s.maxInterval {
		newInterval = s.maxInterval
	}

	return newInterval
}

// jobWorkerPool processes jobs concurrently
func (s *OptimizedScheduler) jobWorkerPool() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		case job := <-s.jobQueue:
			// Get worker from pool (blocking if all busy)
			<-s.workerPool

			// Execute job in goroutine
			go func(j types.Job) {
				defer func() {
					// Return worker to pool
					s.workerPool <- struct{}{}
					// Always release lock
					_ = s.db.ReleaseLock(j.ID, s.instanceID)
				}()

				s.executeJobWithMetrics(j)
			}(job)
		}
	}
}

// executeJobWithMetrics executes a job with performance tracking
func (s *OptimizedScheduler) executeJobWithMetrics(job types.Job) {
	start := time.Now()

	// Update job status
	job.Status = "running"
	job.UpdatedAt = time.Now()

	s.mu.RLock()
	handler := s.handlers[job.ID]
	s.mu.RUnlock()

	if handler == nil {
		s.log.Warnf("no handler for job %s", job.ID)
		s.metrics.RecordEmailFailed(fmt.Errorf("no handler"))
		return
	}

	// Execute with resilience if available
	var err error
	if s.resilience != nil {
		err = s.resilience.Execute(context.Background(), func() error {
			return handler(job)
		})
	} else {
		err = handler(job)
	}

	duration := time.Since(start)

	if err != nil {
		s.metrics.RecordEmailFailed(err)
		s.handleJobFailure(job, err)
	} else {
		s.metrics.RecordEmailSent(duration)
		s.handleJobSuccess(job)
	}
}

// handleJobFailure handles job execution failures with retry logic
func (s *OptimizedScheduler) handleJobFailure(job types.Job, err error) {
	job.Attempts++

	if job.Attempts < job.MaxAttempts {
		// Reschedule with backoff
		delay := computeBackoff(job)
		job.Status = "pending"
		job.RunAt = time.Now().Add(delay)
		job.NextRunAt = job.RunAt
		job.UpdatedAt = time.Now()

		if saveErr := s.db.SaveJob(&job); saveErr != nil {
			s.log.Errorf("save job retry state %s: %v", job.ID, saveErr)
		} else {
			s.mu.Lock()
			s.jobsCache[job.ID] = job
			s.mu.Unlock()
		}

		s.log.Warnf("job %s failed (attempt %d/%d), retry in %v: %v",
			job.ID, job.Attempts, job.MaxAttempts, delay, err)
		return
	}

	// Exhausted retries
	job.Status = "failed"
	job.UpdatedAt = time.Now()

	if saveErr := s.db.SaveJob(&job); saveErr != nil {
		s.log.Errorf("save job failed state %s: %v", job.ID, saveErr)
	}

	s.log.Errorf("job %s permanently failed: %v", job.ID, err)
}

// handleJobSuccess handles successful job execution and rescheduling
func (s *OptimizedScheduler) handleJobSuccess(job types.Job) {
	job.Status = "done"
	job.LastRunAt = time.Now()
	job.UpdatedAt = job.LastRunAt

	// Reschedule if needed
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

// metricsCollector periodically reports scheduler metrics
func (s *OptimizedScheduler) metricsCollector() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			// Report scheduler-specific metrics
			s.mu.RLock()
			totalJobs := len(s.jobsCache)
			s.mu.RUnlock()

			queueSize := len(s.jobQueue)
			activeWorkers := s.maxConcurrency - len(s.workerPool)

			s.log.Infof("Scheduler metrics: %d total jobs, %d queued, %d active workers",
				totalJobs, queueSize, activeWorkers)
		}
	}
}

// AddJob registers a job with enhanced error handling
func (s *OptimizedScheduler) AddJob(job types.Job, handler JobHandler) error {
	if handler == nil {
		return fmt.Errorf("nil handler")
	}

	// Persist with metrics
	start := time.Now()
	if err := s.db.SaveJob(&job); err != nil {
		s.metrics.RecordEmailFailed(err)
		return fmt.Errorf("save job: %w", err)
	}

	// Track handler and cache
	s.mu.Lock()
	s.handlers[job.ID] = handler
	s.jobsCache[job.ID] = job
	s.mu.Unlock()

	s.metrics.RecordEmailSent(time.Since(start))
	return nil
}

// CancelJob cancels a job with metrics
func (s *OptimizedScheduler) CancelJob(jobID string) bool {
	job, err := s.db.GetJob(jobID)
	if err != nil {
		s.log.Warnf("cancel: job not found: %s", jobID)
		s.metrics.RecordEmailFailed(err)
		return false
	}

	job.Status = "cancelled"
	job.UpdatedAt = time.Now()

	if err := s.db.SaveJob(job); err != nil {
		s.log.Errorf("cancel: save job %s: %v", jobID, err)
		s.metrics.RecordEmailFailed(err)
		return false
	}

	s.mu.Lock()
	s.jobsCache[jobID] = *job
	s.mu.Unlock()

	return true
}

// ListJobs returns all jobs from cache
func (s *OptimizedScheduler) ListJobs() ([]types.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]types.Job, 0, len(s.jobsCache))
	for _, j := range s.jobsCache {
		res = append(res, j)
	}

	return res, nil
}

// GetMetrics returns current performance metrics
func (s *OptimizedScheduler) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// Stop gracefully stops the optimized scheduler
func (s *OptimizedScheduler) Stop() {
	close(s.quit)
	s.wg.Wait()

	// Close performance components
	if s.smtpPool != nil {
		s.smtpPool.Close()
	}
	if s.templateCache != nil {
		s.templateCache.Clear()
	}
}

// ReattachHandlers sets handlers for existing jobs
func (s *OptimizedScheduler) ReattachHandlers(handler JobHandler) {
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
