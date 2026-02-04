<div align="center">
   <img src="./assets/readme-banner-mailgrid.svg" alt="Mailgrid Logo" width="100%" height="100%"/>
</p> 
 
<p align="center">
   <a href="https://goreportcard.com/report/github.com/bravo1goingdark/mailgrid">
     <img src="https://img.shields.io/badge/Go%20Report-A-yellow?style=for-the-badge&logo=go" alt="Go Report Card"/>
  </a>
   <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/INSTALLATION.md">
     <img src="https://img.shields.io/badge/Installation-green?style=for-the-badge" alt="Installation Guide"/>
   </a>
   <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/CONTRIBUTING.md">
     <img src="https://img.shields.io/badge/Contributing-purple?style=for-the-badge" alt="Contributing Guide"/>
   </a>
   <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/INSTALLATION.md">
     <img src="https://img.shields.io/badge/Installation-green?style=for-the-badge" alt="Installation Guide"/>
   </a>
   <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/CONTRIBUTING.md">
     <img src="https://img.shields.io/badge/Contributing-purple?style=for-the-badge" alt="Contributing Guide"/>
   </a>
</p> 
 
**Mailgrid** is a high-performance, lightweight CLI tool written in Go for sending bulk emails via SMTP from CSV or Google Sheets. Built for speed, reliability, and minimalism — no bloated web UIs, just powerful automation.
 
## Key Features 
 
- **Bulk email sending** from CSV files or Google Sheets
- **Dynamic templating** with Go's `text/template` for personalized content
- **Real-time monitoring** dashboard with live delivery tracking
- **Resumable delivery** with offset tracking for interrupted campaigns
- **Advanced scheduling** with cron support and auto-start scheduler
- **High performance** with connection pooling, batching, and concurrent processing
- **Production ready** with zero dependencies, cross-platform support
 
## Installation
 
**Quick install:**
```bash
# Linux & macOS
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash
 
# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex
 
```
 
Download binaries: [GitHub Releases](https://github.com/bravo1goingdark/mailgrid/releases/latest) • [Installation Guide](./INSTALLATION.md)
 
**Setup:**
1. Create `config.json`:
    ```json
    {
      "smtp": {
        "host": "smtp.gmail.com",
        "port": 587,
        "username": "your-email@gmail.com",
        "password": "your-app-password",
        "from": "your-email@gmail.com"
      }
    }
    ```
 
2. Test installation:
    ```bash
    mailgrid --env config.json --to test@example.com --subject "Test" --text "Hello!" --dry-run
    ```
 
## Quick Start
 
**Single email:**
```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"
```
 
**Bulk emails from CSV:**
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"
```
 
**With monitoring:**
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --monitor --concurrency 5
```
 
**Preview before sending:**
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --preview
```
 
**Resumable delivery:**
```bash
# Start campaign (automatically tracked)
mailgrid --env config.json --csv recipients.csv --template email.html
 
# Resume if interrupted
mailgrid --env config.json --csv recipients.csv --template email.html --resume
 
# Start fresh
mailgrid --env config.json --csv recipients.csv --template email.html --reset-offset
```
 
## Scheduling & Automation
 
**Schedule emails for later:**
```bash
# One-time scheduling
mailgrid --env config.json --to user@example.com --subject "Reminder" --text "Meeting at 3pm" --schedule-at "2025-01-01T10:00:00Z"
 
# Recurring (every 30 minutes)
mailgrid --env config.json --csv subscribers.csv --template newsletter.html --interval "30m"
 
# Cron-based (daily at 9 AM)
mailgrid --env config.json --csv recipients.csv --template report.html --cron "0 9 * * *"
```
 
**Manage scheduled jobs:**
```bash
mailgrid --jobs-list                    # List all jobs
mailgrid --jobs-cancel "job-id-123"     # Cancel specific job
mailgrid --scheduler-run                # Run as daemon
```
 
## Common Flags
 
| Short | Long Flag | Description |
|-------|-----------|-------------|
| `-e` | `--env` | Path to SMTP config JSON |
| `-f` | `--csv` | Path to recipient CSV file |
| `-u` | `--sheet-url` | Public Google Sheet URL |
| `-t` | `--template` | Path to email HTML template |
| `-s` | `--subject` | Email subject line |
| `-d` | `--dry-run` | Render emails without sending |
| `-p` | `--preview` | Start preview server |
| `-c` | `--concurrency` | Number of concurrent workers |
| `-m` | `--monitor` | Enable monitoring dashboard |
| `-a` | `--attach` | File attachments |
| `-w` | `--webhook` | HTTP URL for notifications |
 
## Documentation
 
- [Complete Documentation](./docs/docs.md) - Full CLI reference and examples
- [Installation Guide](./INSTALLATION.md) - Detailed setup instructions
- [Contributing Guide](./CONTRIBUTING.md) - How to contribute to the project
 
---
 
> Licensed under BSD-3-Clause — see [LICENSE](./LICENSE)
