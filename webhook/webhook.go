package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// CampaignResult represents job completion data sent to webhook
type CampaignResult struct {
	JobID                string    `json:"job_id"`
	Status               string    `json:"status"`
	TotalRecipients      int       `json:"total_recipients"`
	SuccessfulDeliveries int       `json:"successful_deliveries"`
	FailedDeliveries     int       `json:"failed_deliveries"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	DurationSeconds      int       `json:"duration_seconds"`
	CSVFile              string    `json:"csv_file,omitempty"`
	SheetURL             string    `json:"sheet_url,omitempty"`
	TemplateFile         string    `json:"template_file,omitempty"`
	ConcurrentWorkers    int       `json:"concurrent_workers"`
	ErrorMessage         string    `json:"error_message,omitempty"`
}

// Client handles webhook HTTP requests with goroutine tracking
type Client struct {
	httpClient *http.Client
	wg          sync.WaitGroup
	mu          sync.RWMutex
	closed      bool
}

// NewClient creates a new webhook client with timeout configuration
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendNotification sends a POST request to webhook URL with campaign results.
// This is non-blocking and spawns a goroutine for the HTTP request.
func (c *Client) SendNotification(webhookURL string, result CampaignResult) error {
	if webhookURL == "" {
		return nil // No webhook configured
	}

	// Check if client is closed
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("webhook client is closed")
	}
	c.mu.RUnlock()

	// Marshal result to JSON
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mailgrid-Webhook/1.0")

	// Track goroutine with WaitGroup
	c.wg.Add(1)

	// Send request in goroutine (non-blocking)
	go func() {
		defer c.wg.Done()

		// Create a context with timeout for this specific request
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req = req.WithContext(ctx)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf(" Webhook delivery failed: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf(" Webhook delivered successfully to %s (status: %d)", webhookURL, resp.StatusCode)
		} else {
			log.Printf(" Webhook delivery failed: %s returned status %d", webhookURL, resp.StatusCode)
		}
	}()

	return nil
}

// SendNotificationSync sends a synchronous POST request to webhook URL
func (c *Client) SendNotificationSync(webhookURL string, result CampaignResult) error {
	if webhookURL == "" {
		return nil // No webhook configured
	}

	// Marshal result to JSON
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mailgrid-Webhook/1.0")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf(" Webhook delivery failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf(" Webhook delivered successfully to %s (status: %d)", webhookURL, resp.StatusCode)
	} else {
		log.Printf(" Webhook delivery failed: %s returned status %d", webhookURL, resp.StatusCode)
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// ValidateURL performs basic validation on webhook URL
func ValidateURL(url string) error {
	if url == "" {
		return nil // Empty URL is valid (no webhook)
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use http or https scheme")
	}

	return nil
}

// Close waits for all pending webhook requests to complete
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	c.wg.Wait()
}
