package email

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// SendWithClient formats and delivers the email using an active SMTP client.
func SendWithClient(client *smtp.Client, task Task) error {
	if err := client.Mail(task.Recipient.Email); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}
	if err := client.Rcpt(task.Recipient.Email); err != nil {
		return fmt.Errorf("RCPT TO error: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA error: %w", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("Failed to close writer: %v", err)
		}
	}()

	var msg bytes.Buffer
	headers := map[string]string{
		"From":         fmt.Sprintf("<%s>", task.Recipient.Email),
		"To":           task.Recipient.Email,
		"Subject":      task.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.TrimSpace(v)))
	}
	msg.WriteString("\r\n" + strings.TrimSpace(task.Body) + "\r\n")

	_, err = w.Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}
