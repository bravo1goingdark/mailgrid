package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")

	testConfig := map[string]interface{}{
		"host":     "smtp.example.com",
		"port":     587,
		"username": "test@example.com",
		"password": "testpassword",
		"from":     "test@example.com",
	}

	// Write test config to file
	configData, err := json.Marshal(map[string]interface{}{
		"smtp": testConfig,
	})
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.SMTP.Host != "smtp.example.com" {
		t.Errorf("Expected host 'smtp.example.com', got '%s'", config.SMTP.Host)
	}

	if config.SMTP.Port != 587 {
		t.Errorf("Expected port 587, got %d", config.SMTP.Port)
	}

	if config.SMTP.Username != "test@example.com" {
		t.Errorf("Expected username 'test@example.com', got '%s'", config.SMTP.Username)
	}
}

func TestLoadConfigNonExistentFile(t *testing.T) {
	_, err := LoadConfig("non_existent_file.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid_config.json")

	// Write invalid JSON
	err := os.WriteFile(configFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err = LoadConfig(configFile)
	if err == nil {
		t.Error("Expected error when loading invalid JSON config file")
	}
}