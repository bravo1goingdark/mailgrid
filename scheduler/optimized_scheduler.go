package scheduler

import (
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/database"
)

// OptimizedScheduler is a simplified scheduler that wraps the base Scheduler.
// It maintains the same interface for backward compatibility but delegates
// to the simpler, more reliable base implementation.
type OptimizedScheduler struct {
	*Scheduler
	metrics *Metrics
}

// Metrics holds simple scheduler metrics
type Metrics struct {
	mu              sync.RWMutex
	jobsExecuted    int
	jobsFailed      int
	lastExecutionAt time.Time
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordJobExecuted increments the executed count
func (m *Metrics) RecordJobExecuted() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobsExecuted++
	m.lastExecutionAt = time.Now()
}

// RecordJobFailed increments the failed count
func (m *Metrics) RecordJobFailed() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobsFailed++
}

// GetJobsExecuted returns the number of executed jobs
func (m *Metrics) GetJobsExecuted() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobsExecuted
}

// GetJobsFailed returns the number of failed jobs
func (m *Metrics) GetJobsFailed() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobsFailed
}

// OptimizedConfig provides configuration for the scheduler
type OptimizedConfig struct {
	MaxConcurrency  int
	AdaptivePolling bool
	BaseInterval    time.Duration
	MaxInterval     time.Duration
	MetricsPort     int
}

// DefaultOptimizedConfig returns sensible defaults
func DefaultOptimizedConfig() OptimizedConfig {
	return OptimizedConfig{
		MaxConcurrency:  10,
		AdaptivePolling: true,
		BaseInterval:    1 * time.Second,
		MaxInterval:     30 * time.Second,
		MetricsPort:     8090,
	}
}

// DefaultOptimizedConfigWithPort returns config with specified metrics port
func DefaultOptimizedConfigWithPort(port int) OptimizedConfig {
	cfg := DefaultOptimizedConfig()
	cfg.MetricsPort = port
	return cfg
}

// NewOptimizedScheduler creates a scheduler (simplified wrapper around base Scheduler)
func NewOptimizedScheduler(db *database.BoltDBClient, log Logger, smtpCfg any, config OptimizedConfig) (*OptimizedScheduler, error) {
	scheduler := NewScheduler(db, log)
	return &OptimizedScheduler{
		Scheduler: scheduler,
		metrics:   NewMetrics(),
	}, nil
}

// GetMetrics returns the scheduler metrics
func (s *OptimizedScheduler) GetMetrics() *Metrics {
	return s.metrics
}

// CancelJob cancels a scheduled job
func (s *OptimizedScheduler) CancelJob(jobID string) bool {
	return s.Scheduler.CancelJob(jobID)
}

// ReattachHandlers reattaches handlers to existing jobs
func (s *OptimizedScheduler) ReattachHandlers(handler JobHandler) {
	s.Scheduler.ReattachHandlers(handler)
}
