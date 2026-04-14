package scheduler

import (
	"time"

	"github.com/bravo1goingdark/mailgrid/database"
)

// OptimizedScheduler is a simplified scheduler that wraps the base Scheduler.
// It maintains the same interface for backward compatibility but delegates
// to the simpler, more reliable base implementation.
type OptimizedScheduler struct {
	*Scheduler
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
	return &OptimizedScheduler{Scheduler: scheduler}, nil
}

// CancelJob cancels a scheduled job
func (s *OptimizedScheduler) CancelJob(jobID string) bool {
	return s.Scheduler.CancelJob(jobID)
}

// ReattachHandlers reattaches handlers to existing jobs
func (s *OptimizedScheduler) ReattachHandlers(handler JobHandler) {
	s.Scheduler.ReattachHandlers(handler)
}
