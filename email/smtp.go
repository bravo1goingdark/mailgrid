package email

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
)

// ConnectSMTP establishes a persistent, authenticated SMTP client with TLS and context support.
func ConnectSMTP(cfg config.SMTPConfig) (*smtp.Client, error) {
	return ConnectSMTPWithContext(context.Background(), cfg)
}

// ConnectSMTPWithContext establishes a persistent, authenticated SMTP client with TLS and context support.
// This allows for proper cancellation during connection attempts.
func ConnectSMTPWithContext(ctx context.Context, cfg config.SMTPConfig) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Use context-aware dial with configurable timeout (default 10s)
	dialTimeout := cfg.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = 10 * time.Second
	}
	dialer := &net.Dialer{Timeout: dialTimeout}
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
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			client.Close()
			return nil, err
		}
		if err = client.StartTLS(tlsConfig); err != nil {
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

// buildTLSConfig builds TLS configuration based on SMTP config options.
// Returns an error if explicitly configured cert/key files fail to load
// (previously these errors were silently swallowed).
func buildTLSConfig(cfg config.SMTPConfig) (*tls.Config, error) {
	if cfg.InsecureTLS {
		fmt.Println("SECURITY WARNING: TLS certificate verification is disabled (insecure_tls=true). " +
			"This connection is vulnerable to man-in-the-middle attacks.")
	}

	tlsConfig := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.InsecureTLS, //nolint:gosec // user explicitly set this
		MinVersion:         tls.VersionTLS12,
	}

	// Load custom CA certificate if provided.
	if cfg.TLSCertFile != "" {
		cert, err := os.ReadFile(cfg.TLSCertFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS CA cert %q: %w", cfg.TLSCertFile, err)
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cert)
		tlsConfig.RootCAs = certPool
	}

	// Load client certificate/key pair if both are provided.
	if cfg.TLSKeyFile != "" && cfg.TLSCertFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS client cert/key (%q / %q): %w", cfg.TLSCertFile, cfg.TLSKeyFile, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
