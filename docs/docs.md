# Mailgrid Documentation

Complete CLI reference for Mailgrid email automation tool.

## Table of Contents

- [Configuration](#configuration)
- [Data Sources](#data-sources)
- [Email Content](#email-content)
- [Performance Options](#performance-options)
- [Scheduling](#scheduling)
- [Monitoring](#monitoring)
- [Advanced Features](#advanced-features)
- [Examples](#examples)

---

## Configuration

### `--env` / `-e`

Path to SMTP configuration file (required for sending).

```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "Your Name <your-email@gmail.com>"
  }
}
```

---

## Data Sources

### `--csv` / `-f`

Path to CSV file with recipients. Required column: `email`.

```csv
email,name,company
john@example.com,John,Acme Corp
jane@example.com,Jane,Tech Inc
```

```bash
mailgrid --env config.json --csv recipients.csv --template email.html
```

### `--sheet-url` / `-u`

Fetch recipients from a public Google Sheet.

```bash
mailgrid --env config.json --sheet-url "https://docs.google.com/spreadsheets/d/abc123/edit?gid=0#gid=0" \
  --template email.html
```

> **Note:** Only works with public sheets (Anyone with link can view).

---

## Email Content

### `--template` / `-t`

HTML email template using Go templates.

```html
<p>Hello {{ .name }},</p>
<p>Welcome to {{ .company }}!</p>
```

### `--subject` / `-s`

Subject line with template support.

```bash
--subject "Welcome {{ .name }}!"
```

### `--text`

Plain text body (inline or file path).

```bash
# Inline
--text "Hello world!"

# From file
--text ./body.txt
```

### `--to`

Single recipient (mutually exclusive with --csv/--sheet-url).

```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Hi!"
```

### `--cc` / `--bcc`

CC/BCC recipients (comma-separated or `@file.txt`).

```bash
--cc "team@example.com,manager@example.com"
--bcc ./bcc_list.txt
```

### `--attach` / `-a`

File attachments (repeat for multiple files, max 10MB total).

```bash
--attach invoice.pdf --attach terms.pdf
```

---

## Performance Options

### `--concurrency` / `-c`

Number of parallel workers (default: 1).

```bash
--concurrency 5    # 5 parallel SMTP connections
```

**Guidelines:**
- Gmail: 1-2 workers
- SendGrid/Mailgun: 5-10 workers
- Amazon SES: 10-20 workers

### `--retries` / `-r`

Retry attempts per failed email (default: 1).

```bash
--retries 3
```

Uses exponential backoff with jitter.

### `--batch-size` / `-b`

Emails per SMTP batch (default: 1).

```bash
--batch-size 10
```

**Best practices:**
- Consumer inboxes (Gmail, Yahoo): use 1
- Corporate/warmed IPs: 5-10
- Test with `--dry-run` first

### `--dry-run` / `-d`

Render emails without sending.

```bash
--dry-run
```

### `--preview` / `-p`

Start preview server to view rendered emails in browser.

```bash
--preview              # Default port 8080
--preview --port 9000 # Custom port
```

---

## Scheduling

Schedule emails for later or recurring delivery. Jobs persist in BoltDB.

### One-time Scheduling

```bash
# Schedule at specific time (RFC3339)
--schedule-at "2025-01-15T10:00:00Z"
-A "2025-01-15T10:00:00Z"
```

### Recurring Scheduling

```bash
# Every 30 minutes
--interval "30m"
-i "30m"

# Cron expression (daily at 9 AM)
--cron "0 9 * * *"
-C "0 9 * * *"
```

### Job Management

```bash
--jobs-list              # List all jobs
-L

--jobs-cancel "job-id"   # Cancel specific job
-X "job-id"

--scheduler-run          # Run scheduler daemon
-R
```

### Scheduler Options

```bash
--job-retries 3                 # Scheduler-level retries (default: 3)
-J 3
```

---

## Monitoring

### `--monitor` / `-m`

Enable real-time monitoring dashboard.

```bash
--monitor              # Default port 9091
-m

--monitor-port 8080    # Custom port
```

Access at `http://localhost:9091`

### Metrics Endpoint

When scheduler is running:

```bash
curl http://localhost:8090/metrics   # Performance metrics
curl http://localhost:8090/health    # Health check
```

---

## Advanced Features

### `--filter` / `-F`

Filter recipients using expressions. See [Filter Documentation](filter.md) for full syntax.

```bash
--filter 'tier == "premium" and age > 25'
-F 'company contains "Tech"'
```

### `--webhook` / `-w`

Send HTTP POST notification when campaign completes.

```bash
--webhook "https://your-server.com/webhook"
-w "https://your-server.com/webhook"
```

**Payload:**
```json
{
  "job_id": "mailgrid-1633024800",
  "status": "completed",
  "total_recipients": 150,
  "successful_deliveries": 148,
  "failed_deliveries": 2,
  "duration_seconds": 330,
  "csv_file": "recipients.csv"
}
```

### Resumable Delivery

```bash
--resume           # Resume from last offset
--reset-offset    # Clear and start fresh
```

---

## Examples

### Single Email

```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Hi!"
```

### Bulk Email

```bash
mailgrid --env config.json \
  --csv recipients.csv \
  --template email.html \
  --subject "Hi {{.name}}!" \
  --concurrency 5 \
  --retries 3
```

### With Monitoring

```bash
mailgrid --env config.json \
  --csv recipients.csv \
  --template email.html \
  --monitor \
  --concurrency 10
```

### Scheduled Newsletter

```bash
mailgrid --env config.json \
  --csv subscribers.csv \
  --template newsletter.html \
  --cron "0 9 * * 1" \
  --concurrency 5
```

### Filtered Campaign

```bash
mailgrid --env config.json \
  --csv recipients.csv \
  --template email.html \
  --filter 'tier == "premium" && location != "EU"'
```

---

## Quick Reference

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--env` | `-e` | - | SMTP config file |
| `--csv` | `-f` | - | CSV file path |
| `--sheet-url` | `-u` | - | Google Sheet URL |
| `--template` | `-t` | - | HTML template |
| `--subject` | `-s` | "Test Email" | Subject line |
| `--to` | - | - | Single recipient |
| `--text` | - | - | Plain text body |
| `--cc` | - | - | CC recipients |
| `--bcc` | - | - | BCC recipients |
| `--attach` | `-a` | [] | Attachments |
| `--dry-run` | `-d` | false | Preview only |
| `--preview` | `-p` | false | Preview server |
| `--concurrency` | `-c` | 1 | Workers |
| `--retries` | `-r` | 1 | Email retries |
| `--batch-size` | `-b` | 1 | Batch size |
| `--filter` | `-F` | - | Recipient filter |
| `--webhook` | `-w` | - | Webhook URL |
| `--monitor` | `-m` | false | Dashboard |
| `--schedule-at` | `-A` | - | One-time schedule |
| `--interval` | `-i` | - | Repeat interval |
| `--cron` | `-C` | - | Cron schedule |
| `--jobs-list` | `-L` | false | List jobs |
| `--jobs-cancel` | `-X` | - | Cancel job |
| `--scheduler-run` | `-R` | false | Run daemon |
