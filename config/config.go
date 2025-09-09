package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type SMTPConfig struct {
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	Username          string        `json:"username"`
	Password          string        `json:"password"`
	From              string        `json:"from"`
	UseTLS            bool          `json:"use_tls"`
	InsecureSkipVerify bool         `json:"insecure_skip_verify"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
}

type LogConfig struct {
	Level      string `json:"level"` // debug, info, warn, error
	Format     string `json:"format"` // json, text
	File       string `json:"file,omitempty"` // log file path
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"` // days
}

type MetricsConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type AppConfig struct {
	SMTP       SMTPConfig    `json:"smtp"`
	RateLimit  int           `json:"rate_limit"`  // emails per second
	BurstLimit int           `json:"burst_limit"` // burst size
	TimeoutMs  int           `json:"timeout_ms"` // smtp timeout (deprecated)
	Log        LogConfig     `json:"log"`
	Metrics    MetricsConfig `json:"metrics"`
	
	// Security settings
	MaxAttachmentSize int64 `json:"max_attachment_size"` // bytes
	MaxConcurrency    int   `json:"max_concurrency"`
	MaxBatchSize      int   `json:"max_batch_size"`
	MaxRetries        int   `json:"max_retries"`
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
			fmt.Printf("Warning: failed to close config file: %v\n", closeErr)
		}
	}()

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config JSON: %w", err)
	}
	
	// Apply defaults and validate
	if err := cfg.setDefaults(); err != nil {
		return nil, fmt.Errorf("apply config defaults: %w", err)
	}
	
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	
	return &cfg, nil
}

// setDefaults applies sensible defaults to missing config values
func (c *AppConfig) setDefaults() error {
	// SMTP defaults
	if c.SMTP.ConnectionTimeout == 0 {
		c.SMTP.ConnectionTimeout = 10 * time.Second
	}
	if c.SMTP.ReadTimeout == 0 {
		c.SMTP.ReadTimeout = 30 * time.Second
	}
	if c.SMTP.WriteTimeout == 0 {
		c.SMTP.WriteTimeout = 30 * time.Second
	}
	if c.SMTP.Port == 0 {
		if c.SMTP.UseTLS {
			c.SMTP.Port = 587
		} else {
			c.SMTP.Port = 25
		}
	}
	
	// Rate limiting defaults
	if c.RateLimit == 0 {
		c.RateLimit = 10 // 10 emails per second
	}
	if c.BurstLimit == 0 {
		c.BurstLimit = c.RateLimit * 2
	}
	
	// Security defaults
	if c.MaxAttachmentSize == 0 {
		c.MaxAttachmentSize = 10 * 1024 * 1024 // 10 MB
	}
	if c.MaxConcurrency == 0 {
		c.MaxConcurrency = 10
	}
	if c.MaxBatchSize == 0 {
		c.MaxBatchSize = 50
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	
	// Logging defaults
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
	if c.Log.MaxSize == 0 {
		c.Log.MaxSize = 100 // 100 MB
	}
	if c.Log.MaxBackups == 0 {
		c.Log.MaxBackups = 3
	}
	if c.Log.MaxAge == 0 {
		c.Log.MaxAge = 28 // 28 days
	}
	
	// Metrics defaults
	if c.Metrics.Port == 0 {
		c.Metrics.Port = 8090
	}
	
	return nil
}

// validate checks required config fields and limits
func (c *AppConfig) validate() error {
	if c.SMTP.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if c.SMTP.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if c.SMTP.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	if c.SMTP.From == "" {
		return fmt.Errorf("SMTP from address is required")
	}
	
	// Validate limits
	if c.RateLimit < 0 {
		return fmt.Errorf("rate_limit cannot be negative")
	}
	if c.BurstLimit < 0 {
		return fmt.Errorf("burst_limit cannot be negative")
	}
	if c.MaxConcurrency <= 0 || c.MaxConcurrency > 100 {
		return fmt.Errorf("max_concurrency must be between 1 and 100")
	}
	if c.MaxBatchSize <= 0 || c.MaxBatchSize > 1000 {
		return fmt.Errorf("max_batch_size must be between 1 and 1000")
	}
	if c.MaxRetries < 0 || c.MaxRetries > 10 {
		return fmt.Errorf("max_retries must be between 0 and 10")
	}
	if c.MaxAttachmentSize <= 0 || c.MaxAttachmentSize > 100*1024*1024 {
		return fmt.Errorf("max_attachment_size must be between 1 and 100MB")
	}
	
	return nil
}
