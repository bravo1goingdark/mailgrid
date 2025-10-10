package email

import (
	"log"
	"sync"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/monitor"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/bravo1goingdark/mailgrid/webhook"
)

// Task represents an email send job with recipient data.
type Task struct {
	Recipient   parser.Recipient
	Subject     string
	Body        string
	Retries     int
	Attachments []string
	CC          []string
	BCC         []string
}

type worker struct {
	ID            int
	TaskQueue     <-chan Task
	RetryChan     chan<- Task
	Config        config.SMTPConfig
	Wg            *sync.WaitGroup
	RetryWg       *sync.WaitGroup
	BatchSize     int
	Monitor       monitor.Monitor
	OffsetTracker *OffsetTracker
}

// StartDispatcher spawns workers and processes email tasks with retries and batch-mode dispatch.
func StartDispatcher(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int) {
	StartDispatcherWithMonitor(tasks, cfg, concurrency, batchSize, monitor.NewNoOpMonitor(), nil)
}

// StartDispatcherWithMonitor spawns workers with monitoring support and optional offset tracking.
func StartDispatcherWithMonitor(tasks []Task, cfg config.SMTPConfig, concurrency int, batchSize int, mon monitor.Monitor, offsetTracker *OffsetTracker) {
	taskChan := make(chan Task, 1000)
	retryChan := make(chan Task, 500)

	var wg sync.WaitGroup
	var retryWg sync.WaitGroup

	// Initialize all recipients as pending in monitor
	for _, task := range tasks {
		mon.UpdateRecipientStatus(task.Recipient.Email, monitor.StatusPending, 0, "")
	}

	// Spawn workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go startWorker(worker{
			ID:            i + 1,
			TaskQueue:     taskChan,
			RetryChan:     retryChan,
			Config:        cfg,
			Wg:            &wg,
			RetryWg:       &retryWg,
			BatchSize:     batchSize,
			Monitor:       mon,
			OffsetTracker: offsetTracker,
		})
	}

	// Dispatch initial tasks
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// Handle retries
	go func() {
		for task := range retryChan {
			if task.Retries > 0 {
				retryWg.Add(1)
				go func(t Task) {
					defer retryWg.Done()
					taskChan <- t
				}(task)
			} else {
				log.Printf("Permanent failure: %s", task.Recipient.Email)
			}
		}
	}()

	wg.Wait()
	close(retryChan)
	retryWg.Wait()
}

// CampaignConfig represents configuration for a campaign dispatcher
type CampaignConfig struct {
	JobID             string
	TotalRecipients   int
	CSVFile           string
	SheetURL          string
	TemplateFile      string
	ConcurrentWorkers int
	WebhookURL        string
	Monitor           monitor.Monitor
}

// CampaignMetrics tracks email campaign statistics
type CampaignMetrics struct {
	jobID                string
	totalRecipients      int
	csvFile              string
	sheetURL             string
	templateFile         string
	webhookURL           string
	successfulDeliveries int
	failedDeliveries     int
	retryCount           int
	startTime            time.Time
	endTime              time.Time
	mu                   sync.RWMutex
}

// PooledDispatcher manages email dispatching with connection pooling
type PooledDispatcher struct {
	pool      *SMTPPool
	processor *BatchProcessor
	webhook   *webhook.Client
	monitor   monitor.Monitor
	metrics   *CampaignMetrics
	mu        sync.RWMutex
}

// NewPooledDispatcher creates a new pooled dispatcher
func NewPooledDispatcher(cfg config.SMTPConfig, poolSize int, batchSize int) (*PooledDispatcher, error) {
	poolConfig := PoolConfig{
		InitialSize:         poolSize,
		MaxSize:             poolSize * 2,
		MaxIdleTime:         5 * time.Minute,
		MaxWaitTime:         30 * time.Second,
		HealthCheckInterval: 30 * time.Second,
	}

	pool, err := NewSMTPPool(cfg, poolConfig)
	if err != nil {
		return nil, err
	}

	batchConfig := BatchConfig{
		MinBatchSize:     batchSize,
		MaxBatchSize:     batchSize * 10,
		TargetLatency:    500 * time.Millisecond,
		AdaptationPeriod: 1 * time.Minute,
	}

	processor := NewBatchProcessor(pool, batchConfig)
	webhookClient := webhook.NewClient()
	mon := monitor.NewNoOpMonitor()

	metrics := &CampaignMetrics{
		startTime: time.Now(),
	}

	return &PooledDispatcher{
		pool:      pool,
		processor: processor,
		webhook:   webhookClient,
		monitor:   mon,
		metrics:   metrics,
	}, nil
}

// NewPooledDispatcherWithCampaign creates a new pooled dispatcher with campaign configuration
func NewPooledDispatcherWithCampaign(cfg config.SMTPConfig, poolSize int, batchSize int, campaign CampaignConfig) (*PooledDispatcher, error) {
	dispatcher, err := NewPooledDispatcher(cfg, poolSize, batchSize)
	if err != nil {
		return nil, err
	}

	// Set campaign configuration
	dispatcher.metrics.jobID = campaign.JobID
	dispatcher.metrics.totalRecipients = campaign.TotalRecipients
	dispatcher.metrics.csvFile = campaign.CSVFile
	dispatcher.metrics.sheetURL = campaign.SheetURL
	dispatcher.metrics.templateFile = campaign.TemplateFile
	dispatcher.metrics.webhookURL = campaign.WebhookURL

	if campaign.Monitor != nil {
		dispatcher.monitor = campaign.Monitor
	}

	return dispatcher, nil
}

// Close closes the dispatcher and its resources
func (d *PooledDispatcher) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.metrics != nil {
		d.metrics.endTime = time.Now()
	}

	if d.pool != nil {
		return d.pool.Close()
	}
	return nil
}

// GetMetrics returns the current campaign metrics
func (d *PooledDispatcher) GetMetrics() *CampaignMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.metrics
}

// RecordSuccess increments the successful delivery counter
func (d *PooledDispatcher) RecordSuccess() {
	d.metrics.mu.Lock()
	defer d.metrics.mu.Unlock()
	d.metrics.successfulDeliveries++
}

// RecordFailure increments the failed delivery counter
func (d *PooledDispatcher) RecordFailure() {
	d.metrics.mu.Lock()
	defer d.metrics.mu.Unlock()
	d.metrics.failedDeliveries++
}

// RecordSuccessWithDetails records a successful delivery with details
func (d *PooledDispatcher) RecordSuccessWithDetails(email string, duration time.Duration) {
	d.RecordSuccess()
	d.monitor.UpdateRecipientStatus(email, monitor.StatusSent, duration, "")
}

// RecordFailureWithDetails records a failed delivery with details
func (d *PooledDispatcher) RecordFailureWithDetails(email string, duration time.Duration, errorMsg string) {
	d.RecordFailure()
	d.monitor.UpdateRecipientStatus(email, monitor.StatusFailed, duration, errorMsg)
}

// RecordRetryWithDetails records a retry attempt with details
func (d *PooledDispatcher) RecordRetryWithDetails(email string, duration time.Duration, reason string) {
	d.metrics.mu.Lock()
	d.metrics.retryCount++
	d.metrics.mu.Unlock()
	d.monitor.UpdateRecipientStatus(email, monitor.StatusRetry, duration, reason)
}
