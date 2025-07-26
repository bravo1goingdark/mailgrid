package logger

import (
	"os"
	"path/filepath"
	"testing"

	"mailgrid/logger"
)

func TestLogSuccessAndFailure(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	logger.LogSuccess("a@b.com", "hello")
	logger.LogFailure("b@b.com", "bye")

	successPath := filepath.Join(dir, "success.csv")
	if _, err := os.Stat(successPath); err != nil {
		t.Fatalf("success.csv not created: %v", err)
	}
	data, err := os.ReadFile(successPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Error("success.csv empty")
	}

	failPath := filepath.Join(dir, "failed.csv")
	if _, err := os.Stat(failPath); err != nil {
		t.Fatalf("failed.csv not created: %v", err)
	}
	data, err = os.ReadFile(failPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Error("failed.csv empty")
	}
}
