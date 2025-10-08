//go:build integration
// +build integration

package cli_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/cli"
	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/scheduler"
	"github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_Integration(t *testing.T) {
	// Start a mock SMTP server
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	require.NoError(t, server.Start())
	defer server.Stop()

	// Create a temporary database file
	tmpfile, err := ioutil.TempFile("", "test.db")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create a new BoltDB client
	db, err := database.NewDB(tmpfile.Name())
	require.NoError(t, err)
	defer db.Close()

	// Create a new logger
	log := logger.New("test")

	// Create a new scheduler
	s := scheduler.NewScheduler(db, log)
	es := scheduler.NewEmailScheduler(s)

	// Create a new runner
	runner := cli.NewRunner(es)

	// Reattach handlers
	es.ReattachHandlers(runner.EmailJobHandler)

	// Create a temporary config file
	configContent := fmt.Sprintf(`{
		"smtp": {
			"host": "%s",
			"port": %d,
			"username": "",
			"password": "",
			"from": "test@example.com"
		}
	}`, server.HostAddress, server.Port)
	configFile, err := ioutil.TempFile("", "config.json")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())
	_, err = configFile.WriteString(configContent)
	require.NoError(t, err)
	configFile.Close()

	// Schedule an email to be sent in 1 second
	args := types.CLIArgs{
		EnvPath:    configFile.Name(),
		To:         "recipient@example.com",
		Subject:    "Test Email",
		Text:       "This is a test email.",
		ScheduleAt: time.Now().Add(1 * time.Second).Format(time.RFC3339),
	}
	err = runner.Run(context.Background(), args)
	require.NoError(t, err)

	// Wait for the email to be sent
	time.Sleep(2 * time.Second)

	// Verify that the email was sent
	messages := server.Messages()
	assert.Len(t, messages, 1)
	assert.Contains(t, messages[0].MsgRequest(), "Subject: Test Email")
	assert.Contains(t, messages[0].To(), "recipient@example.com")
}
