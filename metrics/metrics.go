package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Metrics collects and exposes performance metrics for email sending
type Metrics struct {
	mu sync.RWMutex

	// Email statistics
	TotalEmailsSent     uint64
	TotalEmailsFailed   uint64
	TotalAttachmentSize uint64
	AvgDeliveryTime     time.Duration
	deliveryTimesamples uint64

	// Connection metrics
	ActiveConnections  int64
	TotalConnections   uint64
	ConnectionErrors   uint64
	ConnectionTimeouts uint64

	// Batch metrics
	BatchesProcessed uint64
	AvgBatchSize     float64
	BatchSuccessRate float64
	batchSizeSum     uint64
	batchCount       uint64

	// Rate limiting
	ThrottleEvents   uint64
	CurrentRateLimit int64
	RateLimitHits    uint64

	// Template metrics
	TemplateCacheHits   uint64
	TemplateCacheMisses uint64
	TemplateCacheSize   int64

	// Error tracking
	ErrorCounts     map[string]uint64
	LastError       time.Time
	ConsecutiveErrs uint64

	// Performance tracking
	startTime       time.Time
	lastMinuteStats minuteStats
	hourlyStats     []minuteStats
}

type minuteStats struct {
	timestamp    time.Time
	emailsSent   uint64
	emailsFailed uint64
	avgLatency   time.Duration
	errorCount   uint64
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	m := &Metrics{
		startTime:   time.Now(),
		ErrorCounts: make(map[string]uint64),
		hourlyStats: make([]minuteStats, 60), // Keep last hour
	}

	// Start background stats collection
	go m.collectStats()

	return m
}

// RecordEmailSent records a successful email delivery
func (m *Metrics) RecordEmailSent(duration time.Duration) {
	atomic.AddUint64(&m.TotalEmailsSent, 1)

	// Update average delivery time
	samples := atomic.AddUint64(&m.deliveryTimesamples, 1)
	current := time.Duration(atomic.LoadUint64((*uint64)(unsafe.Pointer(&m.AvgDeliveryTime))))
	newAvg := time.Duration((int64(current)*int64(samples-1) + int64(duration)) / int64(samples))
	atomic.StoreUint64((*uint64)(unsafe.Pointer(&m.AvgDeliveryTime)), uint64(newAvg))
}

// RecordEmailFailed records a failed email delivery
func (m *Metrics) RecordEmailFailed(err error) {
	atomic.AddUint64(&m.TotalEmailsFailed, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	errStr := err.Error()
	m.ErrorCounts[errStr]++
	m.LastError = time.Now()
	m.ConsecutiveErrs++
}

// RecordBatch records batch processing metrics
func (m *Metrics) RecordBatch(size int, success float64) {
	atomic.AddUint64(&m.BatchesProcessed, 1)
	atomic.AddUint64(&m.batchSizeSum, uint64(size))
	atomic.AddUint64(&m.batchCount, 1)

	count := atomic.LoadUint64(&m.batchCount)
	sum := atomic.LoadUint64(&m.batchSizeSum)
	if count > 0 {
		m.AvgBatchSize = float64(sum) / float64(count)
	}

	// Update success rate with exponential moving average
	alpha := 0.1 // Smoothing factor
	m.BatchSuccessRate = m.BatchSuccessRate*(1-alpha) + success*alpha
}

// RecordConnection records connection metrics
func (m *Metrics) RecordConnection(success bool) {
	if success {
		atomic.AddUint64(&m.TotalConnections, 1)
		atomic.AddInt64(&m.ActiveConnections, 1)
	} else {
		atomic.AddUint64(&m.ConnectionErrors, 1)
	}
}

// RecordConnectionClosed records a closed connection
func (m *Metrics) RecordConnectionClosed() {
	atomic.AddInt64(&m.ActiveConnections, -1)
}

// RecordTemplateCache records template cache metrics
func (m *Metrics) RecordTemplateCache(hit bool, size int64) {
	if hit {
		atomic.AddUint64(&m.TemplateCacheHits, 1)
	} else {
		atomic.AddUint64(&m.TemplateCacheMisses, 1)
	}
	atomic.StoreInt64(&m.TemplateCacheSize, size)
}

// RecordThrottle records rate limiting events
func (m *Metrics) RecordThrottle(limit int64) {
	atomic.AddUint64(&m.ThrottleEvents, 1)
	atomic.StoreInt64(&m.CurrentRateLimit, limit)
}

// GetStats returns current metrics as a JSON string
func (m *Metrics) GetStats() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := struct {
		Uptime            time.Duration     `json:"uptime"`
		EmailsSent        uint64            `json:"emails_sent"`
		EmailsFailed      uint64            `json:"emails_failed"`
		AvgDeliveryTime   time.Duration     `json:"avg_delivery_time"`
		ActiveConnections int64             `json:"active_connections"`
		BatchesProcessed  uint64            `json:"batches_processed"`
		AvgBatchSize      float64           `json:"avg_batch_size"`
		BatchSuccessRate  float64           `json:"batch_success_rate"`
		ErrorCounts       map[string]uint64 `json:"error_counts"`
		LastError         time.Time         `json:"last_error"`
		TemplateCacheHits uint64            `json:"template_cache_hits"`
		ThrottleEvents    uint64            `json:"throttle_events"`
		CurrentRateLimit  int64             `json:"current_rate_limit"`
	}{
		Uptime:            time.Since(m.startTime),
		EmailsSent:        atomic.LoadUint64(&m.TotalEmailsSent),
		EmailsFailed:      atomic.LoadUint64(&m.TotalEmailsFailed),
		AvgDeliveryTime:   time.Duration(atomic.LoadUint64((*uint64)(unsafe.Pointer(&m.AvgDeliveryTime)))),
		ActiveConnections: atomic.LoadInt64(&m.ActiveConnections),
		BatchesProcessed:  atomic.LoadUint64(&m.BatchesProcessed),
		AvgBatchSize:      m.AvgBatchSize,
		BatchSuccessRate:  m.BatchSuccessRate,
		ErrorCounts:       m.ErrorCounts,
		LastError:         m.LastError,
		TemplateCacheHits: atomic.LoadUint64(&m.TemplateCacheHits),
		ThrottleEvents:    atomic.LoadUint64(&m.ThrottleEvents),
		CurrentRateLimit:  atomic.LoadInt64(&m.CurrentRateLimit),
	}

	bytes, _ := json.MarshalIndent(stats, "", "  ")
	return string(bytes)
}

// ServeHTTP implements http.Handler for metrics endpoint
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, m.GetStats())
}

func (m *Metrics) collectStats() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		stats := minuteStats{
			timestamp:    time.Now(),
			emailsSent:   atomic.LoadUint64(&m.TotalEmailsSent),
			emailsFailed: atomic.LoadUint64(&m.TotalEmailsFailed),
			avgLatency:   time.Duration(atomic.LoadUint64((*uint64)(unsafe.Pointer(&m.AvgDeliveryTime)))),
		}

		m.mu.Lock()
		// Rotate hourly stats
		copy(m.hourlyStats[1:], m.hourlyStats)
		m.hourlyStats[0] = stats

		// Calculate delta from last minute
		delta := minuteStats{
			emailsSent:   stats.emailsSent - m.lastMinuteStats.emailsSent,
			emailsFailed: stats.emailsFailed - m.lastMinuteStats.emailsFailed,
		}
		m.lastMinuteStats = stats
		m.mu.Unlock()

		// Log significant changes
		if delta.emailsSent > 1000 || delta.emailsFailed > 100 {
			log.Printf("High traffic detected: %d sent, %d failed in last minute",
				delta.emailsSent, delta.emailsFailed)
		}
	}
}
