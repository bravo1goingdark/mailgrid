package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

type AppConfig struct {
	SMTP      SMTPConfig `json:"smtp"`
	RateLimit int        `json:"rate_limit"` // emails you can send per second
	TimeoutMs int        `json:"timeout_ms"` // smtp timeout
}

// LoadConfig reads JSON config from disk and returns a parsed AppConfig.
// It never terminates the process; callers should handle returned errors.
func LoadConfig(path string) (*AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config JSON: %w", err)
	}
	return &cfg, nil
}
