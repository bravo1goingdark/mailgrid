package config

import (
	"encoding/json"
	"os"
	"testing"

	"mailgrid/config"
)

func TestLoadConfig(t *testing.T) {
	cfg := config.AppConfig{SMTP: config.SMTPConfig{Host: "smtp", Port: 25, Username: "u", Password: "p", From: "me@test"}, RateLimit: 1, TimeoutMs: 1000}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := os.CreateTemp(t.TempDir(), "cfg*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.Write(data); err != nil {
		t.Fatal(err)
	}
	err = tmp.Close()
	if err != nil {
		return
	}

	loaded, err := config.LoadConfig(tmp.Name())
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if loaded.SMTP.Host != "smtp" || loaded.RateLimit != 1 {
		t.Errorf("unexpected config: %+v", loaded)
	}
}
