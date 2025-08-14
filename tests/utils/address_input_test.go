package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bravo1goingdark/mailgrid/utils/valid"
)

func TestParseAddressInputInline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple", "a@a.com,b@b.com", []string{"a@a.com", "b@b.com"}},
		{"with empty and spaces", " a@a.com , , b@b.com ,", []string{"a@a.com", "b@b.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := valid.ParseAddressInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v got %v", tt.expected, got)
			}
		})
	}
}

func TestParseAddressInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "addresses.txt")
	content := "a@a.com\n\nb@b.com\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	expected := []string{"a@a.com", "b@b.com"}
	got, err := valid.ParseAddressInput(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v got %v", expected, got)
	}
}

func TestParseAddressInputFileError(t *testing.T) {
	_, err := valid.ParseAddressInput("nonexistent.txt")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}
