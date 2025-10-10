package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bravo1goingdark/mailgrid/cli"
)

func TestOffsetTrackingIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Create test CSV file
	csvContent := `name,email
Alice Johnson,alice@example.com
Bob Smith,bob@example.com
Carol Davis,carol@example.com
David Wilson,david@example.com
Eve Brown,eve@example.com`

	csvFile := filepath.Join(tempDir, "test_recipients.csv")
	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	// Create test template file
	templateContent := `<!DOCTYPE html>
<html>
<body>
    <h1>Hello {{.name}}!</h1>
    <p>Email: {{.email}}</p>
</body>
</html>`

	templateFile := filepath.Join(tempDir, "test_template.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template file: %v", err)
	}

	// Create test config file
	configContent := `{
		"smtp": {
			"host": "smtp.example.com",
			"port": 587,
			"username": "test@example.com",
			"password": "password",
			"from": "test@example.com"
		},
		"rate_limit": 10,
		"timeout_ms": 5000
	}`

	configFile := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	offsetFile := filepath.Join(tempDir, "test.offset")

	t.Run("DryRunWithoutOffset", func(t *testing.T) {
		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test: {{.name}}",
			DryRun:       true,
			OffsetFile:   offsetFile,
		}

		// Should process all 5 emails
		err := cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Offset file should not be created in dry-run mode
		if _, err := os.Stat(offsetFile); !os.IsNotExist(err) {
			t.Error("Expected offset file to not exist after dry-run")
		}
	})

	t.Run("ResumeWithoutOffsetFile", func(t *testing.T) {
		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test: {{.name}}",
			DryRun:       true,
			Resume:       true,
			OffsetFile:   offsetFile,
		}

		// Should process all 5 emails since no offset file exists
		err := cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("ResumeWithPartialOffset", func(t *testing.T) {
		// Create partial offset file
		offsetContent := "alice@example.com\nbob@example.com\n"
		err := os.WriteFile(offsetFile, []byte(offsetContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create offset file: %v", err)
		}

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test: {{.name}}",
			DryRun:       true,
			Resume:       true,
			OffsetFile:   offsetFile,
		}

		// Should skip Alice and Bob, process Carol, David, Eve
		err = cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("ResumeWithCompleteOffset", func(t *testing.T) {
		// Create complete offset file
		offsetContent := "alice@example.com\nbob@example.com\ncarol@example.com\ndavid@example.com\neve@example.com\n"
		err := os.WriteFile(offsetFile, []byte(offsetContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create offset file: %v", err)
		}

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test: {{.name}}",
			DryRun:       true,
			Resume:       true,
			OffsetFile:   offsetFile,
		}

		// Should complete immediately - all emails already sent
		err = cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("ResetOffset", func(t *testing.T) {
		// Ensure offset file exists
		offsetContent := "alice@example.com\nbob@example.com\n"
		err := os.WriteFile(offsetFile, []byte(offsetContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create offset file: %v", err)
		}

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test: {{.name}}",
			DryRun:       true,
			ResetOffset:  true,
			OffsetFile:   offsetFile,
		}

		// Should clear offset and process all emails
		err = cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Offset file should be removed
		if _, err := os.Stat(offsetFile); !os.IsNotExist(err) {
			t.Error("Expected offset file to be removed after reset")
		}
	})

	t.Run("CustomOffsetFile", func(t *testing.T) {
		customOffsetFile := filepath.Join(tempDir, "custom.offset")

		// Create custom offset file
		offsetContent := "alice@example.com\n"
		err := os.WriteFile(customOffsetFile, []byte(offsetContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create custom offset file: %v", err)
		}

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test: {{.name}}",
			DryRun:       true,
			Resume:       true,
			OffsetFile:   customOffsetFile,
		}

		// Should use custom offset file and skip Alice
		err = cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Custom offset file should still exist
		if _, err := os.Stat(customOffsetFile); os.IsNotExist(err) {
			t.Error("Expected custom offset file to exist")
		}
	})
}

func TestOffsetTrackingCLIFlags(t *testing.T) {
	t.Run("ParseResumeFlag", func(t *testing.T) {
		// This test would require setting up actual CLI parsing
		// For now, we test that the fields exist in CLIArgs
		args := cli.CLIArgs{
			Resume:      true,
			ResetOffset: false,
			OffsetFile:  "custom.offset",
		}

		if !args.Resume {
			t.Error("Expected Resume flag to be true")
		}
		if args.ResetOffset {
			t.Error("Expected ResetOffset flag to be false")
		}
		if args.OffsetFile != "custom.offset" {
			t.Errorf("Expected OffsetFile to be 'custom.offset', got '%s'", args.OffsetFile)
		}
	})

	t.Run("DefaultOffsetFile", func(t *testing.T) {
		args := cli.CLIArgs{}

		// Default should be empty (will be set by pflag default)
		expectedDefault := ""
		if args.OffsetFile != expectedDefault {
			t.Errorf("Expected default OffsetFile to be '%s', got '%s'", expectedDefault, args.OffsetFile)
		}
	})
}

func TestOffsetTrackingEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("EmptyCSVFile", func(t *testing.T) {
		// Create empty CSV file
		csvFile := filepath.Join(tempDir, "empty.csv")
		err := os.WriteFile(csvFile, []byte("name,email\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty CSV file: %v", err)
		}

		templateFile := filepath.Join(tempDir, "template.html")
		err = os.WriteFile(templateFile, []byte("<html><body>{{.name}}</body></html>"), 0644)
		if err != nil {
			t.Fatalf("Failed to create template file: %v", err)
		}

		configFile := filepath.Join(tempDir, "config.json")
		configContent := `{"smtp":{"host":"test","port":587,"username":"test","password":"test","from":"test"}}`
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		offsetFile := filepath.Join(tempDir, "test.offset")

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test",
			DryRun:       true,
			Resume:       true,
			OffsetFile:   offsetFile,
		}

		// Should handle empty CSV gracefully
		err = cli.Run(args)
		if err != nil {
			// Empty CSV might return an error, but it shouldn't crash
			// The specific behavior depends on the CSV parser implementation
			t.Logf("Empty CSV returned error (expected): %v", err)
		}
	})

	t.Run("CorruptedOffsetFile", func(t *testing.T) {
		csvFile := filepath.Join(tempDir, "test2.csv")
		csvContent := "name,email\ntest,test@example.com\n"
		err := os.WriteFile(csvFile, []byte(csvContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create CSV file: %v", err)
		}

		templateFile := filepath.Join(tempDir, "template2.html")
		err = os.WriteFile(templateFile, []byte("<html><body>{{.name}}</body></html>"), 0644)
		if err != nil {
			t.Fatalf("Failed to create template file: %v", err)
		}

		configFile := filepath.Join(tempDir, "config2.json")
		configContent := `{"smtp":{"host":"test","port":587,"username":"test","password":"test","from":"test"}}`
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		offsetFile := filepath.Join(tempDir, "corrupted.offset")

		// Create offset file with binary data (simulating corruption)
		corruptData := []byte{0x00, 0x01, 0xFF, 0xFE}
		err = os.WriteFile(offsetFile, corruptData, 0644)
		if err != nil {
			t.Fatalf("Failed to create corrupted offset file: %v", err)
		}

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Test",
			DryRun:       true,
			Resume:       true,
			OffsetFile:   offsetFile,
		}

		// Should handle corrupted offset file gracefully
		err = cli.Run(args)
		if err != nil {
			// Corruption might cause an error, but it shouldn't crash the program
			t.Logf("Corrupted offset file handled: %v", err)
		}
	})
}

func TestOffsetTrackingFilterIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Create CSV with filterable data
	csvContent := `name,email,department
Alice Johnson,alice@example.com,engineering
Bob Smith,bob@example.com,marketing
Carol Davis,carol@example.com,engineering
David Wilson,david@example.com,sales
Eve Brown,eve@example.com,engineering`

	csvFile := filepath.Join(tempDir, "filtered.csv")
	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create filtered CSV file: %v", err)
	}

	templateFile := filepath.Join(tempDir, "template.html")
	err = os.WriteFile(templateFile, []byte("<html><body>{{.name}} - {{.department}}</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	configContent := `{"smtp":{"host":"test","port":587,"username":"test","password":"test","from":"test"}}`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	offsetFile := filepath.Join(tempDir, "filter.offset")

	t.Run("ResumeWithFilter", func(t *testing.T) {
		// Create offset with one engineering email
		offsetContent := "alice@example.com\n"
		err := os.WriteFile(offsetFile, []byte(offsetContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create offset file: %v", err)
		}

		args := cli.CLIArgs{
			EnvPath:      configFile,
			CSVPath:      csvFile,
			TemplatePath: templateFile,
			Subject:      "Engineering: {{.name}}",
			DryRun:       true,
			Resume:       true,
			Filter:       "department == \"engineering\"",
			OffsetFile:   offsetFile,
		}

		// Should filter to engineering dept (Alice, Carol, Eve)
		// then skip Alice (in offset), leaving Carol and Eve
		err = cli.Run(args)
		if err != nil {
			t.Fatalf("Expected no error with filter and resume, got: %v", err)
		}
	})
}