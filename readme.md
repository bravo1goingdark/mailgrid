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

**Mailgrid** is a high-performance, lightweight CLI tool written in Go for sending bulk emails via SMTP from CSV or Google Sheets. Built for speed, reliability, and minimalism — no bloated web UIs, just powerful automation.

---

## Table of Contents

- [Key Features](#key-features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
  - [Single Email](#single-email)
  - [Bulk Emails from CSV](#bulk-emails-from-csv)
  - [With Monitoring](#with-monitoring)
  - [Preview Before Sending](#preview-before-sending)
  - [Resumable Delivery](#resumable-delivery)
- [Scheduling](#scheduling)
  - [Schedule for Later](#schedule-for-later)
  - [Recurring Emails](#recurring-emails)
  - [Job Management](#job-management)
- [Advanced Features](#advanced-features)
  - [Filtering Recipients](#filtering-recipients)
  - [Attachments](#attachments)
  - [Webhooks](#webhooks)
- [CLI Flags Reference](#cli-flags-reference)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Bulk Email Sending** | Send emails to thousands of recipients from CSV or Google Sheets |
| **Dynamic Templating** | Personalize content using Go's `text/template` with CSV columns |
| **Real-time Monitoring** | Live dashboard showing delivery progress, success/failure rates |
| **Resumable Delivery** | Resume interrupted campaigns from where they left off |
| **Advanced Scheduling** | One-time, interval-based, or cron-based scheduling |
| **High Performance** | Concurrent workers, connection pooling, and batch processing |
| **Production Ready** | Retry logic, circuit breakers, and comprehensive error handling |

---

## Quick Start

```bash
# Send to a single recipient
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"

# Send bulk emails from CSV
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"

# Preview and monitor
mailgrid --env config.json --csv recipients.csv --template email.html --preview --monitor
```

---

## Installation

### Quick Install

```bash
# Linux & macOS
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex
```

### Download Binaries

Get pre-built binaries from [GitHub Releases](https://github.com/bravo1goingdark/mailgrid/releases/latest)

---

## Configuration

Create a `config.json` file with your SMTP settings:

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

> **Note:** For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833). For other providers, use your regular SMTP credentials.

---

## Usage Examples

### Single Email

```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome to our service!"
```

### Bulk Emails from CSV

Prepare a CSV file (`recipients.csv`):
```csv
email,name,company
john@example.com,John,Acme Corp
jane@example.com,Jane,Tech Inc
```

Send bulk emails:
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"
```

### With Monitoring

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --monitor --concurrency 5
```
Access the dashboard at `http://localhost:9091`

### Preview Before Sending

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --preview
```
Opens a local server to preview rendered emails in your browser.

### Resumable Delivery

```bash
# Start campaign (progress automatically tracked)
mailgrid --env config.json --csv recipients.csv --template email.html

# Resume if interrupted
mailgrid --env config.json --csv recipients.csv --template email.html --resume

# Start fresh (clear offset)
mailgrid --env config.json --csv recipients.csv --template email.html --reset-offset
```

---

## Scheduling

### Schedule for Later

```bash
# One-time scheduling (RFC3339 format)
mailgrid --env config.json --to user@example.com --subject "Reminder" \
  --text "Meeting at 3pm" --schedule-at "2025-01-01T10:00:00Z"
```

### Recurring Emails

```bash
# Every 30 minutes
mailgrid --env config.json --csv subscribers.csv --template newsletter.html --interval "30m"

# Daily at 9 AM (cron)
mailgrid --env config.json --csv recipients.csv --template report.html --cron "0 9 * * *"
```

### Job Management

```bash
mailgrid --jobs-list                    # List all scheduled jobs
mailgrid --jobs-cancel "job-id-123"     # Cancel a specific job
mailgrid --scheduler-run                 # Run scheduler as daemon
```

---

## Advanced Features

### Filtering Recipients

Filter recipients using expressions:

```bash
# Send only to premium users in New York
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium" && city == "New York"'
```

Supported operators: `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `startsWith`, `endsWith`

### Attachments

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --attach brochure.pdf --attach terms.pdf
```

### Webhooks

Receive notifications when campaigns complete:

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --webhook "https://your-server.com/webhook"
```

---

## CLI Flags Reference

| Short | Long Flag | Description |
|-------|-----------|-------------|
| `-e` | `--env` | Path to SMTP config JSON |
| `-f` | `--csv` | Path to recipient CSV file |
| `-u` | `--sheet-url` | Public Google Sheet URL |
| `-t` | `--template` | Path to HTML email template |
| `-s` | `--subject` | Email subject line |
| `-d` | `--dry-run` | Render emails without sending |
| `-p` | `--preview` | Start preview server |
| `-c` | `--concurrency` | Number of concurrent workers (default: 1) |
| `-r` | `--retries` | Max retries per email (default: 1) |
| `-b` | `--batch-size` | Emails per SMTP batch |
| `-m` | `--monitor` | Enable monitoring dashboard |
| `-a` | `--attach` | File attachments (repeatable) |
| `-w` | `--webhook` | Webhook URL for notifications |
| `-F` | `--filter` | Filter expression for recipients |
| `-A` | `--schedule-at` | Schedule time (RFC3339) |
| `-i` | `--interval` | Repeat interval (e.g., 30m, 1h) |
| `-C` | `--cron` | Cron expression for scheduling |
| `-L` | `--jobs-list` | List scheduled jobs |
| `-X` | `--jobs-cancel` | Cancel job by ID |
| `-R` | `--scheduler-run` | Run scheduler in foreground |
| `-D` | `--scheduler-db` | Path to scheduler database |
| | `--resume` | Resume from last offset |
| | `--reset-offset` | Clear offset and start fresh |

---

## Documentation

- [CLI Reference](./docs/docs.md) - Complete flag documentation
- [Filter Syntax](./docs/filter.md) - Advanced filtering options
- [Installation Guide](./INSTALLATION.md) - Detailed setup instructions

---

## Contributing

Contributions are welcome! Please read our [Contributing Guide](./CONTRIBUTING.md) for details.

---

## License

Licensed under BSD-3-Clause — see [LICENSE](./LICENSE)
