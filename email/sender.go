package email

import (
	"bytes"
	"fmt"
	"log"
	"mailgrid/config"
	"net/smtp"
	"strings"
)

// SendWithClient formats and delivers the email using an active SMTP client.
// It uses cfg.SMTP.From as both the envelope sender and header From address.
func SendWithClient(client *smtp.Client, cfg config.SMTPConfig, task Task) error {
	from := strings.TrimSpace(cfg.From)
	if from == "" {
		return fmt.Errorf("SMTP sender 'from' field in config is empty")
	}

	// SMTP envelope sender (must match SMTP auth user)
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}

	// Recipient
	if err := client.Rcpt(task.Recipient.Email); err != nil {
		return fmt.Errorf("RCPT TO error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command error: %w", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("Error closing SMTP writer: %v", err)
		}
	}()

	var msg bytes.Buffer
	headers := map[string]string{
		"From":         fmt.Sprintf("Mailgrid <%s>", from),
		"To":           task.Recipient.Email,
		"Subject":      task.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}
	for k, v := range headers {
		_, err := fmt.Fprintf(&msg, "%s: %s\r\n", k, strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("header write error: %w", err)
		}
	}
	msg.WriteString("\r\n" + strings.TrimSpace(task.Body))

	if _, err = w.Write(msg.Bytes()); err != nil {
		return fmt.Errorf("SMTP body write error: %w", err)
	}

	return nil
}
