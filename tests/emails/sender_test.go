package email

import (
	"bufio"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/bravo1goingdark/mailgrid/config"
	emailpkg "github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/parser"
)

func TestSendWithClient_RCPTOrderAndDedup(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	var (
		rcpts []string
		data  strings.Builder
		wg    sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			_ = serverConn.Close()
		}()
		reader := bufio.NewReader(serverConn)
		writer := bufio.NewWriter(serverConn)
		_, _ = fmt.Fprint(writer, "220 mock SMTP ready\r\n")
		_ = writer.Flush()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(cmd, "HELO") || strings.HasPrefix(cmd, "EHLO"):
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "MAIL FROM"):
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "RCPT TO"):
				addr := strings.TrimSpace(strings.Trim(strings.TrimPrefix(cmd, "RCPT TO:"), "<>"))
				rcpts = append(rcpts, addr)
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "DATA"):
				_, _ = fmt.Fprint(writer, "354 End data with <CRLF>.<CRLF>\r\n")
				_ = writer.Flush()
				for {
					l, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						_, _ = fmt.Fprint(writer, "250 OK\r\n")
						break
					}
					data.WriteString(l)
				}
			case strings.HasPrefix(cmd, "QUIT"):
				_, _ = fmt.Fprint(writer, "221 Bye\r\n")
				_ = writer.Flush()
				return
			default:
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			}
			_ = writer.Flush()
		}
	}()

	client, err := smtp.NewClient(clientConn, "localhost")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Hello("localhost"); err != nil {
		t.Fatalf("hello: %v", err)
	}

	task := emailpkg.Task{
		Recipient: parser.Recipient{Email: "primary@example.com"},
		Subject:   "Test",
		Body:      "Hello",
		CC:        []string{"primary@example.com", "cc1@example.com", "cc2@example.com"},
		BCC:       []string{"cc1@example.com", "bcc1@example.com", "primary@example.com"},
	}

	cfg := config.SMTPConfig{From: "sender@example.com"}
	if err := emailpkg.SendWithClient(client, cfg, task); err != nil {
		t.Fatalf("SendWithClient: %v", err)
	}
	if err := client.Quit(); err != nil {
		t.Fatalf("quit: %v", err)
	}
	wg.Wait()

	expected := []string{"primary@example.com", "cc1@example.com", "cc2@example.com", "bcc1@example.com"}
	if !reflect.DeepEqual(rcpts, expected) {
		t.Fatalf("RCPT order mismatch: got %v, want %v", rcpts, expected)
	}

	msg, err := mail.ReadMessage(strings.NewReader(data.String()))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	if got := msg.Header.Get("Cc"); got != "cc1@example.com, cc2@example.com" {
		t.Errorf("unexpected Cc header: %q", got)
	}
	if strings.Contains(data.String(), "bcc1@example.com") {
		t.Errorf("BCC address leaked in message data")
	}
}

func TestSendWithClient_SanitizeAndDedupRecipients(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	var (
		rcpts []string
		data  strings.Builder
		wg    sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { _ = serverConn.Close() }()
		reader := bufio.NewReader(serverConn)
		writer := bufio.NewWriter(serverConn)
		_, _ = fmt.Fprint(writer, "220 mock SMTP ready\r\n")
		_ = writer.Flush()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(cmd, "HELO") || strings.HasPrefix(cmd, "EHLO"):
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "MAIL FROM"):
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "RCPT TO"):
				addr := strings.TrimSpace(strings.Trim(strings.TrimPrefix(cmd, "RCPT TO:"), "<>"))
				rcpts = append(rcpts, addr)
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "DATA"):
				_, _ = fmt.Fprint(writer, "354 End data with <CRLF>.<CRLF>\r\n")
				_ = writer.Flush()
				for {
					l, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						_, _ = fmt.Fprint(writer, "250 OK\r\n")
						break
					}
					data.WriteString(l)
				}
			case strings.HasPrefix(cmd, "QUIT"):
				_, _ = fmt.Fprint(writer, "221 Bye\r\n")
				_ = writer.Flush()
				return
			default:
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			}
			_ = writer.Flush()
		}
	}()

	client, err := smtp.NewClient(clientConn, "localhost")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Hello("localhost"); err != nil {
		t.Fatalf("hello: %v", err)
	}

	task := emailpkg.Task{
		Recipient: parser.Recipient{Email: "to@example.com"},
		Subject:   "Test",
		Body:      "Hello",
		CC: []string{
			"",
			"  ",
			"to@example.com",
			"cc1@example.com",
			"CC1@example.com",
			"dup@example.com",
			"Dup@example.com",
		},
		BCC: []string{
			"",
			"   ",
			"to@example.com",
			"cc1@example.com",
			"bcc1@example.com",
			"BCC1@example.com",
			"dup@example.com",
			"Dup@example.com",
		},
	}

	cfg := config.SMTPConfig{From: "sender@example.com"}
	if err := emailpkg.SendWithClient(client, cfg, task); err != nil {
		t.Fatalf("SendWithClient: %v", err)
	}
	if err := client.Quit(); err != nil {
		t.Fatalf("quit: %v", err)
	}
	wg.Wait()

	expected := []string{"to@example.com", "cc1@example.com", "dup@example.com", "bcc1@example.com"}
	if !reflect.DeepEqual(rcpts, expected) {
		t.Fatalf("RCPT list mismatch: got %v, want %v", rcpts, expected)
	}

	msg, err := mail.ReadMessage(strings.NewReader(data.String()))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}

	if got := msg.Header.Get("Cc"); got != "cc1@example.com, dup@example.com" {
		t.Errorf("unexpected Cc header: %q", got)
	}

	if strings.Contains(strings.ToLower(data.String()), "bcc1@example.com") {
		t.Errorf("BCC address leaked in message data")
	}

	ccList := []string{}
	for _, a := range strings.Split(msg.Header.Get("Cc"), ",") {
		a = strings.TrimSpace(a)
		if a != "" {
			ccList = append(ccList, a)
		}
	}
	if !reflect.DeepEqual(rcpts[1:1+len(ccList)], ccList) {
		t.Errorf("envelope CCs mismatch RCPTs: got %v, want %v", rcpts[1:1+len(ccList)], ccList)
	}
	if bccEnv := rcpts[1+len(ccList):]; len(bccEnv) != 1 || bccEnv[0] != "bcc1@example.com" {
		t.Errorf("unexpected BCC envelope addresses: %v", bccEnv)
	}
}
func TestSendWithClient_AllRecipientPaths(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	var (
		rcpts []string
		data  strings.Builder
		wg    sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { _ = serverConn.Close() }()
		reader := bufio.NewReader(serverConn)
		writer := bufio.NewWriter(serverConn)
		_, _ = fmt.Fprint(writer, "220 mock SMTP ready\r\n")
		_ = writer.Flush()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(cmd, "HELO") || strings.HasPrefix(cmd, "EHLO"):
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "MAIL FROM"):
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "RCPT TO"):
				addr := strings.TrimSpace(strings.Trim(strings.TrimPrefix(cmd, "RCPT TO:"), "<>"))
				rcpts = append(rcpts, addr)
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			case strings.HasPrefix(cmd, "DATA"):
				_, _ = fmt.Fprint(writer, "354 End data with <CRLF>.<CRLF>\r\n")
				_ = writer.Flush()
				for {
					l, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						_, _ = fmt.Fprint(writer, "250 OK\r\n")
						break
					}
					data.WriteString(l)
				}
			case strings.HasPrefix(cmd, "QUIT"):
				_, _ = fmt.Fprint(writer, "221 Bye\r\n")
				_ = writer.Flush()
				return
			default:
				_, _ = fmt.Fprint(writer, "250 OK\r\n")
			}
			_ = writer.Flush()
		}
	}()

	client, err := smtp.NewClient(clientConn, "localhost")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Hello("localhost"); err != nil {
		t.Fatalf("hello: %v", err)
	}

	task := emailpkg.Task{
		Recipient: parser.Recipient{Email: "To@Example.com"},
		Subject:   "Test",
		Body:      "Hello",
		CC: []string{
			"to@example.com",
			"TO@EXAMPLE.COM",
			"Cc@example.com",
			"CC@Example.com",
		},
		BCC: []string{
			"Cc@example.com",
			"bcc@example.com",
			"BCC@EXAMPLE.com",
			"to@EXAMPLE.com",
		},
	}

	cfg := config.SMTPConfig{From: "sender@example.com"}
	if err := emailpkg.SendWithClient(client, cfg, task); err != nil {
		t.Fatalf("SendWithClient: %v", err)
	}
	if err := client.Quit(); err != nil {
		t.Fatalf("quit: %v", err)
	}
	wg.Wait()

	expected := []string{"To@Example.com", "Cc@example.com", "bcc@example.com"}
	if !reflect.DeepEqual(rcpts, expected) {
		t.Fatalf("RCPT list mismatch: got %v, want %v", rcpts, expected)
	}
	if len(rcpts) == 0 || !strings.EqualFold(rcpts[0], task.Recipient.Email) {
		t.Fatalf("primary recipient not first: %v", rcpts)
	}
	seen := map[string]struct{}{}
	for _, r := range rcpts {
		key := strings.ToLower(r)
		if _, ok := seen[key]; ok {
			t.Errorf("duplicate RCPT TO for %q", r)
		}
		seen[key] = struct{}{}
	}

	msg, err := mail.ReadMessage(strings.NewReader(data.String()))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	if got := msg.Header.Get("Cc"); got != "Cc@example.com" {
		t.Errorf("unexpected Cc header: %q", got)
	}
	if strings.Contains(strings.ToLower(data.String()), "bcc@example.com") {
		t.Errorf("BCC address leaked in message data")
	}
}
