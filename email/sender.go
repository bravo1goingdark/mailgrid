package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"

	"mailgrid/config"
)

func SendEmail(cfg config.SMTPConfig, to string, subject string, body string) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	var msg bytes.Buffer

	// RFC 5322-compliant message
	headers := map[string]string{
		"From":         fmt.Sprintf("<%s>", cfg.From),
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.TrimSpace(v)))
	}
	msg.WriteString("\r\n") // end of headers
	msg.WriteString(strings.TrimSpace(body))
	msg.WriteString("\r\n") // end of body

	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg.Bytes())
}
