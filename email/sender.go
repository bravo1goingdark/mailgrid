package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"github.com/bravo1goingdark/mailgrid/config"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	boundary := "mixed_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	headers := map[string]string{
		"From":         fmt.Sprintf("Mailgrid <%s>", from),
		"To":           task.Recipient.Email,
		"Subject":      task.Subject,
		"MIME-Version": "1.0",
	}
	if len(task.Attachments) > 0 {
		headers["Content-Type"] = "multipart/mixed; boundary=" + boundary
	} else {
		headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	}

	for k, v := range headers {
		_, err := fmt.Fprintf(&msg, "%s: %s\r\n", k, strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("header write error: %w", err)
		}
	}
	msg.WriteString("\r\n")

	if len(task.Attachments) > 0 {
		if task.Body != "" {
			fmt.Fprintf(&msg, "--%s\r\n", boundary)
			msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			msg.WriteString(strings.TrimSpace(task.Body) + "\r\n")
		}

		for _, path := range task.Attachments {
			fmt.Fprintf(&msg, "--%s\r\n", boundary)
			mt := mime.TypeByExtension(filepath.Ext(path))
			if mt == "" {
				mt = "application/octet-stream"
			}
			msg.WriteString("Content-Type: " + mt + "\r\n")
			msg.WriteString("Content-Disposition: attachment; filename=\"" + filepath.Base(path) + "\"\r\n")
			msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open attachment: %w", err)
			}
			enc := base64.NewEncoder(base64.StdEncoding, &msg)
			if _, err := io.Copy(enc, file); err != nil {
				file.Close()
				enc.Close()
				return fmt.Errorf("encode attachment: %w", err)
			}
			err = enc.Close()
			if err != nil {
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}
			msg.WriteString("\r\n")
		}
		fmt.Fprintf(&msg, "--%s--", boundary)
	} else {
		msg.WriteString(strings.TrimSpace(task.Body))
	}

	if _, err = w.Write(msg.Bytes()); err != nil {
		return fmt.Errorf("SMTP body write error: %w", err)
	}

	return nil
}
