<div align="center">
  <img src="./assets/readme-banner-mailgrid.svg" alt="Mailgrid Logo" width="100%" height="100%"/>
</div>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/bravo1goingdark/mailgrid">
    <img src="https://img.shields.io/badge/Go%20Report-A-yellow?style=for-the-badge&logo=go" alt="Go Report Card"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/actions/workflows/go.yml">
    <img src="https://img.shields.io/badge/Tests-passing-brightgreen?style=for-the-badge" alt="CI Status"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/releases/latest">
    <img src="https://img.shields.io/badge/Release-v0.1.0-blue?style=for-the-badge" alt="Latest Release"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-BSD--3--Clause-green?style=for-the-badge" alt="License"/>
  </a>
</p>

**Mailgrid** is a high-performance CLI tool written in Go for sending bulk emails via SMTP. Send to CSV, Google Sheets, or single recipients — with templating, scheduling, monitoring, and more.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Examples](#examples)
  - [Single Email](#single-email)
  - [Bulk from CSV](#bulk-from-csv)
  - [From Google Sheets](#from-google-sheets)
  - [With CC/BCC](#with-ccbcc)
  - [Preview Mode](#preview-mode)
  - [Dry Run](#dry-run)
  - [Monitoring Dashboard](#monitoring-dashboard)
- [Performance](#performance)
  - [Concurrency](#concurrency)
  - [Retries](#retries)
  - [Batch Size](#batch-size)
- [Scheduling](#scheduling)
  - [One-time](#one-time-scheduling)
  - [Recurring](#recurring-scheduling)
  - [Job Management](#job-management)
- [Advanced Features](#advanced-features)
  - [Filtering](#filtering)
  - [Attachments](#attachments)
  - [Webhooks](#webhooks)
  - [Resumable Delivery](#resumable-delivery)
- [CLI Reference](#cli-reference)
- [Documentation](#documentation)

---

## Features

| Feature | Description |
|---------|-------------|
| **Single Email** | Send to one recipient without CSV |
| **Bulk CSV** | Send to thousands from CSV file |
| **Google Sheets** | Fetch recipients from public Google Sheet |
| **HTML Templates** | Personalized emails with Go templates |
| **CC/BCC** | Add carbon copy recipients |
| **Attachments** | Attach files to emails (max 10MB) |
| **Preview** | Preview rendered emails in browser |
| **Dry Run** | Test without sending |
| **Monitoring** | Real-time dashboard with live stats |
| **Scheduling** | One-time or recurring (cron/interval) |
| **Filtering** | Filter recipients with expressions |
| **Resumable** | Resume interrupted campaigns |
| **Webhooks** | Get notified when campaigns complete |
| **Concurrent** | Parallel SMTP workers |
| **Retry** | Automatic retry with backoff |
| **Circuit Breaker** | Built-in resilience |

---

## Quick Start

```bash
# Single email
mailgrid -e config.json -to user@example.com -s "Hello" -text "Hi!"

# Bulk from CSV
mailgrid -e config.json -f recipients.csv -t template.html -s "Hi {{.name}}!"

# With monitoring
mailgrid -e config.json -f recipients.csv -t template.html -m -c 5
```

---

## Installation

```bash
# Linux/macOS
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex

# Or download from GitHub Releases
https://github.com/bravo1goingdark/mailgrid/releases/latest
```

---

## Configuration

Create `config.json`:

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

> **Gmail:** Use an [App Password](https://support.google.com/accounts/answer/185833)

---

## Examples

### Single Email

```bash
mailgrid -e config.json -to user@example.com -s "Hello" -text "Welcome!"
```

### Bulk from CSV

```csv
# recipients.csv
email,name,company
john@example.com,John,Acme
jane@example.com,Jane,Tech Inc
```

```bash
mailgrid -e config.json -f recipients.csv -t template.html -s "Hi {{.name}}!"
```

### From Google Sheets

```bash
mailgrid -e config.json -u "https://docs.google.com/spreadsheets/d/abc123/edit?gid=0" \
  -t template.html -s "Newsletter"
```

### With CC/BCC

```bash
# CC recipients (visible to all)
mailgrid -e config.json -f recipients.csv -t template.html --cc "manager@example.com"

# BCC recipients (hidden)
mailgrid -e config.json -f recipients.csv -t template.html --bcc "archive@example.com"

# Both
mailgrid -e config.json -f recipients.csv -t template.html \
  --cc "team@example.com" --bcc "boss@example.com"
```

### Preview Mode

```bash
# Preview at default port 8080
mailgrid -e config.json -f recipients.csv -t template.html -p

# Custom port
mailgrid -e config.json -f recipients.csv -t template.html -p --port 9000
```

Opens browser at `http://localhost:8080` to preview rendered emails.

### Dry Run

```bash
# Render emails without sending
mailgrid -e config.json -f recipients.csv -t template.html -d
```

### Monitoring Dashboard

```bash
# Enable with default port 9091
mailgrid -e config.json -f recipients.csv -t template.html -m

# Custom port
mailgrid -e config.json -f recipients.csv -t template.html -m --monitor-port 8080
```

Access at `http://localhost:9091` — shows live stats, progress, and logs.

---

## Performance

### Concurrency

```bash
# Multiple parallel workers
mailgrid -e config.json -f recipients.csv -t template.html -c 5

# Guidelines:
#   Gmail: 1-2 workers
#   SendGrid/Mailgun: 5-10
#   Amazon SES: 10-20
```

### Retries

```bash
# Retry failed emails (default: 1)
mailgrid -e config.json -f recipients.csv -t template.html -r 3
```

Uses exponential backoff with jitter.

### Batch Size

```bash
# Emails per SMTP batch (default: 1)
mailgrid -e config.json -f recipients.csv -t template.html -b 10
```

> Consumer inboxes (Gmail/Yahoo): use 1  
> Corporate/warmed IPs: 5-10

---

## Scheduling

### One-time Scheduling

```bash
# At specific time (RFC3339)
mailgrid -e config.json -f recipients.csv -t template.html \
  -A "2025-01-15T10:00:00Z"
```

### Recurring Scheduling

```bash
# Every 30 minutes
mailgrid -e config.json -f recipients.csv -t template.html -i "30m"

# Daily at 9 AM (cron)
mailgrid -e config.json -f recipients.csv -t template.html -C "0 9 * * *"

# Weekly on Monday
mailgrid -e config.json -f recipients.csv -t template.html -C "0 9 * * 1"
```

### Job Management

```bash
mailgrid -L                    # List scheduled jobs
mailgrid -X "job-id-123"      # Cancel a job
mailgrid -R                   # Run scheduler as daemon
```

---

## Advanced Features

### Filtering

Filter recipients with expressions:

```bash
# Premium users only
--filter 'tier == "premium"'

# Complex filter
--filter '(tier == "vip" or tier == "premium") && location == "US"'

# Email domain
--filter 'email contains "@company.com"'
```

Operators: `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `startsWith`, `endsWith`, `and`, `or`, `not`

### Attachments

```bash
mailgrid -e config.json -f recipients.csv -t template.html \
  -a invoice.pdf -a terms.pdf
```

Max 10MB total.

### Webhooks

Get notified when campaign completes:

```bash
mailgrid -e config.json -f recipients.csv -t template.html \
  -w "https://your-server.com/webhook"
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

### Resumable Delivery

```bash
# Start campaign (tracked automatically)
mailgrid -e config.json -f recipients.csv -t template.html

# Resume if interrupted
mailgrid -e config.json -f recipients.csv -t template.html --resume

# Start fresh
mailgrid -e config.json -f recipients.csv -t template.html --reset-offset
```

---

## CLI Reference

| Short | Flag | Description |
|-------|------|-------------|
| **Core** |||
| `-e` | `--env` | SMTP config file |
| `-f` | `--csv` | CSV file |
| `-u` | `--sheet-url` | Google Sheet URL |
| `-t` | `--template` | HTML template |
| `-s` | `--subject` | Email subject |
| **Content** |||
| `-to` | `--to` | Single recipient |
| `-text` | `--text` | Plain text body |
| `-cc` | `--cc` | CC recipients |
| `-bcc` | `--bcc` | BCC recipients |
| `-a` | `--attach` | Attachments |
| **Testing** |||
| `-d` | `--dry-run` | Preview without sending |
| `-p` | `--preview` | Browser preview server |
| | `--port` | Preview server port (default: 8080) |
| **Performance** |||
| `-c` | `--concurrency` | Workers (default: 1) |
| `-r` | `--retries` | Email retries (default: 1) |
| `-b` | `--batch-size` | Batch size (default: 1) |
| **Monitoring** |||
| `-m` | `--monitor` | Enable dashboard |
| | `--monitor-port` | Dashboard port (default: 9091) |
| **Scheduling** |||
| `-A` | `--schedule-at` | One-time (RFC3339) |
| `-i` | `--interval` | Repeat interval |
| `-C` | `--cron` | Cron expression |
| `-J` | `--job-retries` | Scheduler retries (default: 3) |
| **Jobs** |||
| `-L` | `--jobs-list` | List jobs |
| `-X` | `--jobs-cancel` | Cancel job |
| `-R` | `--scheduler-run` | Run daemon |
| **Advanced** |||
| `-F` | `--filter` | Recipient filter |
| `-w` | `--webhook` | Webhook URL |
| | `--resume` | Resume from offset |
| | `--reset-offset` | Clear offset |
| | `--version` | Show version |

---

## Documentation

- [Full CLI Docs](./docs/docs.md)
- [Filter Syntax](./docs/filter.md)
- [Installation](./INSTALLATION.md)
- [Contributing](./CONTRIBUTING.md)

---

## License

BSD-3-Clause — see [LICENSE](./LICENSE)
