package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestLoadConfigWithDefaults(t *testing.T) {
	// Create a minimal config file
	minimalConfig := `{
		"smtp": {
			"host": "smtp.example.com",
			"username": "test@example.com", 
			"password": "password123",
			"from": "test@example.com"
		}
	}`

	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(minimalConfig); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test defaults are applied
	if cfg.SMTP.Port == 0 {
		t.Error("Expected default SMTP port to be set")
	}
	if cfg.SMTP.ConnectionTimeout == 0 {
		t.Error("Expected default connection timeout to be set")
	}
	if cfg.RateLimit == 0 {
		t.Error("Expected default rate limit to be set")
	}
	if cfg.BurstLimit == 0 {
		t.Error("Expected default burst limit to be set")
	}
	if cfg.MaxConcurrency == 0 {
		t.Error("Expected default max concurrency to be set")
	}
	if cfg.Log.Level == "" {
		t.Error("Expected default log level to be set")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
		errorString string
	}{
		{
			name: "valid config",
			config: `{
				"smtp": {
					"host": "smtp.example.com",
					"username": "test@example.com",
					"password": "password123",
					"from": "test@example.com"
				}
			}`,
			expectError: false,
		},
		{
			name: "missing SMTP host",
			config: `{
				"smtp": {
					"username": "test@example.com",
					"password": "password123",
					"from": "test@example.com"
				}
			}`,
			expectError: true,
			errorString: "SMTP host is required",
		},
		{
			name: "missing SMTP username",
			config: `{
				"smtp": {
					"host": "smtp.example.com",
					"password": "password123",
					"from": "test@example.com"
				}
			}`,
			expectError: true,
			errorString: "SMTP username is required",
		},
		{
			name: "invalid concurrency",
			config: `{
				"smtp": {
					"host": "smtp.example.com",
					"username": "test@example.com",
					"password": "password123",
					"from": "test@example.com"
				},
				"max_concurrency": -1
			}`,
			expectError: true,
			errorString: "max_concurrency must be between 1 and 100",
		},
		{
			name: "invalid batch size",
			config: `{
				"smtp": {
					"host": "smtp.example.com",
					"username": "test@example.com",
					"password": "password123",
					"from": "test@example.com"
				},
				"max_batch_size": 2000
			}`,
			expectError: true,
			errorString: "max_batch_size must be between 1 and 1000",
		},
		{
			name: "negative rate limit",
			config: `{
				"smtp": {
					"host": "smtp.example.com",
					"username": "test@example.com",
					"password": "password123",
					"from": "test@example.com"
				},
				"rate_limit": -1
			}`,
			expectError: true,
			errorString: "rate_limit cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-config-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.config); err != nil {
				t.Fatal(err)
			}
			tmpFile.Close()

			_, err = LoadConfig(tmpFile.Name())
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorString != "" && !contains(err.Error(), tt.errorString) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSMTPPortDefaults(t *testing.T) {
	tests := []struct {
		name         string
		useTLS       bool
		expectedPort int
	}{
		{"TLS enabled", true, 587},
		{"TLS disabled", false, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AppConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.example.com",
					Username: "test@example.com",
					Password: "password123",
					From:     "test@example.com",
					UseTLS:   tt.useTLS,
				},
			}

			err := cfg.setDefaults()
			if err != nil {
				t.Fatalf("setDefaults failed: %v", err)
			}

			if cfg.SMTP.Port != tt.expectedPort {
				t.Errorf("Expected port %d for TLS=%v, got %d", tt.expectedPort, tt.useTLS, cfg.SMTP.Port)
			}
		})
	}
}

func TestTimeoutDefaults(t *testing.T) {
	cfg := &AppConfig{
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Username: "test@example.com",
			Password: "password123",
			From:     "test@example.com",
		},
	}

	err := cfg.setDefaults()
	if err != nil {
		t.Fatalf("setDefaults failed: %v", err)
	}

	if cfg.SMTP.ConnectionTimeout != 10*time.Second {
		t.Errorf("Expected connection timeout 10s, got %v", cfg.SMTP.ConnectionTimeout)
	}
	if cfg.SMTP.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", cfg.SMTP.ReadTimeout)
	}
	if cfg.SMTP.WriteTimeout != 30*time.Second {
		t.Errorf("Expected write timeout 30s, got %v", cfg.SMTP.WriteTimeout)
	}
}

func TestConfigJSONMarshal(t *testing.T) {
	cfg := AppConfig{
		SMTP: SMTPConfig{
			Host:              "smtp.example.com",
			Port:              587,
			Username:          "test@example.com",
			Password:          "password123",
			From:              "test@example.com",
			UseTLS:            true,
			ConnectionTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
		},
		RateLimit:         10,
		BurstLimit:        20,
		MaxConcurrency:    5,
		MaxBatchSize:      100,
		MaxAttachmentSize: 10 * 1024 * 1024,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var cfg2 AppConfig
	err = json.Unmarshal(data, &cfg2)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if cfg2.SMTP.Host != cfg.SMTP.Host {
		t.Errorf("Host mismatch after marshal/unmarshal: expected %q, got %q", cfg.SMTP.Host, cfg2.SMTP.Host)
	}
	if cfg2.RateLimit != cfg.RateLimit {
		t.Errorf("RateLimit mismatch after marshal/unmarshal: expected %d, got %d", cfg.RateLimit, cfg2.RateLimit)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
