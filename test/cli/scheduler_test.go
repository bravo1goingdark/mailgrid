//go:build integration
// +build integration

package cli_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/cli"
	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/scheduler"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_Integration(t *testing.T) {
	// Start a mock SMTP server
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	require.NoError(t, server.Start())
	defer server.Stop()

	// Create a temporary config file pointing at the mock SMTP
	configContent := fmt.Sprintf(`{
		"smtp": {
			"host": "%s",
			"port": %d,
			"username": "",
			"password": "",
			"from": "test@example.com"
		}
	}`, server.HostAddress, server.Port)
	configFile, err := os.CreateTemp("", "config.json")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())
	_, err = configFile.WriteString(configContent)
	require.NoError(t, err)
	configFile.Close()

	// Load SMTP config for handler use
	smtpCfg, err := config.LoadConfig(configFile.Name())
	require.NoError(t, err)

	// Create a temporary BoltDB
	tmpDB, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	db, err := database.NewDB(tmpDB.Name())
	require.NoError(t, err)
	defer db.Close()

	log := logger.New("test-scheduler")
	sched := scheduler.NewScheduler(db, log)

	// Channel to signal job completion
	done := make(chan error, 1)

	args := types.CLIArgs{
		EnvPath: configFile.Name(),
		To:      "recipient@example.com",
		Subject: "Test Email",
		Text:    "This is a test email.",
	}

	job, err := scheduler.NewJob(args, time.Now().Add(200*time.Millisecond), "", "")
	require.NoError(t, err)

	handler := func(j types.Job) error {
		var a types.CLIArgs
		if err := cli.DecodeJobArgs(j, &a); err != nil {
			done <- err
			return err
		}
		cliArgs := cli.CLIArgs{
			EnvPath:      a.EnvPath,
			To:           a.To,
			Subject:      a.Subject,
			Text:         a.Text,
			TemplatePath: a.Template,
		}
		err := cli.SendSingleEmail(cliArgs, smtpCfg.SMTP)
		done <- err
		return err
	}

	err = sched.AddJob(job, handler)
	require.NoError(t, err)

	// Wait for job execution
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Job did not execute in time")
	}

	// Verify job completed
	var jobStatus string
	deadline := time.Now().Add(3 * time.Second)
	for {
		jobs, err := sched.ListJobs()
		require.NoError(t, err)
		if len(jobs) == 1 && (jobs[0].Status == "done" || jobs[0].Status == "pending") {
			jobStatus = jobs[0].Status
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Job did not complete in time; last status: %v", jobs)
		}
		time.Sleep(50 * time.Millisecond)
	}
	assert.Equal(t, "done", jobStatus)
}
