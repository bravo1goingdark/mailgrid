package cli

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/config"
	"github.com/bravo1goingdark/mailgrid/email"
	"github.com/bravo1goingdark/mailgrid/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareEmailTasks(t *testing.T) {
	tests := []struct {
		name         string
		recipients   []parser.Recipient
		templatePath string
		subjectTpl   string
		attachments  []string
		ccList       []string
		bccList      []string
		wantLen      int
		wantError    bool
	}{
		{
			name: "valid recipients with template",
			recipients: []parser.Recipient{
				{
					Email: "user1@example.com",
					Data:  map[string]string{"name": "John", "email": "user1@example.com"},
				},
				{
					Email: "user2@example.com",
					Data:  map[string]string{"name": "Jane", "email": "user2@example.com"},
				},
			},
			templatePath: "", // Will be empty since we can't create actual template files in test
			subjectTpl:   "Hello {{.name}}",
			attachments:  []string{},
			ccList:       []string{},
			bccList:      []string{},
			wantLen:      2,
			wantError:    false,
		},
		{
			name: "recipients with missing fields",
			recipients: []parser.Recipient{
				{
					Email: "user1@example.com",
					Data:  map[string]string{"name": "John", "email": "user1@example.com"},
				},
				{
					Email: "user2@example.com",
					Data:  map[string]string{"name": "", "email": "user2@example.com"}, // missing name
				},
			},
			templatePath: "",
			subjectTpl:   "Hello {{.name}}",
			attachments:  []string{},
			ccList:       []string{},
			bccList:      []string{},
			wantLen:      1, // One recipient should be skipped
			wantError:    false,
		},
		{
			name: "invalid subject template",
			recipients: []parser.Recipient{
				{
					Email: "user1@example.com",
					Data:  map[string]string{"name": "John", "email": "user1@example.com"},
				},
			},
			templatePath: "",
			subjectTpl:   "Hello {{.name", // Invalid template syntax
			attachments:  []string{},
			ccList:       []string{},
			bccList:      []string{},
			wantLen:      0,
			wantError:    true,
		},
		{
			name: "with attachments and CC/BCC",
			recipients: []parser.Recipient{
				{
					Email: "user1@example.com",
					Data:  map[string]string{"name": "John", "email": "user1@example.com"},
				},
			},
			templatePath: "",
			subjectTpl:   "Hello {{.name}}",
			attachments:  []string{"file1.pdf", "file2.jpg"},
			ccList:       []string{"cc@example.com"},
			bccList:      []string{"bcc@example.com"},
			wantLen:      1,
			wantError:    false,
		},
		{
			name:         "empty recipients",
			recipients:   []parser.Recipient{},
			templatePath: "",
			subjectTpl:   "Hello World",
			attachments:  []string{},
			ccList:       []string{},
			bccList:      []string{},
			wantLen:      0,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := PrepareEmailTasks(
				tt.recipients,
				tt.templatePath,
				tt.subjectTpl,
				tt.attachments,
				tt.ccList,
				tt.bccList,
			)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, tasks, tt.wantLen)

			// Verify task properties for non-empty results
			if tt.wantLen > 0 {
				task := tasks[0]
				assert.NotEmpty(t, task.Subject)
				assert.Equal(t, tt.attachments, task.Attachments)
				assert.Equal(t, tt.ccList, task.CC)
				assert.Equal(t, tt.bccList, task.BCC)
				assert.Equal(t, 0, task.Retries)
			}
		})
	}
}

func TestHasMissingFields(t *testing.T) {
	tests := []struct {
		name      string
		recipient parser.Recipient
		want      bool
	}{
		{
			name: "all fields present",
			recipient: parser.Recipient{
				Email: "user@example.com",
				Data:  map[string]string{"name": "John", "email": "user@example.com"},
			},
			want: false,
		},
		{
			name: "missing field",
			recipient: parser.Recipient{
				Email: "user@example.com",
				Data:  map[string]string{"name": "", "email": "user@example.com"},
			},
			want: true,
		},
		{
			name: "all fields empty",
			recipient: parser.Recipient{
				Email: "user@example.com",
				Data:  map[string]string{"name": "", "email": ""},
			},
			want: true,
		},
		{
			name: "empty data map",
			recipient: parser.Recipient{
				Email: "user@example.com",
				Data:  map[string]string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasMissingFields(tt.recipient)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSendSingleEmail_Validation(t *testing.T) {
	tests := []struct {
		name      string
		args      CLIArgs
		wantError bool
		errorMsg  string
	}{
		{
			name: "missing --to flag",
			args: CLIArgs{
				TemplatePath: "template.html",
			},
			wantError: true,
			errorMsg:  "--to flag is required for single email sending",
		},
		{
			name: "both template and text provided",
			args: CLIArgs{
				To:           "user@example.com",
				TemplatePath: "template.html",
				Text:         "Hello world",
			},
			wantError: true,
			errorMsg:  "either --template or --text must be provided, but not both",
		},
		{
			name: "neither template nor text provided",
			args: CLIArgs{
				To: "user@example.com",
			},
			wantError: true,
			errorMsg:  "either --template or --text must be provided, but not both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use empty SMTP config since we're just testing validation
			smtpConfig := config.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
			}
			err := SendSingleEmail(tt.args, smtpConfig)

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				// Note: This will likely error due to missing template/text file
				// but that's expected since we're just testing the validation logic
				if err != nil {
					t.Logf("Expected error in test setup: %v", err)
				}
			}
		})
	}
}

func TestPrintDryRun(t *testing.T) {
	// This test ensures printDryRun doesn't panic and handles empty/non-empty task lists
	tests := []struct {
		name  string
		tasks []email.Task
	}{
		{
			name:  "empty tasks",
			tasks: []email.Task{},
		},
		{
			name: "single task",
			tasks: []email.Task{
				{
					Recipient: parser.Recipient{
						Email: "user@example.com",
						Data:  map[string]string{"name": "John"},
					},
					Subject:     "Test Subject",
					Body:        "Test Body",
					Attachments: []string{"file.pdf"},
				},
			},
		},
		{
			name: "multiple tasks",
			tasks: []email.Task{
				{
					Recipient: parser.Recipient{
						Email: "user1@example.com",
						Data:  map[string]string{"name": "John"},
					},
					Subject: "Test Subject 1",
					Body:    "Test Body 1",
				},
				{
					Recipient: parser.Recipient{
						Email: "user2@example.com",
						Data:  map[string]string{"name": "Jane"},
					},
					Subject: "Test Subject 2",
					Body:    "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				printDryRun(tt.tasks)
			})
		})
	}
}