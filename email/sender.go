// Package email provides helpers for formatting and delivering mail via SMTP.
//
// SendWithClient sanitizing and deduplicates recipients case-insensitively across
// the To, CC, and BCC lists in that order. The primary recipient is always
// addressed first, and only one RCPT command is issued per unique address. CC
// headers include only deduplicated CC recipients while BCC addresses are added
// to the envelope but never leaked via headers.
package email

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/bravo1goingdark/mailgrid/config"
	"io"
	"log"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var bufWriterPool = sync.Pool{New: func() any { return bufio.NewWriterSize(io.Discard, 64*1024) }}
var copyBufPool = sync.Pool{New: func() any { b := make([]byte, 32*1024); return &b }}

// SendWithClient formats and delivers an email using an active SMTP client.
// Recipients from the To, CC, and BCC fields are trimmed, deduplicated in that
// order (case-insensitively), and issued RCPT commands exactly once with the
// primary recipient always first. The CC header is rendered from the unique CC
// list, while BCC addresses are kept solely on the SMTP envelope.
// It uses cfg.SMTP.From as both the envelope sender and header From address.
func SendWithClient(client *smtp.Client, cfg config.SMTPConfig, task Task) (err error) {
	from := strings.TrimSpace(cfg.From)
	if from == "" {
		return fmt.Errorf("SMTP sender 'from' field in config is empty")
	}

	// SMTP envelope sender (must match SMTP auth user)
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM error: %w", err)
	}

	to := strings.TrimSpace(task.Recipient.Email)
	if to == "" {
		return fmt.Errorf("recipient email is empty")
	}

	body := strings.TrimSpace(task.Body)

	seen := make(map[string]struct{}, 1+len(task.CC)+len(task.BCC))
	seen[strings.ToLower(to)] = struct{}{}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO error for %s: %w", to, err)
	}

	var ccFailures []string
	var bccFailures []string
	var rcptErr error

	uniqueCC := make([]string, 0, len(task.CC))
	for _, cc := range task.CC {
		cc = strings.TrimSpace(cc)
		if cc == "" {
			continue
		}
		key := strings.ToLower(cc)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		uniqueCC = append(uniqueCC, cc)
		if err := client.Rcpt(cc); err != nil {
			ccFailures = append(ccFailures, cc)
			log.Printf("️ Failed to add CC: %s (%v)", cc, err)
			// Track first RCPT error but continue processing others
			if rcptErr == nil {
				rcptErr = fmt.Errorf("failed to add CC recipient %s: %w", cc, err)
			}
		}
	}

	for _, bcc := range task.BCC {
		bcc = strings.TrimSpace(bcc)
		if bcc == "" {
			continue
		}
		key := strings.ToLower(bcc)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if err := client.Rcpt(bcc); err != nil {
			bccFailures = append(bccFailures, bcc)
			log.Printf("️ Failed to add BCC: %s (%v)", bcc, err)
			// Track first RCPT error but continue processing others
			if rcptErr == nil {
				rcptErr = fmt.Errorf("failed to add BCC recipient %s: %w", bcc, err)
			}
		}
	}

	// If there were RCPT failures, return the error
	if rcptErr != nil {
		return rcptErr
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command error: %w", err)
	}

	bw := bufWriterPool.Get().(*bufio.Writer)
	bw.Reset(w)
	defer func() {
		if cerr := bw.Flush(); cerr != nil && err == nil {
			err = fmt.Errorf("flush SMTP writer: %w", cerr)
		}
		bw.Reset(io.Discard)
		bufWriterPool.Put(bw)
		if cerr := w.Close(); cerr != nil {
			log.Printf("Error closing SMTP writer: %v", cerr)
		}
	}()

	boundary := "mixed_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	headers := map[string]string{
		"From":         fmt.Sprintf("Mailgrid <%s>", from),
		"To":           to,
		"Subject":      task.Subject,
		"MIME-Version": "1.0",
	}
	if len(uniqueCC) > 0 {
		headers["CC"] = strings.Join(uniqueCC, ", ")
	}
	if len(task.Attachments) > 0 {
		headers["Content-Type"] = "multipart/mixed; boundary=" + boundary
	} else {
		headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	}

	for k, v := range headers {
		if _, err = bw.WriteString(k + ": " + strings.TrimSpace(v) + "\r\n"); err != nil {
			return fmt.Errorf("write header: %w", err)
		}
	}
	if _, err = bw.WriteString("\r\n"); err != nil {
		return fmt.Errorf("write header/body separator: %w", err)
	}

	if len(task.Attachments) > 0 {
		if task.Body != "" {
			if _, err = bw.WriteString("--" + boundary + "\r\n"); err != nil {
				return fmt.Errorf("write body boundary: %w", err)
			}
			if _, err = bw.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"); err != nil {
				return fmt.Errorf("write body headers: %w", err)
			}
			if _, err = bw.WriteString(body + "\r\n"); err != nil {
				return fmt.Errorf("write body: %w", err)
			}
		}

		for _, path := range task.Attachments {
			if _, err = bw.WriteString("--" + boundary + "\r\n"); err != nil {
				return fmt.Errorf("write attachment boundary: %w", err)
			}
			mt := mime.TypeByExtension(filepath.Ext(path))
			if mt == "" {
				mt = "application/octet-stream"
			}
			if _, err = bw.WriteString("Content-Type: " + mt + "\r\n"); err != nil {
				return fmt.Errorf("write content type: %w", err)
			}
			if _, err = bw.WriteString("Content-Disposition: attachment; filename=\"" + filepath.Base(path) + "\"\r\n"); err != nil {
				return fmt.Errorf("write content disposition: %w", err)
			}
			if _, err = bw.WriteString("Content-Transfer-Encoding: base64\r\n\r\n"); err != nil {
				return fmt.Errorf("write transfer encoding: %w", err)
			}
			file, ferr := os.Open(path)
			if ferr != nil {
				return fmt.Errorf("open attachment: %w", ferr)
			}
			enc := base64.NewEncoder(base64.StdEncoding, bw)
			bufPtr := copyBufPool.Get().(*[]byte)
			_, err = io.CopyBuffer(enc, file, *bufPtr)
			copyBufPool.Put(bufPtr)
			if err != nil {
				if cerr := file.Close(); cerr != nil {
					log.Printf("Error closing attachment file %s: %v", path, cerr)
				}
				if cerr := enc.Close(); cerr != nil {
					log.Printf("Error closing base64 encoder: %v", cerr)
				}
				return fmt.Errorf("encode attachment: %w", err)
			}
			if err = enc.Close(); err != nil {
				if cerr := file.Close(); cerr != nil {
					log.Printf("Error closing attachment file %s: %v", path, cerr)
				}
				return fmt.Errorf("close encoder: %w", err)
			}
			if err = file.Close(); err != nil {
				return fmt.Errorf("close attachment: %w", err)
			}
			if _, err = bw.WriteString("\r\n"); err != nil {
				return fmt.Errorf("write attachment newline: %w", err)
			}
		}
		if _, err = bw.WriteString("--" + boundary + "--"); err != nil {
			return fmt.Errorf("write closing boundary: %w", err)
		}
	} else {
		if _, err = bw.WriteString(body); err != nil {
			return fmt.Errorf("write body: %w", err)
		}
	}

	return err
}
