---
layout: default
title: CLI Reference
nav_order: 5
---

# CLI Reference
{: .no_toc }

Complete reference for all MailGrid command-line flags and options.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Basic Usage

```bash
mailgrid [flags]
```

### Quick Examples

```bash
# Single email
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"

# Bulk emails from CSV
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"

# Monitor campaign progress
mailgrid --env config.json --csv recipients.csv --template email.html --monitor --concurrency 5

# Schedule recurring emails
mailgrid --env config.json --to user@example.com --subject "Weekly Report" --text "Report..." --cron "0 9 * * 1"
```

---

## Configuration Flags

### `--env`
{: .d-inline-block }
Required
{: .label .label-red }

Path to SMTP configuration JSON file.

```bash
--env config.json
```

**Example config.json:**
```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com"
  },
  "rate_limit": 10,
  "timeout_ms": 5000
}
```

---

## Recipient Flags

### `--csv` | `-f`

Path to CSV file containing recipients.

```bash
--csv recipients.csv
```

**Required column:** `email` (case-insensitive)
**Optional columns:** Any additional fields for template substitution

**Example CSV:**
```csv
email,name,company
john@example.com,John Doe,Acme Corp
jane@example.com,Jane Smith,Tech Inc
```

### `--sheet-url` | `-u`

Public Google Sheets URL as CSV data source.

```bash
--sheet-url "https://docs.google.com/spreadsheets/d/1EUh5VWlSNtrlEIJ6SjJAQ9kYAcf4XrlsIIwXtYjImKc/edit?gid=1980978683#gid=1980978683"
```

{: .note }
> Only works with public Google Sheets (set to "Anyone with the link can view").

### `--to`

Single recipient email address (mutually exclusive with `--csv` and `--sheet-url`).

```bash
--to user@example.com
```

---

## Content Flags

### `--template` | `-t`

Path to HTML email template with Go template syntax.

```bash
--template email.html
```

**Default:** `example/welcome.html`

**Template syntax:**
```html
<h1>Hello {{.name}}!</h1>
<p>Welcome to {{.company}}.</p>
```

### `--subject` | `-s`

Email subject line (supports template syntax).

```bash
--subject "Welcome {{.name}}!"
```

**Default:** `Test Email from Mailgrid`

### `--text`

Plain-text email body (mutually exclusive with `--template`).

```bash
# Inline text
--text "This is a test email"

# From file
--text ./body.txt
```

### `--cc`

CC recipients (visible to all recipients).

```bash
# Comma-separated list
--cc "team@example.com,manager@example.com"

# From file
--cc @cc_list.txt
```

### `--bcc`

BCC recipients (hidden from other recipients).

```bash
# Comma-separated list
--bcc "admin@example.com"

# From file
--bcc @bcc_list.txt
```

### `--attach` | `-a`

File attachments (repeat flag for multiple files, max 10MB total).

```bash
--attach brochure.pdf --attach terms.pdf
```

---

## Processing Flags

### `--concurrency` | `-c`

Number of parallel SMTP workers.

```bash
--concurrency 5
```

**Default:** `1`
**Recommended:** ≤ 5 for consumer email providers

### `--retries` | `-r`

Retry attempts per failed email (exponential backoff).

```bash
--retries 3
```

**Default:** `1`

### `--batch-size` | `-b`

Number of emails per SMTP batch.

```bash
--batch-size 5
```

**Default:** `1`
**Note:** Use `1` for Gmail/Yahoo/Outlook to avoid rate limiting

### `--filter` | `-F`

Logical filter for recipients.

```bash
--filter 'tier = "pro" and age > 25'
```

**Supported operators:** `=`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `and`, `or`

---

## Monitoring Flags

### `--monitor` | `-m`

Enable real-time monitoring dashboard.

```bash
--monitor
```

**Dashboard URL:** `http://localhost:9091` (default)

### `--monitor-port`

Custom port for monitoring dashboard.

```bash
--monitor-port 8080
```

**Default:** `9091`

---

## Preview & Testing Flags

### `--preview` | `-p`

Start preview server to view rendered emails.

```bash
--preview
```

**Server URL:** `http://localhost:8080` (default)

### `--port`

Custom port for preview server.

```bash
--port 7070
```

**Default:** `8080`

### `--dry-run` | `-d`

Render emails without sending (testing mode).

```bash
--dry-run
```

---

## Scheduling Flags

### `--schedule-at` | `-A`

Schedule one-time send at specific RFC3339 timestamp.

```bash
--schedule-at "2025-09-08T09:00:00Z"
```

### `--interval` | `-i`

Recurring schedule using Go duration.

```bash
--interval 30m    # Every 30 minutes
--interval 1h     # Every hour
--interval 24h    # Every 24 hours
```

### `--cron` | `-C`

Recurring schedule using 5-field cron expression.

```bash
--cron "0 9 * * *"      # Daily at 9:00 AM
--cron "0 9 * * 1"      # Weekly on Monday at 9:00 AM
--cron "0 9 1 * *"      # Monthly on 1st at 9:00 AM
```

**Format:** `minute hour day month weekday`

### `--job-retries` | `-J`

Scheduler-level retry attempts on handler failure.

```bash
--job-retries 3
```

**Default:** `3`

### `--job-backoff` | `-B`

Base backoff duration for scheduler retries.

```bash
--job-backoff 2s
```

**Default:** `2s`

### `--scheduler-db` | `-D`

Path to BoltDB file for job persistence.

```bash
--scheduler-db custom-schedules.db
```

**Default:** `mailgrid.db`

---

## Job Management Flags

### `--jobs-list` | `-L`

List all scheduled jobs.

```bash
--jobs-list
```

### `--jobs-cancel` | `-X`

Cancel job by ID.

```bash
--jobs-cancel "job-id-123"
```

### `--scheduler-run` | `-R`

Run scheduler dispatcher in foreground.

```bash
--scheduler-run
```

Press `Ctrl+C` to stop.

---

## Integration Flags

### `--webhook` | `-w`

HTTP URL for POST notifications on campaign completion.

```bash
--webhook "https://api.example.com/webhooks/mailgrid"
```

**Payload includes:**
- Job ID and status
- Total/successful/failed delivery counts
- Start/end times and duration
- File paths and configuration

---

## Flag Combinations

### High-Performance Bulk Campaign

```bash
mailgrid --env config.json \
         --csv large_list.csv \
         --template campaign.html \
         --subject "Special Offer" \
         --concurrency 10 \
         --batch-size 5 \
         --retries 3 \
         --monitor \
         --webhook "https://api.example.com/webhooks"
```

### Scheduled Newsletter

```bash
mailgrid --env config.json \
         --csv subscribers.csv \
         --template newsletter.html \
         --subject "Weekly Newsletter" \
         --cron "0 9 * * 1" \
         --monitor
```

### Safe Testing

```bash
mailgrid --env config.json \
         --csv test_recipients.csv \
         --template email.html \
         --subject "Test Email" \
         --dry-run \
         --preview
```

---

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | Configuration error |
| `4` | File not found |
| `5` | SMTP connection error |

---

## Environment Variables

MailGrid respects these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `MAILGRID_CONFIG` | Default config file path | `config.json` |
| `MAILGRID_DB` | Default scheduler database | `mailgrid.db` |
| `MAILGRID_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

---

## Short Flag Reference

| Short | Long | Description |
|-------|------|-------------|
| `-e` | `--env` | SMTP config file |
| `-f` | `--csv` | CSV recipient file |
| `-u` | `--sheet-url` | Google Sheets URL |
| `-t` | `--template` | HTML template file |
| `-s` | `--subject` | Email subject |
| `-d` | `--dry-run` | Test mode (no sending) |
| `-p` | `--preview` | Preview server |
| `-c` | `--concurrency` | Parallel workers |
| `-r` | `--retries` | Retry attempts |
| `-b` | `--batch-size` | Emails per batch |
| `-F` | `--filter` | Recipient filter |
| `-a` | `--attach` | File attachments |
| `-w` | `--webhook` | Notification URL |
| `-m` | `--monitor` | Monitoring dashboard |
| `-A` | `--schedule-at` | One-time schedule |
| `-i` | `--interval` | Recurring interval |
| `-C` | `--cron` | Cron schedule |
| `-J` | `--job-retries` | Scheduler retries |
| `-B` | `--job-backoff` | Scheduler backoff |
| `-L` | `--jobs-list` | List jobs |
| `-X` | `--jobs-cancel` | Cancel job |
| `-R` | `--scheduler-run` | Run scheduler |
| `-D` | `--scheduler-db` | Scheduler database |