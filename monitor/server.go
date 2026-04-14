package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// EmailStatus represents the status of an individual email
type EmailStatus string

const (
	StatusPending EmailStatus = "pending"
	StatusSending EmailStatus = "sending"
	StatusSent    EmailStatus = "sent"
	StatusFailed  EmailStatus = "failed"
	StatusRetry   EmailStatus = "retry"

	// maxSSEClients caps open SSE connections to prevent resource exhaustion.
	maxSSEClients = 50
)

// RecipientStatus tracks the status of a single recipient
type RecipientStatus struct {
	Email         string      `json:"email"`
	Status        EmailStatus `json:"status"`
	Attempts      int         `json:"attempts"`
	LastAttempt   time.Time   `json:"last_attempt"`
	Error         string      `json:"error,omitempty"`
	Duration      int64       `json:"duration_ms"` // Duration in milliseconds
	domainCounted bool        `json:"-"`
}

// CampaignStats represents real-time campaign statistics
type CampaignStats struct {
	JobID             string                      `json:"job_id"`
	StartTime         time.Time                   `json:"start_time"`
	TotalRecipients   int                         `json:"total_recipients"`
	PendingCount      int                         `json:"pending_count"`
	SendingCount      int                         `json:"sending_count"`
	SentCount         int                         `json:"sent_count"`
	FailedCount       int                         `json:"failed_count"`
	RetryCount        int                         `json:"retry_count"`
	EmailsPerSecond   float64                     `json:"emails_per_second"`
	EstimatedTimeLeft string                      `json:"estimated_time_left"`
	AvgDurationMs     float64                     `json:"avg_duration_ms"`
	Recipients        map[string]*RecipientStatus `json:"recipients"`
	DomainBreakdown   map[string]int              `json:"domain_breakdown"`
	SMTPResponseCodes map[string]int              `json:"smtp_response_codes"`
	ConfigSummary     ConfigSummary               `json:"config_summary"`
	LogEntries        []LogEntry                  `json:"log_entries"`
}

// ConfigSummary holds campaign configuration details
type ConfigSummary struct {
	CSVFile           string `json:"csv_file"`
	SheetURL          string `json:"sheet_url"`
	TemplateFile      string `json:"template_file"`
	ConcurrentWorkers int    `json:"concurrent_workers"`
	BatchSize         int    `json:"batch_size"`
	RetryLimit        int    `json:"retry_limit"`
	FilterExpression  string `json:"filter_expression"`
}

// LogEntry represents a monitoring log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Email     string    `json:"email,omitempty"`
}

// SSEClient represents a connected SSE client with atomic last-active tracking.
type SSEClient struct {
	Chan       chan CampaignStats
	lastActive atomic.Int64 // Unix nanos — updated without holding s.mu
	RemoteAddr string
}

// Server provides real-time monitoring dashboard
type Server struct {
	mu            sync.RWMutex
	stats         *CampaignStats
	server        *http.Server
	clients       map[string]*SSEClient // Map client ID → client
	clientID      uint64                // Counter for unique client IDs
	stopping      bool
	clientTimeout time.Duration
	cleanupTicker *time.Ticker
	startTime     time.Time

	// Broadcast debounce: callers set dirty; the broadcaster goroutine flushes at 100ms cadence.
	dirty atomic.Bool
	quit  chan struct{} // closed by Stop() to terminate background goroutines
}

// NewServer creates a new monitoring server. clientTimeout controls how long
// an idle SSE connection is kept alive; pass 0 to use the 5-minute default.
func NewServer(port int, clientTimeout time.Duration) *Server {
	if clientTimeout <= 0 {
		clientTimeout = 5 * time.Minute
	}

	stats := &CampaignStats{
		Recipients:        make(map[string]*RecipientStatus),
		DomainBreakdown:   make(map[string]int),
		SMTPResponseCodes: make(map[string]int),
		LogEntries:        make([]LogEntry, 0, 1000),
	}

	s := &Server{
		stats:         stats,
		clients:       make(map[string]*SSEClient),
		clientTimeout: clientTimeout,
		cleanupTicker: time.NewTicker(1 * time.Minute),
		startTime:     time.Now(),
		quit:          make(chan struct{}),
	}

	s.server = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           http.HandlerFunc(s.serveHTTP),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start background goroutines.
	go s.cleanupInactiveClients()
	go s.runBroadcaster()

	return s
}

// serveHTTP routes requests to appropriate handlers
func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.handleDashboard(w, r)
	case "/api/status":
		s.handleStatusAPI(w, r)
	case "/api/stream":
		s.handleStatusStream(w, r)
	case "/metrics":
		s.handleMetrics(w, r)
	case "/health":
		s.handleHealth(w, r)
	case "/ready":
		s.handleReady(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleMetrics writes a minimal Prometheus-compatible text exposition of
// current campaign counters. No external dependencies — plain fmt.Fprintf.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	stats := s.stats
	var sent, failed, pending, total int
	var durationSecs float64
	if stats != nil {
		sent = stats.SentCount
		failed = stats.FailedCount
		pending = stats.PendingCount
		total = stats.TotalRecipients
		durationSecs = time.Since(stats.StartTime).Seconds()
		if durationSecs < 0 {
			durationSecs = 0
		}
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintf(w, "# HELP mailgrid_emails_sent_total Total emails successfully sent\n")
	fmt.Fprintf(w, "# TYPE mailgrid_emails_sent_total counter\n")
	fmt.Fprintf(w, "mailgrid_emails_sent_total %d\n", sent)
	fmt.Fprintf(w, "# HELP mailgrid_emails_failed_total Total emails that permanently failed\n")
	fmt.Fprintf(w, "# TYPE mailgrid_emails_failed_total counter\n")
	fmt.Fprintf(w, "mailgrid_emails_failed_total %d\n", failed)
	fmt.Fprintf(w, "# HELP mailgrid_emails_pending Total emails pending\n")
	fmt.Fprintf(w, "# TYPE mailgrid_emails_pending gauge\n")
	fmt.Fprintf(w, "mailgrid_emails_pending %d\n", pending)
	fmt.Fprintf(w, "# HELP mailgrid_emails_total Total recipients in campaign\n")
	fmt.Fprintf(w, "# TYPE mailgrid_emails_total gauge\n")
	fmt.Fprintf(w, "mailgrid_emails_total %d\n", total)
	fmt.Fprintf(w, "# HELP mailgrid_campaign_duration_seconds Elapsed seconds since campaign start\n")
	fmt.Fprintf(w, "# TYPE mailgrid_campaign_duration_seconds gauge\n")
	fmt.Fprintf(w, "mailgrid_campaign_duration_seconds %.3f\n", durationSecs)
}

// handleHealth returns a basic health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// handleReady checks if the system is ready to handle traffic
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uptime := time.Since(s.startTime)
	if uptime < 5*time.Second {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "System still starting up (uptime: %v)", uptime)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Ready")
}

// Start starts the monitoring server
func (s *Server) Start() error {
	log.Printf("  Starting monitoring dashboard on http://localhost%s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the monitoring server
func (s *Server) Stop() error {
	s.mu.Lock()
	s.stopping = true

	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}

	// Close all client channels before releasing the lock.
	for _, client := range s.clients {
		close(client.Chan)
	}
	s.clients = make(map[string]*SSEClient)
	s.mu.Unlock()

	// Signal background goroutines to exit.
	close(s.quit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// InitializeCampaign initializes campaign tracking
func (s *Server) InitializeCampaign(jobID string, config ConfigSummary, totalRecipients int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.JobID = jobID
	s.stats.StartTime = time.Now()
	s.stats.TotalRecipients = totalRecipients
	s.stats.PendingCount = totalRecipients
	s.stats.ConfigSummary = config

	s.broadcastUpdate()
}

// UpdateRecipientStatus updates the status of a specific recipient.
// The actual SSE broadcast is debounced via the broadcaster goroutine.
func (s *Server) UpdateRecipientStatus(email string, status EmailStatus, duration time.Duration, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	recipient, exists := s.stats.Recipients[email]
	if !exists {
		recipient = &RecipientStatus{Email: email}
		s.stats.Recipients[email] = recipient
	}

	oldStatus := recipient.Status
	if oldStatus != "" {
		s.decrementStatusCount(oldStatus)
	}
	s.incrementStatusCount(status)

	recipient.Status = status
	recipient.LastAttempt = time.Now()
	recipient.Duration = duration.Nanoseconds() / 1e6

	if status == StatusRetry {
		recipient.Attempts++
	}
	if errorMsg != "" {
		recipient.Error = errorMsg
	}

	shouldIncrementDomain := !exists || (oldStatus == StatusPending && status != StatusPending)
	if shouldIncrementDomain && !recipient.domainCounted {
		if domain := extractDomain(email); domain != "" {
			s.stats.DomainBreakdown[domain]++
		}
		recipient.domainCounted = true
	}

	s.calculateMetrics()
	s.addLogEntry("INFO", fmt.Sprintf("Email %s: %s", email, status), email)
	s.broadcastUpdate()
}

// AddSMTPResponse records an SMTP response code
func (s *Server) AddSMTPResponse(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.SMTPResponseCodes[code]++
	s.broadcastUpdate()
}

// AddLogEntry adds a log entry to the monitoring dashboard
func (s *Server) AddLogEntry(level, message, email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addLogEntry(level, message, email)
	s.broadcastUpdate()
}

// generateClientID generates a unique client ID. Must be called with s.mu held.
func (s *Server) generateClientID() string {
	s.clientID++
	return fmt.Sprintf("client-%d", s.clientID)
}

func (s *Server) addLogEntry(level, message, email string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Email:     email,
	}
	s.stats.LogEntries = append(s.stats.LogEntries, entry)
	// Keep only the latest 1000 entries (trim from front).
	if len(s.stats.LogEntries) > 1000 {
		s.stats.LogEntries = s.stats.LogEntries[len(s.stats.LogEntries)-1000:]
	}
}

func (s *Server) incrementStatusCount(status EmailStatus) {
	switch status {
	case StatusPending:
		s.stats.PendingCount++
	case StatusSending:
		s.stats.SendingCount++
	case StatusSent:
		s.stats.SentCount++
	case StatusFailed:
		s.stats.FailedCount++
	case StatusRetry:
		s.stats.RetryCount++
	}
}

func (s *Server) decrementStatusCount(status EmailStatus) {
	switch status {
	case StatusPending:
		s.stats.PendingCount--
	case StatusSending:
		s.stats.SendingCount--
	case StatusSent:
		s.stats.SentCount--
	case StatusFailed:
		s.stats.FailedCount--
	case StatusRetry:
		s.stats.RetryCount--
	}
}

func (s *Server) calculateMetrics() {
	if s.stats.StartTime.IsZero() {
		return
	}
	elapsed := time.Since(s.stats.StartTime)
	if elapsed.Seconds() > 0 {
		s.stats.EmailsPerSecond = float64(s.stats.SentCount) / elapsed.Seconds()
	}

	var totalDuration int64
	var count int
	for _, recipient := range s.stats.Recipients {
		if recipient.Status == StatusSent && recipient.Duration > 0 {
			totalDuration += recipient.Duration
			count++
		}
	}
	if count > 0 {
		s.stats.AvgDurationMs = float64(totalDuration) / float64(count)
	}

	if s.stats.EmailsPerSecond > 0 {
		remaining := s.stats.TotalRecipients - s.stats.SentCount - s.stats.FailedCount
		if remaining > 0 {
			estimatedSeconds := float64(remaining) / s.stats.EmailsPerSecond
			s.stats.EstimatedTimeLeft = time.Duration(estimatedSeconds * float64(time.Second)).Round(time.Second).String()
		} else {
			s.stats.EstimatedTimeLeft = "0s"
		}
	}
}

// broadcastUpdate marks stats as dirty. The actual SSE send happens in the
// broadcaster goroutine at most once per 100ms, eliminating the previous
// deadlock (callers hold s.mu; old code tried to re-acquire s.mu inside here).
// Must be called with s.mu held (read or write).
func (s *Server) broadcastUpdate() {
	if !s.stopping {
		s.dirty.Store(true)
	}
}

// runBroadcaster is a background goroutine that flushes pending broadcasts
// at a 100ms cadence, debouncing high-frequency stat updates.
func (s *Server) runBroadcaster() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if s.dirty.Swap(false) {
				s.flushBroadcast()
			}
		case <-s.quit:
			return
		}
	}
}

// flushBroadcast copies current stats under a read lock and sends to all
// connected SSE clients. Stale clients (channel full) are removed.
func (s *Server) flushBroadcast() {
	// Deep-copy stats under read lock.
	s.mu.RLock()
	statsCopy := *s.stats
	statsCopy.Recipients = make(map[string]*RecipientStatus, len(s.stats.Recipients))
	for k, v := range s.stats.Recipients {
		cp := *v
		statsCopy.Recipients[k] = &cp
	}
	statsCopy.DomainBreakdown = make(map[string]int, len(s.stats.DomainBreakdown))
	for k, v := range s.stats.DomainBreakdown {
		statsCopy.DomainBreakdown[k] = v
	}
	statsCopy.SMTPResponseCodes = make(map[string]int, len(s.stats.SMTPResponseCodes))
	for k, v := range s.stats.SMTPResponseCodes {
		statsCopy.SMTPResponseCodes[k] = v
	}
	if len(s.stats.LogEntries) > 0 {
		statsCopy.LogEntries = make([]LogEntry, len(s.stats.LogEntries))
		copy(statsCopy.LogEntries, s.stats.LogEntries)
	}
	// Snapshot client references (not their channels content).
	type clientRef struct {
		id     string
		client *SSEClient
	}
	s.mu.RUnlock()

	s.mu.RLock()
	clientRefs := make([]clientRef, 0, len(s.clients))
	for id, c := range s.clients {
		clientRefs = append(clientRefs, clientRef{id, c})
	}
	s.mu.RUnlock()

	// Send to clients without holding any lock.
	var stale []string
	for _, ref := range clientRefs {
		select {
		case ref.client.Chan <- statsCopy:
			ref.client.lastActive.Store(time.Now().UnixNano())
		default:
			// Channel full — mark client as stale.
			stale = append(stale, ref.id)
		}
	}

	// Remove stale clients under write lock.
	if len(stale) > 0 {
		s.mu.Lock()
		for _, id := range stale {
			if client, ok := s.clients[id]; ok {
				close(client.Chan)
				delete(s.clients, id)
				log.Printf("[DISCONNECT] Removed stale SSE client: %s", id)
			}
		}
		s.mu.Unlock()
	}
}

// cleanupInactiveClients removes SSE clients that have been idle beyond clientTimeout.
func (s *Server) cleanupInactiveClients() {
	for {
		select {
		case <-s.cleanupTicker.C:
		case <-s.quit:
			return
		}

		s.mu.Lock()
		if s.stopping {
			s.mu.Unlock()
			return
		}

		now := time.Now()
		var toRemove []string
		for id, client := range s.clients {
			lastActive := time.Unix(0, client.lastActive.Load())
			if now.Sub(lastActive) > s.clientTimeout {
				toRemove = append(toRemove, id)
			}
		}
		for _, id := range toRemove {
			if client, ok := s.clients[id]; ok {
				close(client.Chan)
				delete(s.clients, id)
				log.Printf("[DISCONNECT] Removed inactive SSE client: %s", id)
			}
		}
		s.mu.Unlock()
	}
}

// handleDashboard serves the monitoring dashboard HTML
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Mailgrid Monitor</title>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .stat-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stat-value { font-size: 2em; font-weight: bold; color: #2563eb; }
        .stat-label { color: #666; margin-top: 5px; }
        .recipients-table { background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .table-header { background: #f8fafc; padding: 15px; border-bottom: 1px solid #e2e8f0; font-weight: bold; }
        .table-row { padding: 10px 15px; border-bottom: 1px solid #e2e8f0; display: grid; grid-template-columns: 2fr 1fr 1fr 1fr 3fr; gap: 10px; align-items: center; }
        .status { padding: 4px 8px; border-radius: 4px; font-size: 0.8em; font-weight: bold; text-transform: uppercase; }
        .status-pending { background: #fef3c7; color: #92400e; }
        .status-sending { background: #dbeafe; color: #1e40af; }
        .status-sent { background: #d1fae5; color: #065f46; }
        .status-failed { background: #fee2e2; color: #991b1b; }
        .status-retry { background: #fde68a; color: #92400e; }
        .logs { background: white; border-radius: 8px; padding: 20px; margin-top: 20px; max-height: 300px; overflow-y: auto; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .log-entry { padding: 5px 0; border-bottom: 1px solid #f0f0f0; font-family: monospace; font-size: 0.9em; }
        .progress-bar { background: #e5e7eb; border-radius: 4px; overflow: hidden; margin: 10px 0; }
        .progress-fill { background: #10b981; height: 8px; transition: width 0.3s ease; }
        .two-column { display: grid; grid-template-columns: 2fr 1fr; gap: 20px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1> Mailgrid Monitor</h1>
            <div id="campaign-info">
                <strong>Campaign:</strong> <span id="job-id">-</span> |
                <strong>Started:</strong> <span id="start-time">-</span>
            </div>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value" id="total-recipients">0</div>
                <div class="stat-label">Total Recipients</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="sent-count">0</div>
                <div class="stat-label">Sent</div>
                <div class="progress-bar">
                    <div class="progress-fill" id="sent-progress" style="width: 0%"></div>
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="failed-count">0</div>
                <div class="stat-label">Failed</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="emails-per-second">0.0</div>
                <div class="stat-label">Emails/sec</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="estimated-time">-</div>
                <div class="stat-label">Est. Time Left</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="avg-duration">0ms</div>
                <div class="stat-label">Avg Duration</div>
            </div>
        </div>

        <div class="two-column">
            <div>
                <div class="recipients-table">
                    <div class="table-header">Recipients Status</div>
                    <div class="table-row" style="font-weight: bold; background: #f8fafc;">
                        <div>Email</div>
                        <div>Status</div>
                        <div>Attempts</div>
                        <div>Duration</div>
                        <div>Error</div>
                    </div>
                    <div id="recipients-list"></div>
                </div>
            </div>

            <div>
                <div class="logs">
                    <h3>Live Logs</h3>
                    <div id="log-entries"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        let eventSource;

        function connectEventSource() {
            eventSource = new EventSource('/api/stream');

            eventSource.onmessage = function(event) {
                const stats = JSON.parse(event.data);
                updateDashboard(stats);
            };

            eventSource.onerror = function() {
                setTimeout(connectEventSource, 5000);
            };
        }

        function updateDashboard(stats) {
            document.getElementById('job-id').textContent = stats.job_id || '-';
            document.getElementById('start-time').textContent = stats.start_time ? new Date(stats.start_time).toLocaleString() : '-';
            document.getElementById('total-recipients').textContent = stats.total_recipients || 0;
            document.getElementById('sent-count').textContent = stats.sent_count || 0;
            document.getElementById('failed-count').textContent = stats.failed_count || 0;
            document.getElementById('emails-per-second').textContent = (stats.emails_per_second || 0).toFixed(1);
            document.getElementById('estimated-time').textContent = stats.estimated_time_left || '-';
            document.getElementById('avg-duration').textContent = Math.round(stats.avg_duration_ms || 0) + 'ms';

            const progress = stats.total_recipients > 0 ? (stats.sent_count / stats.total_recipients) * 100 : 0;
            document.getElementById('sent-progress').style.width = progress + '%';

            const recipientsList = document.getElementById('recipients-list');
            recipientsList.innerHTML = '';

            Object.values(stats.recipients || {}).slice(0, 20).forEach(recipient => {
                const row = document.createElement('div');
                row.className = 'table-row';
                row.innerHTML = ` + "`" + `
                    <div>${recipient.email}</div>
                    <div><span class="status status-${recipient.status}">${recipient.status}</span></div>
                    <div>${recipient.attempts || 0}</div>
                    <div>${recipient.duration_ms || 0}ms</div>
                    <div style="font-size: 0.8em; color: #666;">${recipient.error || ''}</div>
                ` + "`" + `;
                recipientsList.appendChild(row);
            });

            const logEntries = document.getElementById('log-entries');
            logEntries.innerHTML = '';

            (stats.log_entries || []).slice(-10).reverse().forEach(entry => {
                const logDiv = document.createElement('div');
                logDiv.className = 'log-entry';
                const timestamp = new Date(entry.timestamp).toLocaleTimeString();
                logDiv.textContent = ` + "`${timestamp} [${entry.level}] ${entry.message}`" + `;
                logEntries.appendChild(logDiv);
            });
        }

        connectEventSource();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, dashboardHTML)
}

// handleStatusAPI returns current status as JSON
func (s *Server) handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.stats); err != nil {
		http.Error(w, "failed to encode status", http.StatusInternalServerError)
	}
}

// handleStatusStream provides real-time updates via Server-Sent Events.
// No CORS wildcard is set — the dashboard is localhost-only.
// Connections beyond maxSSEClients are rejected with 429.
func (s *Server) handleStatusStream(w http.ResponseWriter, r *http.Request) {
	// Check connection cap before registering client.
	s.mu.Lock()
	if s.stopping {
		s.mu.Unlock()
		return
	}
	if len(s.clients) >= maxSSEClients {
		s.mu.Unlock()
		http.Error(w, "Too many SSE connections", http.StatusTooManyRequests)
		return
	}

	clientID := s.generateClientID()
	clientChan := make(chan CampaignStats, 10)
	c := &SSEClient{
		Chan:       clientChan,
		RemoteAddr: r.RemoteAddr,
	}
	c.lastActive.Store(time.Now().UnixNano())
	s.clients[clientID] = c
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
		close(clientChan)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// No Access-Control-Allow-Origin header — dashboard is localhost-only.

	// Send initial state immediately.
	s.mu.RLock()
	initial := *s.stats
	s.mu.RUnlock()

	data, _ := json.Marshal(initial)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	for {
		select {
		case stats, ok := <-clientChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(stats)
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			c.lastActive.Store(time.Now().UnixNano())
		case <-r.Context().Done():
			return
		}
	}
}

// extractDomain extracts the domain part from an email address.
func extractDomain(email string) string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return email[i+1:]
		}
	}
	return ""
}

// GetStats returns a shallow copy of the current campaign stats.
func (s *Server) GetStats() CampaignStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.stats
}

// GetRecipients returns a slice of recipient statuses for the dashboard table.
func (s *Server) GetRecipients() []RecipientStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recipients := make([]RecipientStatus, 0, len(s.stats.Recipients))
	for _, recipient := range s.stats.Recipients {
		recipients = append(recipients, *recipient)
	}
	return recipients
}
