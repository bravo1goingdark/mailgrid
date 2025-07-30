package config

import (
	"encoding/json"
	"log"
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

func LoadConfig(path string) (*AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Failed to close config file: %v", err)
		}
	}(file)

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Failed to decode config JSON: %v", err)
	}
	return &cfg, nil
}
