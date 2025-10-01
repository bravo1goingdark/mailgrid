package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/metrics"
)

// SchedulerManager provides automatic scheduler lifecycle management
// It automatically starts the scheduler when jobs are scheduled and manages cleanup
type SchedulerManager struct {
	mu               sync.RWMutex
	scheduler        *OptimizedScheduler
	isRunning        bool
	dbPath           string
	smtpConfig       config.SMTPConfig
	config           OptimizedConfig
	logger           Logger
	shutdownTimer    *time.Timer
	shutdownDelay    time.Duration
	metricsServer    *metrics.Server
}

// ManagerConfig provides configuration for the scheduler manager
type ManagerConfig struct {
	DBPath           string
	SMTPConfig       config.SMTPConfig
	OptimizedConfig  OptimizedConfig
	ShutdownDelay    time.Duration // Time to wait before auto-shutdown when no jobs
	AutoShutdown     bool          // Whether to auto-shutdown when idle
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
	
	// Start metrics server if enabled
	if sm.config.MetricsPort > 0 {
		sm.metricsServer = metrics.NewServer(scheduler.GetMetrics(), sm.config.MetricsPort)
		go func() {
			if err := sm.metricsServer.Start(); err != nil {
				sm.logger.Warnf("Failed to start metrics server: %v", err)
			}
		}()
	}
	
	sm.logger.Infof("Scheduler started automatically with database: %s", sm.dbPath)
	sm.logger.Infof("Metrics server available at: http://localhost:%d/metrics", sm.config.MetricsPort)
	
	// Start monitoring for auto-shutdown
	go sm.monitorActivity()
	
	return nil
}

// monitorActivity monitors scheduler activity and handles auto-shutdown
func (sm *SchedulerManager) monitorActivity() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if !sm.isRunning {
				return
			}
			
			jobs, err := sm.scheduler.ListJobs()
			if err != nil {
				continue
			}
			
			// Check if there are any pending jobs
			hasPendingJobs := false
			now := time.Now()
			for _, job := range jobs {
				if job.Status == "pending" || job.Status == "running" {
					hasPendingJobs = true
					break
				}
				// Also consider jobs that might be scheduled in the near future
				if job.Status == "done" && !job.NextRunAt.IsZero() && job.NextRunAt.After(now) {
					hasPendingJobs = true
					break
				}
			}
			
			if !hasPendingJobs {
				sm.scheduleShutdown()
			} else {
				// Cancel any pending shutdown if jobs are active
				sm.mu.Lock()
				if sm.shutdownTimer != nil {
					sm.shutdownTimer.Stop()
					sm.shutdownTimer = nil
				}
				sm.mu.Unlock()
			}
		}
	}
}

// scheduleShutdown schedules an automatic shutdown after a delay
func (sm *SchedulerManager) scheduleShutdown() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Don't schedule if already scheduled or if auto-shutdown is disabled
	if sm.shutdownTimer != nil || sm.shutdownDelay <= 0 {
		return
	}
	
	sm.shutdownTimer = time.AfterFunc(sm.shutdownDelay, func() {
		sm.logger.Infof("Auto-shutting down scheduler after %v of inactivity", sm.shutdownDelay)
		sm.Stop()
	})
	
	sm.logger.Infof("Scheduled auto-shutdown in %v (no pending jobs)", sm.shutdownDelay)
}

// ScheduleJob schedules a job, automatically starting the scheduler if needed
func (sm *SchedulerManager) ScheduleJob(args types.CLIArgs, runAt time.Time, cronExpr, interval string, handler JobHandler) error {
	// Ensure scheduler is started
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	
	// Create and schedule the job
	job := NewJob(args, runAt, cronExpr, interval)
	
	sm.mu.RLock()
	scheduler := sm.scheduler
	sm.mu.RUnlock()
	
	if err := scheduler.AddJob(job, handler); err != nil {
		return fmt.Errorf("failed to add job: %w", err)
	}
	
	sm.logger.Infof("Job %s scheduled successfully", job.ID)
	return nil
}

// ListJobs returns all scheduled jobs
func (sm *SchedulerManager) ListJobs() ([]types.Job, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if !sm.isRunning {
		// If scheduler is not running, create a temporary connection to read jobs
		db, err := database.NewDB(sm.dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()
		
		return db.LoadJobs()
	}
	
	return sm.scheduler.ListJobs()
}

// CancelJob cancels a scheduled job
func (sm *SchedulerManager) CancelJob(jobID string) error {
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	
	sm.mu.RLock()
	scheduler := sm.scheduler
	sm.mu.RUnlock()
	
	if !scheduler.CancelJob(jobID) {
		return fmt.Errorf("job not found or couldn't be cancelled: %s", jobID)
	}
	
	sm.logger.Infof("Job %s cancelled successfully", jobID)
	return nil
}

// GetMetrics returns current scheduler metrics
func (sm *SchedulerManager) GetMetrics() *metrics.Metrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if !sm.isRunning {
		return nil
	}
	
	return sm.scheduler.GetMetrics()
}

// IsRunning returns whether the scheduler is currently running
func (sm *SchedulerManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isRunning
}

// Stop gracefully stops the scheduler and all associated services
func (sm *SchedulerManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.isRunning {
		return
	}
	
	// Cancel shutdown timer if active
	if sm.shutdownTimer != nil {
		sm.shutdownTimer.Stop()
		sm.shutdownTimer = nil
	}
	
	// Stop metrics server
	if sm.metricsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		sm.metricsServer.Stop(ctx)
		sm.metricsServer = nil
	}
	
	// Stop scheduler
	if sm.scheduler != nil {
		sm.scheduler.Stop()
		sm.scheduler = nil
	}
	
	sm.isRunning = false
	sm.logger.Infof("Scheduler stopped")
}

// RunDaemon runs the scheduler as a daemon process
func (sm *SchedulerManager) RunDaemon(ctx context.Context) error {
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler daemon: %w", err)
	}
	
	sm.logger.Infof("Scheduler daemon started. Press Ctrl+C to stop.")
	sm.logger.Infof("Database: %s", sm.dbPath)
	if sm.config.MetricsPort > 0 {
		sm.logger.Infof("Metrics: http://localhost:%d/metrics", sm.config.MetricsPort)
		sm.logger.Infof("Health: http://localhost:%d/health", sm.config.MetricsPort)
	}
	
	// Disable auto-shutdown in daemon mode
	sm.mu.Lock()
	if sm.shutdownTimer != nil {
		sm.shutdownTimer.Stop()
		sm.shutdownTimer = nil
	}
	sm.mu.Unlock()
	
	// Wait for context cancellation
	<-ctx.Done()
	
	sm.logger.Infof("Shutting down scheduler daemon...")
	sm.Stop()
	
	return nil
}

// AttachDefaultHandler attaches a default handler for job execution
func (sm *SchedulerManager) AttachDefaultHandler(handler JobHandler) error {
	if err := sm.ensureStarted(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	
	sm.mu.RLock()
	scheduler := sm.scheduler
	sm.mu.RUnlock()
	
	scheduler.ReattachHandlers(handler)
	sm.logger.Infof("Default job handler attached")
	
	return nil
}

// Global instance for easy access
var globalManager *SchedulerManager

// InitGlobalManager initializes the global scheduler manager
func InitGlobalManager(config ManagerConfig) {
	globalManager = NewSchedulerManager(config)
}

// GetGlobalManager returns the global scheduler manager
func GetGlobalManager() *SchedulerManager {
	if globalManager == nil {
		// Initialize with defaults if not set
		globalManager = NewSchedulerManager(DefaultManagerConfig())
	}
	return globalManager
}

// ScheduleGlobal schedules a job using the global manager
func ScheduleGlobal(args types.CLIArgs, runAt time.Time, cronExpr, interval string, handler JobHandler) error {
	return GetGlobalManager().ScheduleJob(args, runAt, cronExpr, interval, handler)
}

// StopGlobal stops the global scheduler manager
func StopGlobal() {
	if globalManager != nil {
		globalManager.Stop()
	}
}