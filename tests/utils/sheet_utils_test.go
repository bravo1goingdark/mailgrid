package utils

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/utils"
)

func TestExtractSheetInfo(t *testing.T) {
	id, gid, err := utils.ExtractSheetInfo("https://docs.google.com/spreadsheets/d/abc123/edit#gid=789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "abc123" || gid != "789" {
		t.Errorf("unexpected values: id=%s gid=%s", id, gid)
	}
}

func TestExtractSheetInfo_DefaultGID(t *testing.T) {
	_, gid, err := utils.ExtractSheetInfo("https://docs.google.com/spreadsheets/d/abc123/edit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gid != "0" {
		t.Errorf("expected gid 0, got %s", gid)
	}
}

func TestExtractSheetInfo_Invalid(t *testing.T) {
	_, _, err := utils.ExtractSheetInfo("invalid")
	if err == nil {
		t.Fatal("expected error for invalid url")
	}
}
