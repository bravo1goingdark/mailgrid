package metrics

import (
	"context"
	"expvar"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Metrics holds application metrics
type Metrics struct {
	mu                    sync.RWMutex
	EmailsSent           *expvar.Int
	EmailsFailed         *expvar.Int
	EmailsRetried        *expvar.Int
	SMTPConnections      *expvar.Int
	ActiveWorkers        *expvar.Int
	JobsScheduled        *expvar.Int
	JobsCompleted        *expvar.Int
	JobsFailed           *expvar.Int
	ResponseTimes        *expvar.Map
	ErrorCounts          *expvar.Map
	startTime            time.Time
	log                  *logrus.Logger
}

var (
	instance *Metrics
	once     sync.Once
)

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			EmailsSent:      expvar.NewInt("emails_sent_total"),
			EmailsFailed:    expvar.NewInt("emails_failed_total"),
			EmailsRetried:   expvar.NewInt("emails_retried_total"),
			SMTPConnections: expvar.NewInt("smtp_connections_active"),
			ActiveWorkers:   expvar.NewInt("workers_active"),
			JobsScheduled:   expvar.NewInt("jobs_scheduled_total"),
			JobsCompleted:   expvar.NewInt("jobs_completed_total"),
			JobsFailed:      expvar.NewInt("jobs_failed_total"),
			ResponseTimes:   expvar.NewMap("response_times_ms"),
			ErrorCounts:     expvar.NewMap("error_counts"),
			startTime:       time.Now(),
			log:            logrus.New(),
		}
		
		// Register uptime metric
		expvar.Publish("uptime_seconds", expvar.Func(func() any {
			return int64(time.Since(instance.startTime).Seconds())
		}))
	})
	return instance
}

// RecordEmailSent increments the emails sent counter
func (m *Metrics) RecordEmailSent() {
	m.EmailsSent.Add(1)
}

// RecordEmailFailed increments the emails failed counter
func (m *Metrics) RecordEmailFailed() {
	m.EmailsFailed.Add(1)
}

// RecordEmailRetried increments the emails retried counter
func (m *Metrics) RecordEmailRetried() {
	m.EmailsRetried.Add(1)
}

// RecordSMTPConnection increments active SMTP connections
func (m *Metrics) RecordSMTPConnection() {
	m.SMTPConnections.Add(1)
}

// RecordSMTPDisconnection decrements active SMTP connections
func (m *Metrics) RecordSMTPDisconnection() {
	m.SMTPConnections.Add(-1)
}

// RecordWorkerStart increments active workers
func (m *Metrics) RecordWorkerStart() {
	m.ActiveWorkers.Add(1)
}

// RecordWorkerStop decrements active workers
func (m *Metrics) RecordWorkerStop() {
	m.ActiveWorkers.Add(-1)
}

// RecordJobScheduled increments scheduled jobs counter
func (m *Metrics) RecordJobScheduled() {
	m.JobsScheduled.Add(1)
}

// RecordJobCompleted increments completed jobs counter
func (m *Metrics) RecordJobCompleted() {
	m.JobsCompleted.Add(1)
}

// RecordJobFailed increments failed jobs counter
func (m *Metrics) RecordJobFailed() {
	m.JobsFailed.Add(1)
}

// RecordResponseTime records operation response time
func (m *Metrics) RecordResponseTime(operation string, duration time.Duration) {
	m.ResponseTimes.Add(operation, int64(duration.Milliseconds()))
}

// RecordError records error by type
func (m *Metrics) RecordError(errorType string) {
	m.ErrorCounts.Add(errorType, 1)
}

// StartMetricsServer starts the metrics HTTP server
func (m *Metrics) StartMetricsServer(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", expvar.Handler())
	mux.HandleFunc("/health", m.healthHandler)
	mux.HandleFunc("/ready", m.readinessHandler)
	
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			m.log.Errorf("Metrics server shutdown error: %v", err)
		}
	}()
	
	m.log.Infof("Metrics server starting on port %d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	
	return nil
}

func (m *Metrics) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

func (m *Metrics) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check if workers are available
	activeWorkers := m.ActiveWorkers.Value()
	if activeWorkers > 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready","active_workers":` + strconv.FormatInt(activeWorkers, 10) + `}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not_ready","active_workers":0}`))
	}
}
