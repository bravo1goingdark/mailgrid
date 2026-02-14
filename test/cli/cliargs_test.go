package cli_test

import (
	"os"
	"testing"

	"github.com/bravo1goingdark/mailgrid/cli"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected cli.CLIArgs
	}{
		{
			name: "basic email flags",
			args: []string{
				"--env", "config.json",
				"--csv", "recipients.csv",
				"--template", "template.html",
				"--subject", "Test Subject",
				"--concurrency", "5",
			},
			expected: cli.CLIArgs{
				EnvPath:      "config.json",
				CSVPath:      "recipients.csv",
				TemplatePath: "template.html",
				Subject:      "Test Subject",
				Concurrency:  5,
				DryRun:       false,
				ShowPreview:  false,
				PreviewPort:  8080,
				RetryLimit:   1,
				BatchSize:    1,
				JobRetries:   3,
				MonitorPort:  9091,
				Attachments:  []string{},
			},
		},
		{
			name: "scheduling flags",
			args: []string{
				"--schedule-at", "2025-12-01T10:00:00Z",
				"--interval", "1h",
				"--cron", "0 9 * * 1",
				"--job-retries", "5",
			},
			expected: cli.CLIArgs{
				Subject:     "Test Email from Mailgrid",
				Concurrency: 1,
				ScheduleAt:  "2025-12-01T10:00:00Z",
				Interval:    "1h",
				Cron:        "0 9 * * 1",
				JobRetries:  5,
				RetryLimit:  1,
				BatchSize:   1,
				PreviewPort: 8080,
				MonitorPort: 9091,
				Attachments: []string{},
			},
		},
		{
			name: "boolean flags",
			args: []string{
				"--dry-run",
				"--preview",
				"--jobs-list",
				"--scheduler-run",
			},
			expected: cli.CLIArgs{
				Subject:      "Test Email from Mailgrid",
				Concurrency:  1,
				DryRun:       true,
				ShowPreview:  true,
				ListJobs:     true,
				SchedulerRun: true,
				RetryLimit:   1,
				BatchSize:    1,
				PreviewPort:  8080,
				JobRetries:   3,
				MonitorPort:  9091,
				Attachments:  []string{},
			},
		},
		{
			name: "attachments and addresses",
			args: []string{
				"--attach", "file1.pdf",
				"--attach", "file2.jpg",
				"--cc", "cc@example.com",
				"--bcc", "bcc@example.com",
				"--to", "recipient@example.com",
				"--text", "Hello world",
			},
			expected: cli.CLIArgs{
				Subject:     "Test Email from Mailgrid",
				Concurrency: 1,
				Attachments: []string{"file1.pdf", "file2.jpg"},
				Cc:          "cc@example.com",
				Bcc:         "bcc@example.com",
				To:          "recipient@example.com",
				Text:        "Hello world",
				RetryLimit:  1,
				BatchSize:   1,
				PreviewPort: 8080,
				JobRetries:  3,
				MonitorPort: 9091,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			oldArgs := os.Args
			os.Args = append([]string{"mailgrid"}, tt.args...)
			defer func() { os.Args = oldArgs }()

			result := cli.ParseFlags()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCLIArgs_DefaultValues(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	oldArgs := os.Args
	os.Args = []string{"mailgrid"}
	defer func() { os.Args = oldArgs }()

	result := cli.ParseFlags()

	assert.Equal(t, "", result.EnvPath)
	assert.Equal(t, "", result.CSVPath)
	assert.Equal(t, "", result.TemplatePath)
	assert.Equal(t, "Test Email from Mailgrid", result.Subject)
	assert.Equal(t, false, result.DryRun)
	assert.Equal(t, false, result.ShowPreview)
	assert.Equal(t, 8080, result.PreviewPort)
	assert.Equal(t, 1, result.Concurrency)
	assert.Equal(t, 1, result.RetryLimit)
	assert.Equal(t, 1, result.BatchSize)
	assert.Equal(t, 3, result.JobRetries)
	assert.Equal(t, false, result.Monitor)
	assert.Equal(t, 9091, result.MonitorPort)
}

func TestCLIArgs_ShortFlags(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	oldArgs := os.Args
	os.Args = []string{
		"mailgrid",
		"-t", "template.html",
		"-s", "Test Subject",
		"-p",
		"-c", "3",
		"-r", "2",
	}
	defer func() { os.Args = oldArgs }()

	result := cli.ParseFlags()

	assert.Equal(t, "template.html", result.TemplatePath)
	assert.Equal(t, "Test Subject", result.Subject)
	assert.Equal(t, true, result.ShowPreview)
	assert.Equal(t, 3, result.Concurrency)
	assert.Equal(t, 2, result.RetryLimit)
}
