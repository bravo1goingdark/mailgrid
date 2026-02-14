package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
)

// SchedulerManager provides automatic scheduler lifecycle management
// It automatically starts the scheduler when jobs are scheduled and manages cleanup
type SchedulerManager struct {
	mu            sync.RWMutex
	scheduler     *OptimizedScheduler
	isRunning     bool
	dbPath        string
	smtpConfig    config.SMTPConfig
	config        OptimizedConfig
	logger        Logger
	shutdownTimer *time.Timer
	shutdownDelay time.Duration
}

// ManagerConfig provides configuration for the scheduler manager
type ManagerConfig struct {
	DBPath          string
	SMTPConfig      config.SMTPConfig
	OptimizedConfig OptimizedConfig
	ShutdownDelay   time.Duration // Time to wait before auto-shutdown when no jobs
	AutoShutdown    bool          // Whether to auto-shutdown when idle
}

// DefaultManagerConfig returns sensible defaults for the manager
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		DBPath:          "scheduler.db",
		OptimizedConfig: DefaultOptimizedConfig(),
		ShutdownDelay:   5 * time.Minute,
		AutoShutdown:    true,
	}
}

// NewSchedulerManager creates a new scheduler manager
func NewSchedulerManager(config ManagerConfig) *SchedulerManager {
	return &SchedulerManager{
		dbPath:        config.DBPath,
		smtpConfig:    config.SMTPConfig,
		config:        config.OptimizedConfig,
		logger:        logger.New("scheduler-manager"),
		shutdownDelay: config.ShutdownDelay,
	}
}

// ensureStarted ensures the scheduler is running, starting it if necessary
func (sm *SchedulerManager) ensureStarted() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isRunning {
		// Cancel any pending shutdown
		if sm.shutdownTimer != nil {
			sm.shutdownTimer.Stop()
			sm.shutdownTimer = nil
		}
		return nil
	}

	// Start the scheduler
	db, err := database.NewDB(sm.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open scheduler database: %w", err)
	}

	scheduler, err := NewOptimizedScheduler(db, sm.logger, sm.smtpConfig, sm.config)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create optimized scheduler: %w", err)
	}

	sm.scheduler = scheduler
	sm.isRunning = true

	sm.logger.Infof("Scheduler started automatically with database: %s", sm.dbPath)

	return nil
}

// ScheduleJob schedules a job, automatically starting the scheduler if needed
func (sm *SchedulerManager) ScheduleJob(args types.CLIArgs, runAt time.Time, cronExpr, interval string, handler JobHandler) error {
	// Ensure scheduler is started
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Create and schedule the job
	job := NewJob(args, runAt, cronExpr, interval)

	scheduler := sm.scheduler
	if err := scheduler.AddJob(job, handler); err != nil {
		return fmt.Errorf("failed to add job to scheduler: %w", err)
	}

	sm.logger.Infof("Job %s scheduled successfully", job.ID)
	return nil
}

// ListJobs returns all scheduled jobs
func (sm *SchedulerManager) ListJobs() ([]types.Job, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isRunning {
		return nil, fmt.Errorf("scheduler not running")
	}

	return sm.scheduler.ListJobs()
}

// CancelJob cancels a scheduled job
func (sm *SchedulerManager) CancelJob(jobID string) error {
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	scheduler := sm.scheduler
	if !scheduler.CancelJob(jobID) {
		return fmt.Errorf("failed to cancel job %s", jobID)
	}

	return nil
}

// IsRunning returns whether the scheduler is currently running
func (sm *SchedulerManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isRunning
}

// Stop gracefully stops the scheduler and all associated services
func (sm *SchedulerManager) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Cancel any pending shutdown
	if sm.shutdownTimer != nil {
		sm.shutdownTimer.Stop()
		sm.shutdownTimer = nil
	}

	// Stop scheduler
	if sm.scheduler != nil {
		sm.scheduler.Stop()
		sm.scheduler = nil
	}

	sm.isRunning = false
	sm.logger.Infof("Scheduler manager stopped")
	return nil
}

// RunDaemon runs the scheduler as a daemon process
func (sm *SchedulerManager) RunDaemon() error {
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler daemon: %w", err)
	}

	sm.logger.Infof("Scheduler daemon running...")

	// Wait for context cancellation or stop signal
	// This is a blocking call that keeps the daemon running
	select {}
}

// Global manager instance for singleton access
var globalManager *SchedulerManager
var globalManagerMu sync.Mutex

// InitGlobalManager initializes the global scheduler manager singleton
func InitGlobalManager(config ManagerConfig) {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()

	if globalManager != nil {
		globalManager.Stop()
	}

	globalManager = NewSchedulerManager(config)
}

// GetGlobalManager returns the global scheduler manager singleton
// Returns nil if not initialized
func GetGlobalManager() *SchedulerManager {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()
	return globalManager
}
