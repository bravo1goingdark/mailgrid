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

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("SMTP dial error: %w", err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("SMTP client init error: %w", err)
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsconfig := &tls.Config{ServerName: cfg.Host}
		if err = client.StartTLS(tlsconfig); err != nil {
			err := client.Close()
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("STARTTLS error: %w", err)
		}
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err = client.Auth(auth); err != nil {
		err := client.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("SMTP auth error: %w", err)
	}

	return client, nil
}
