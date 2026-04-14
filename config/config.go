package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type SMTPConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	From        string `json:"from"`
	TLSCertFile string `json:"tls_cert_file,omitempty"` // Path to custom CA certificate
	TLSKeyFile  string `json:"tls_key_file,omitempty"`  // Path to client certificate
	InsecureTLS bool   `json:"insecure_tls,omitempty"`  // Skip TLS verification (use with caution)

	// DialTimeout overrides the default 10-second TCP connect timeout.
	// Zero means use the default.
	DialTimeout time.Duration `json:"-"` // set from CLI flag, not the JSON file
}

type AppConfig struct {
	SMTP      SMTPConfig `json:"smtp"`
	TimeoutMs int        `json:"timeout_ms"` // smtp timeout in milliseconds
}

// Validate checks that all required SMTP fields are present.
// Call this immediately after LoadConfig to catch misconfigured deployments
// before any recipients are loaded or any SMTP connections are attempted.
func Validate(cfg SMTPConfig) error {
	if cfg.Host == "" {
		return fmt.Errorf("smtp.host is required")
	}
	if cfg.Port == 0 {
		return fmt.Errorf("smtp.port is required")
	}
	if cfg.Username == "" {
		return fmt.Errorf("smtp.username is required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("smtp.password is required")
	}
	if cfg.From == "" {
		return fmt.Errorf("smtp.from is required")
	}
	return nil
}

// LoadConfig reads JSON config from disk and returns a parsed AppConfig.
// It never terminates the process; callers should handle returned errors.
func LoadConfig(path string) (*AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %q: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log to stderr since callers may override the returned error
			fmt.Fprintf(os.Stderr, "Warning: failed to close config file: %v\n", closeErr)
		}
	}()

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config JSON: %w", err)
	}
	return &cfg, nil
}
