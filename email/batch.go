package email

import (
	"context"
	"sync"
	"time"
)

// BatchMetrics tracks performance metrics for batch processing
type BatchMetrics struct {
	mu sync.RWMutex

	// Average processing time per email
	avgProcessingTime time.Duration
	// Number of samples for the average
	samples int64

	// Success/failure counts
	successes int64
	failures  int64

	// Batch size metrics
	minSuccessfulBatch int
	maxSuccessfulBatch int
	optimalBatchSize   int

	// Time window metrics
	windowStart time.Time
	windowStats map[time.Time]batchStats
}

type batchStats struct {
	size      int
	duration  time.Duration
	successes int
	failures  int
}

// BatchProcessor handles adaptive batch processing of emails
type BatchProcessor struct {
	// Configuration
	minBatchSize     int
	maxBatchSize     int
	targetLatency    time.Duration
	adaptationPeriod time.Duration

	// Metrics
	metrics *BatchMetrics

	// State
	mu            sync.RWMutex
	currentBatch  []Task
	batchSize     int
	lastAdapted   time.Time
	flushTrigger  chan struct{}
	processorPool *SMTPPool
}

// NewBatchProcessor creates a new batch processor with given configuration
func NewBatchProcessor(pool *SMTPPool, config BatchConfig) *BatchProcessor {
	if config.MinBatchSize <= 0 {
		config.MinBatchSize = 10
	}
	if config.MaxBatchSize <= 0 {
		config.MaxBatchSize = 1000
	}
	if config.TargetLatency <= 0 {
		config.TargetLatency = 500 * time.Millisecond
	}
	if config.AdaptationPeriod <= 0 {
		config.AdaptationPeriod = 1 * time.Minute
	}

	return &BatchProcessor{
		minBatchSize:     config.MinBatchSize,
		maxBatchSize:     config.MaxBatchSize,
		targetLatency:    config.TargetLatency,
		adaptationPeriod: config.AdaptationPeriod,
		metrics: &BatchMetrics{
			windowStats:        make(map[time.Time]batchStats),
			windowStart:        time.Now(),
			optimalBatchSize:  config.MinBatchSize,
			minSuccessfulBatch: config.MaxBatchSize,
			maxSuccessfulBatch: config.MinBatchSize,
		},
		batchSize:     config.MinBatchSize,
		flushTrigger:  make(chan struct{}, 1),
		processorPool: pool,
		currentBatch:  make([]Task, 0, config.MinBatchSize),
	}
}

// Add adds a task to the current batch
func (p *BatchProcessor) Add(task Task) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentBatch = append(p.currentBatch, task)

	// If batch is full, trigger flush
	if len(p.currentBatch) >= p.batchSize {
		select {
		case p.flushTrigger <- struct{}{}:
		default:
		}
	}
}

// Start starts the batch processor
func (p *BatchProcessor) Start(ctx context.Context) {
	// Start batch processing loop
	go p.processingLoop(ctx)

	// Start adaptation loop
	go p.adaptationLoop(ctx)
}

func (p *BatchProcessor) processingLoop(ctx context.Context) {
	timer := time.NewTimer(p.targetLatency)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-p.flushTrigger:
			p.flush()
			timer.Reset(p.targetLatency)

		case <-timer.C:
			p.mu.Lock()
			if len(p.currentBatch) > 0 {
				p.flush()
			}
			p.mu.Unlock()
			timer.Reset(p.targetLatency)
		}
	}
}

func (p *BatchProcessor) flush() {
	if len(p.currentBatch) == 0 {
		return
	}

	batch := p.currentBatch
	p.currentBatch = make([]Task, 0, p.batchSize)
	p.mu.Unlock()

	start := time.Now()
	successes, failures := p.processBatch(batch)
	duration := time.Since(start)

	// Update metrics
	p.metrics.mu.Lock()
	p.metrics.windowStats[start] = batchStats{
		size:      len(batch),
		duration:  duration,
		successes: successes,
		failures:  failures,
	}
	p.metrics.avgProcessingTime = time.Duration(
		(int64(p.metrics.avgProcessingTime)*p.metrics.samples + int64(duration)) /
			(p.metrics.samples + 1),
	)
	p.metrics.samples++
	p.metrics.successes += int64(successes)
	p.metrics.failures += int64(failures)

	if successes > failures && len(batch) > p.metrics.maxSuccessfulBatch {
		p.metrics.maxSuccessfulBatch = len(batch)
	}
	if successes > failures && len(batch) < p.metrics.minSuccessfulBatch {
		p.metrics.minSuccessfulBatch = len(batch)
	}
	p.metrics.mu.Unlock()

	p.mu.Lock()
}

func (p *BatchProcessor) processBatch(batch []Task) (successes, failures int) {
	// Get connection from pool
	client, err := p.processorPool.Get(context.Background())
	if err != nil {
		return 0, len(batch)
	}
	defer p.processorPool.Put(client)

	for _, task := range batch {
		if err := SendWithClient(client, p.processorPool.smtpCfg, task); err != nil {
			failures++
		} else {
			successes++
		}
	}

	return successes, failures
}

func (p *BatchProcessor) adaptationLoop(ctx context.Context) {
	ticker := time.NewTicker(p.adaptationPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.adapt()
		}
	}
}

func (p *BatchProcessor) adapt() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	// Skip if not enough data
	if p.metrics.samples < 10 {
		return
	}

	// Calculate success rate and average latency for current batch size
	var totalSuccesses, totalFailures int
	var totalLatency time.Duration
	var batchCount int

	now := time.Now()
	windowStart := now.Add(-p.adaptationPeriod)

	for timestamp, stats := range p.metrics.windowStats {
		if timestamp.Before(windowStart) {
			delete(p.metrics.windowStats, timestamp)
			continue
		}
		totalSuccesses += stats.successes
		totalFailures += stats.failures
		totalLatency += stats.duration
		batchCount++
	}

	if batchCount == 0 {
		return
	}

	avgLatency := totalLatency / time.Duration(batchCount)
	successRate := float64(totalSuccesses) / float64(totalSuccesses+totalFailures)

	// Adjust batch size based on metrics
	p.mu.Lock()
	defer p.mu.Unlock()

	if successRate > 0.95 && avgLatency < p.targetLatency {
		// Increase batch size if performing well
		newSize := int(float64(p.batchSize) * 1.2)
		if newSize > p.maxBatchSize {
			newSize = p.maxBatchSize
		}
		p.batchSize = newSize
	} else if successRate < 0.8 || avgLatency > p.targetLatency*2 {
		// Decrease batch size if struggling
		newSize := int(float64(p.batchSize) * 0.8)
		if newSize < p.minBatchSize {
			newSize = p.minBatchSize
		}
		p.batchSize = newSize
	}

	// Record optimal batch size if current performance is good
	if successRate > 0.9 && avgLatency <= p.targetLatency {
		p.metrics.optimalBatchSize = p.batchSize
	}
}

// GetMetrics returns current batch processing metrics
func (p *BatchProcessor) GetMetrics() BatchStats {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return BatchStats{
		AvgProcessingTime:   p.metrics.avgProcessingTime,
		Successes:          p.metrics.successes,
		Failures:           p.metrics.failures,
		OptimalBatchSize:   p.metrics.optimalBatchSize,
		MinSuccessfulBatch: p.metrics.minSuccessfulBatch,
		MaxSuccessfulBatch: p.metrics.maxSuccessfulBatch,
		CurrentBatchSize:   p.batchSize,
	}
}

// BatchStats provides metrics about batch processing
type BatchStats struct {
	AvgProcessingTime   time.Duration
	Successes          int64
	Failures           int64
	OptimalBatchSize   int
	MinSuccessfulBatch int
	MaxSuccessfulBatch int
	CurrentBatchSize   int
}

// BatchConfig provides configuration for batch processing
type BatchConfig struct {
	MinBatchSize      int
	MaxBatchSize      int
	TargetLatency     time.Duration
	AdaptationPeriod  time.Duration
}