<div align="center">
  <img src="./assets/readme-banner-mailgrid.svg" alt="Mailgrid Logo" width="100%" height="100%"/>
</div>

<p align="center">
  <a href="https://pkg.go.dev/github.com/bravo1goingdark/mailgrid">
    <img src="https://pkg.go.dev/badge/github.com/bravo1goingdark/mailgrid.svg" alt="Go Reference"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/bravo1goingdark/mailgrid">
    <img src="https://goreportcard.com/badge/github.com/bravo1goingdark/mailgrid" alt="Go Report Card"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/actions/workflows/go.yml">
    <img src="https://github.com/bravo1goingdark/mailgrid/actions/workflows/go.yml/badge.svg" alt="CI"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/releases/latest">
    <img src="https://img.shields.io/github/v/release/bravo1goingdark/mailgrid" alt="Latest Release"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-BSD--3--Clause-blue.svg" alt="License"/>
  </a>
</p>

# Mailgrid

Production-ready CLI for bulk email campaigns via SMTP. Supports CSV/Google Sheets recipients, HTML templating, scheduling, real-time monitoring, recipient filtering, and resumable delivery.

## Install

```bash
go install github.com/bravo1goingdark/mailgrid/cmd/mailgrid@latest
```

Or download a binary from [Releases](https://github.com/bravo1goingdark/mailgrid/releases/latest).

**Build from source:**

```bash
git clone https://github.com/bravo1goingdark/mailgrid.git
cd mailgrid
make build
```

## Quick Start

```bash
# 1. Create config.json
cat > config.json <<'EOF'
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "you@gmail.com",
    "password": "your-app-password",
    "from": "You <you@gmail.com>"
  }
}
EOF

# 2. Send a single email
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Hi!"

# 3. Bulk send from CSV
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"
```

## Features

| Feature | |
|---|---|
| CSV & Google Sheets | Read recipients from CSV files or public Google Sheets |
| HTML Templates | Personalized emails with Go `text/template` syntax |
| Concurrency | Parallel SMTP workers for throughput |
| Scheduling | One-time, interval, or cron-based job scheduling (BoltDB-backed) |
| Monitoring | Real-time dashboard with SSE live updates |
| Filtering | Expression-based recipient filtering |
| Resumable | Interrupted campaigns pick up where they left off |
| Webhooks | HTTP POST notification on campaign completion |
| Dry Run & Preview | Test templates without sending |
| CC / BCC | Carbon copy and blind carbon copy support |
| Attachments | File attachments up to 10MB |
| Auto-Reconnect | SMTP reconnection on connection failure |
| Graceful Shutdown | Clean exit on SIGINT/SIGTERM |

## Usage

### SMTP Configuration

Create a JSON config file:

```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "you@gmail.com",
    "password": "your-app-password",
    "from": "You <you@gmail.com>"
  },
  "timeout_ms": 5000
}
```

> For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833).

### CSV Recipients

```csv
email,name,company
john@example.com,John,Acme Corp
jane@example.com,Jane,Tech Inc
```

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"
```

### Google Sheets

```bash
mailgrid --env config.json \
  --sheet-url "https://docs.google.com/spreadsheets/d/abc123/edit?gid=0" \
  --template email.html --subject "Newsletter"
```

### Templates

```html
<!-- email.html -->
<p>Hello {{ .name }},</p>
<p>Welcome to {{ .company }}!</p>
```

Template variables: `{{ .email }}`, `{{ .name }}`, `{{ .company }}` — any CSV column.

### Single Recipient

```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"
```

### Attachments

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --attach invoice.pdf --attach terms.pdf
```

### CC / BCC

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --cc "team@example.com" --bcc "archive@example.com"
```

### Dry Run

Render emails without sending:

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --dry-run
```

### Preview Server

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --preview
# Opens http://localhost:8080
```

### Monitoring Dashboard

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --monitor
# Opens http://localhost:9091
```

Shows live stats: progress, emails/sec, per-recipient status, domain breakdown, SMTP response codes.

### Filtering

Filter recipients with expressions:

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium" && location == "US"'
```

Supported: `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `startsWith`, `endsWith`, `and`, `or`, `not`. See [filter docs](./docs/filter.md).

### Scheduling

```bash
# One-time (RFC3339)
mailgrid --env config.json --csv recipients.csv --template email.html \
  --schedule-at "2025-06-15T10:00:00Z"

# Every 30 minutes
mailgrid --env config.json --csv recipients.csv --template email.html --interval "30m"

# Cron (daily at 9 AM)
mailgrid --env config.json --csv recipients.csv --template email.html --cron "0 9 * * *"

# Job management
mailgrid --jobs-list
mailgrid --jobs-cancel "job-id-123"
mailgrid --scheduler-run    # run as daemon
```

### Resumable Delivery

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --resume
mailgrid --env config.json --csv recipients.csv --template email.html --reset-offset
```

### Webhooks

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --webhook "https://your-server.com/webhook"
```

Payload:

```json
{
  "job_id": "mailgrid-123",
  "status": "completed",
  "total_recipients": 150,
  "successful_deliveries": 148,
  "failed_deliveries": 2,
  "duration_seconds": 330
}
```

### Performance Tuning

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --concurrency 10 --retries 3 --batch-size 5
```

| Provider | Concurrency |
|----------|-------------|
| Gmail | 1–2 |
| SendGrid / Mailgun | 5–10 |
| Amazon SES | 10–20 |

### TLS Configuration

```json
{
  "smtp": {
    "host": "smtp.company.com",
    "port": 587,
    "username": "user@company.com",
    "password": "password",
    "from": "Mailer <noreply@company.com>",
    "tls_cert_file": "/etc/mailgrid/ca.pem",
    "tls_key_file": "/etc/mailgrid/client.pem",
    "insecure_tls": false
  }
}
```

## CLI Reference

```
Core
  -e, --env string           SMTP config file
  -f, --csv string           CSV file path
  -u, --sheet-url string     Google Sheet URL
  -t, --template string      HTML template path
  -s, --subject string       Email subject (default "Test Email from Mailgrid")

Content
      --to string            Single recipient
      --text string          Plain text body or path to .txt file
      --cc string            CC recipients (comma-separated or file)
      --bcc string           BCC recipients (comma-separated or file)
  -a, --attach strings       Attachments (repeatable)

Testing
  -d, --dry-run              Render without sending
  -p, --preview              Start preview server
      --preview-port int     Preview port (default 8080)

Performance
  -c, --concurrency int      Workers (default 1)
  -r, --retries int          Email retries (default 1)
  -b, --batch-size int       Batch size (default 1)

Monitoring
  -m, --monitor              Enable dashboard
      --monitor-port int     Dashboard port (default 9091)

Filtering
  -F, --filter string        Recipient filter expression

Scheduling
  -A, --schedule-at string   Schedule time (RFC3339)
  -i, --interval string      Repeat interval
  -C, --cron string          Cron expression
  -J, --job-retries int      Scheduler retries (default 3)

Jobs
  -L, --jobs-list            List scheduled jobs
  -X, --jobs-cancel string   Cancel a job by ID
  -R, --scheduler-run        Run scheduler daemon

Advanced
  -w, --webhook string       Webhook URL
      --db-path string       BoltDB path (default "mailgrid.db")
      --resume               Resume from offset
      --reset-offset         Clear offset
```

## Project Structure

```
mailgrid/
├── cmd/mailgrid/        # Entry point
├── cli/                 # CLI parsing, orchestration
├── config/              # SMTP config loading
├── database/            # BoltDB job persistence
├── email/               # SMTP client, dispatcher, worker pool
├── internal/types/      # Shared types (Job, CLIArgs)
├── logger/              # Structured logging
├── monitor/             # HTTP dashboard with SSE
├── offset/              # Resumable delivery tracking
├── parser/              # CSV parsing, expression evaluation, filtering
├── scheduler/           # Cron/interval job scheduling
├── utils/               # Templates, helpers
├── webhook/             # Webhook notifications
├── test/                # Integration tests
└── docs/                # Documentation
```

## Documentation

- [Full CLI Reference](./docs/docs.md)
- [Filter Syntax](./docs/filter.md)

## License

BSD-3-Clause — see [LICENSE](./LICENSE).
