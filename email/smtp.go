package email

import (
	"crypto/tls"
	"fmt"
	"mailgrid/config"
	"net/smtp"
)

// ConnectSMTP establishes a persistent, authenticated SMTP client with TLS.
func ConnectSMTP(cfg config.SMTPConfig) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}
	if ok, _ := conn.Extension("STARTTLS"); ok {
		tlsconfig := &tls.Config{ServerName: cfg.Host}
		if err = conn.StartTLS(tlsconfig); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("starttls error: %w", err)
		}
	}
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err = conn.Auth(auth); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("auth error: %w", err)
	}
	return conn, nil
}
