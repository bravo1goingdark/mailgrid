package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/bravo1goingdark/mailgrid/config"
	"net"
	"net/smtp"
	"time"
)

// ConnectSMTP establishes a persistent, authenticated SMTP client with TLS and context support.
func ConnectSMTP(cfg config.SMTPConfig) (*smtp.Client, error) {
	return ConnectSMTPWithContext(context.Background(), cfg)
}

// ConnectSMTPWithContext establishes a persistent, authenticated SMTP client with TLS and context support.
// This allows for proper cancellation during connection attempts.
func ConnectSMTPWithContext(ctx context.Context, cfg config.SMTPConfig) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Use context-aware dial with timeout
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("SMTP dial error: %w", err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SMTP client init error: %w", err)
	}

	// Check for context cancellation before proceeding
	if ctx.Err() != nil {
		client.Close()
		return nil, ctx.Err()
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsconfig := &tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}
		if err = client.StartTLS(tlsconfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("STARTTLS error: %w", err)
		}
	}

	// Check for context cancellation before auth
	if ctx.Err() != nil {
		client.Close()
		return nil, ctx.Err()
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err = client.Auth(auth); err != nil {
		client.Close()
		return nil, fmt.Errorf("SMTP auth error: %w", err)
	}

	return client, nil
}
