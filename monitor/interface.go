package monitor

import "time"

// Monitor interface for reporting email sending progress
type Monitor interface {
	// InitializeCampaign sets up campaign tracking
	InitializeCampaign(jobID string, config ConfigSummary, totalRecipients int)

	// InitializePending registers a batch of recipients in the Pending state
	// under a single lock acquisition. Use this in place of N sequential
	// UpdateRecipientStatus calls when seeding the dashboard at dispatch start.
	InitializePending(emails []string)

	// UpdateRecipientStatus updates the status of a specific recipient
	UpdateRecipientStatus(email string, status EmailStatus, duration time.Duration, errorMsg string)

	// AddSMTPResponse records an SMTP response code
	AddSMTPResponse(code string)

	// AddLogEntry adds a log entry to the monitoring dashboard
	AddLogEntry(level, message, email string)
}

// NoOpMonitor is a monitor that does nothing (null object pattern)
type NoOpMonitor struct{}

func (n *NoOpMonitor) InitializeCampaign(jobID string, config ConfigSummary, totalRecipients int) {}
func (n *NoOpMonitor) InitializePending(emails []string)                                          {}
func (n *NoOpMonitor) UpdateRecipientStatus(email string, status EmailStatus, duration time.Duration, errorMsg string) {
}
func (n *NoOpMonitor) AddSMTPResponse(code string)              {}
func (n *NoOpMonitor) AddLogEntry(level, message, email string) {}

// NewNoOpMonitor creates a no-op monitor
func NewNoOpMonitor() Monitor {
	return &NoOpMonitor{}
}
