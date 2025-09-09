package email

import (
	"crypto/tls"
	"fmt"
	"github.com/bravo1goingdark/mailgrid/config"
	"net"
	"net/smtp"
	"time"
)

// ConnectSMTP establishes a persistent, authenticated SMTP client with TLS.
func ConnectSMTP(cfg config.SMTPConfig) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Use configured timeout or default
	connTimeout := cfg.ConnectionTimeout
	if connTimeout == 0 {
		connTimeout = 10 * time.Second
	}

	conn, err := net.DialTimeout("tcp", addr, connTimeout)
	if err != nil {
		return nil, fmt.Errorf("SMTP dial error: %w", err)
	}

	// Set read/write timeouts if configured
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if cfg.ReadTimeout > 0 {
			if err := tcpConn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout)); err != nil {
				conn.Close()
				return nil, fmt.Errorf("set read timeout: %w", err)
			}
		}
		if cfg.WriteTimeout > 0 {
			if err := tcpConn.SetWriteDeadline(time.Now().Add(cfg.WriteTimeout)); err != nil {
				conn.Close()
				return nil, fmt.Errorf("set write timeout: %w", err)
			}
		}
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SMTP client init error: %w", err)
	}

	// Always try STARTTLS if available (unless explicitly disabled)
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsconfig := &tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
		if err = client.StartTLS(tlsconfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("STARTTLS error: %w", err)
		}
	} else if cfg.UseTLS {
		// TLS required but not available
		client.Close()
		return nil, fmt.Errorf("TLS required but STARTTLS not supported by server")
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err = client.Auth(auth); err != nil {
		client.Close()
		return nil, fmt.Errorf("SMTP auth error: %w", err)
	}

	return client, nil
}
