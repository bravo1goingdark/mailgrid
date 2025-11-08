package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/bravo1goingdark/mailgrid/parser"
)

func TestParseCSVFromReader(t *testing.T) {
	csvData := "email,name,company\nuser1@example.com,Alice,Acme\nuser2@example.com,Bob,Widgets\n"
	r := strings.NewReader(csvData)

	recipients, err := parser.ParseCSVFromReader(r)
	if err != nil {
		t.Fatalf("ParseCSVFromReader error: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(recipients))
	}

	if recipients[0].Email != "user1@example.com" || recipients[0].Data["name"] != "Alice" || recipients[0].Data["company"] != "Acme" {
		t.Errorf("unexpected first recipient: %+v", recipients[0])
	}
}

func TestParseCSVFromReader_SkipsMalformedRows(t *testing.T) {
	csvData := strings.Join([]string{
		"email,name,company",
		"user1@example.com,Alice,Acme",
		"user2@example.com,Bob,Widgets",
		"badrow@example.com,OnlyName", // missing company field
		"too,many,fields,here",
	}, "\n") + "\n"

	recipients, err := parser.ParseCSVFromReader(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ParseCSVFromReader error: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(recipients))
	}

	if recipients[0].Email != "user1@example.com" {
		t.Errorf("unexpected first recipient: %+v", recipients[0])
	}
	if recipients[1].Email != "user2@example.com" {
		t.Errorf("unexpected second recipient: %+v", recipients[1])
	}
}

func TestParseCSV_MissingEmailColumn(t *testing.T) {
	csvData := "name,company\nAlice,Acme\n"
	r := strings.NewReader(csvData)
	_, err := parser.ParseCSVFromReader(r)
	if err == nil {
		t.Fatal("expected error for missing email column")
	}
}

func TestParseCSV(t *testing.T) {
	csvData := "email,name\nuser@example.com,Alice\n"
	tmp, err := os.CreateTemp(t.TempDir(), "recipients*.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString(csvData); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	recipients, err := parser.ParseCSV(tmp.Name())
	if err != nil {
		t.Fatalf("ParseCSV error: %v", err)
	}
	if len(recipients) != 1 || recipients[0].Email != "user@example.com" {
		t.Errorf("unexpected recipients: %+v", recipients)
	}
}
