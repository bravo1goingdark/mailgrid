package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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

// Server provides real-time monitoring dashboard
type Server struct {
	mu        sync.RWMutex
	stats     *CampaignStats
	server    *http.Server
	dashboard *DashboardServer
	clients   map[chan CampaignStats]bool
	stopping  bool
}

// NewServer creates a new monitoring server
func NewServer(port int) *Server {
	stats := &CampaignStats{
		Recipients:        make(map[string]*RecipientStatus),
		DomainBreakdown:   make(map[string]int),
		SMTPResponseCodes: make(map[string]int),
		LogEntries:        make([]LogEntry, 0, 1000),
	}

	server := &Server{
		stats:   stats,
		clients: make(map[chan CampaignStats]bool),
	}

	// Create dashboard
	server.dashboard = NewDashboardServer(server)

	// Set up HTTP server
	server.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.dashboard,
	}

	return server
}

// Start starts the monitoring server
func (s *Server) Start() error {
	log.Printf("🖥️  Starting monitoring dashboard on http://localhost%s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the monitoring server
func (s *Server) Stop() error {
	s.mu.Lock()
	s.stopping = true

	// Close all client channels
	for clientChan := range s.clients {
		close(clientChan)
	}
	s.clients = make(map[chan CampaignStats]bool)
	s.mu.Unlock()

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

// UpdateRecipientStatus updates the status of a specific recipient
func (s *Server) UpdateRecipientStatus(email string, status EmailStatus, duration time.Duration, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	recipient, exists := s.stats.Recipients[email]
	if !exists {
		recipient = &RecipientStatus{
			Email: email,
		}
		s.stats.Recipients[email] = recipient
	}

	// Update counts based on status change
	oldStatus := recipient.Status
	if oldStatus != "" {
		s.decrementStatusCount(oldStatus)
	}
	s.incrementStatusCount(status)

	recipient.Status = status
	recipient.LastAttempt = time.Now()
	recipient.Duration = duration.Nanoseconds() / 1e6 // Convert to milliseconds

	if status == StatusRetry {
		recipient.Attempts++
	}

	if errorMsg != "" {
		recipient.Error = errorMsg
	}

	// Update domain breakdown only when first tracking this recipient or when leaving pending
	shouldIncrementDomain := !exists || (oldStatus == StatusPending && status != StatusPending)
	if shouldIncrementDomain && !recipient.domainCounted {
		domain := extractDomain(email)
		if domain != "" {
			s.stats.DomainBreakdown[domain]++
		}
		recipient.domainCounted = true
	}

	// Calculate real-time metrics
	s.calculateMetrics()

	// Add log entry
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

func (s *Server) addLogEntry(level, message, email string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Email:     email,
	}

	s.stats.LogEntries = append(s.stats.LogEntries, entry)

	// Keep only the latest 1000 entries
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

	// Calculate average duration
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

	// Estimate time left
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

func (s *Server) broadcastUpdate() {
	if s.stopping {
		return
	}

	// Create a copy of stats for broadcasting
	statsCopy := *s.stats

	// Broadcast to all connected clients
	for clientChan := range s.clients {
		select {
		case clientChan <- statsCopy:
		default:
			// Client channel is full, remove it
			delete(s.clients, clientChan)
			close(clientChan)
		}
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
            <h1>📧 Mailgrid Monitor</h1>
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
                // Connection lost, retrying in 5 seconds
                setTimeout(connectEventSource, 5000);
            };
        }

        function updateDashboard(stats) {
            // Update basic stats
            document.getElementById('job-id').textContent = stats.job_id || '-';
            document.getElementById('start-time').textContent = stats.start_time ? new Date(stats.start_time).toLocaleString() : '-';
            document.getElementById('total-recipients').textContent = stats.total_recipients || 0;
            document.getElementById('sent-count').textContent = stats.sent_count || 0;
            document.getElementById('failed-count').textContent = stats.failed_count || 0;
            document.getElementById('emails-per-second').textContent = (stats.emails_per_second || 0).toFixed(1);
            document.getElementById('estimated-time').textContent = stats.estimated_time_left || '-';
            document.getElementById('avg-duration').textContent = Math.round(stats.avg_duration_ms || 0) + 'ms';

            // Update progress bar
            const progress = stats.total_recipients > 0 ? (stats.sent_count / stats.total_recipients) * 100 : 0;
            document.getElementById('sent-progress').style.width = progress + '%';

            // Update recipients list
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

            // Update logs
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

        // Start the connection
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
	json.NewEncoder(w).Encode(s.stats)
}

// handleStatusStream provides real-time updates via Server-Sent Events
func (s *Server) handleStatusStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan CampaignStats, 10)

	s.mu.Lock()
	s.clients[clientChan] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientChan)
		s.mu.Unlock()
		close(clientChan)
	}()

	// Send initial state
	s.mu.RLock()
	initial := *s.stats
	s.mu.RUnlock()

	data, _ := json.Marshal(initial)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()

	// Stream updates
	for {
		select {
		case stats, ok := <-clientChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(stats)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// Helper function to extract domain from email address
func extractDomain(email string) string {
	at := len(email)
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			at = i
			break
		}
	}
	if at < len(email) {
		return email[at+1:]
	}
	return ""
}

// GetStats returns the current campaign statistics for the dashboard
func (s *Server) GetStats() *CampaignStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	statsCopy := *s.stats
	return &statsCopy
}

// GetRecipients returns a slice of recipients for the dashboard table
func (s *Server) GetRecipients() []RecipientStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recipients := make([]RecipientStatus, 0, len(s.stats.Recipients))
	for _, recipient := range s.stats.Recipients {
		recipients = append(recipients, *recipient)
	}

	return recipients
}
