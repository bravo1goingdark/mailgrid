---
layout: default
title: Email Scheduling
nav_order: 6
---

# Email Scheduling
{: .no_toc }

Automate email campaigns with MailGrid's powerful scheduling system.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

MailGrid includes a production-ready scheduler that supports:
- **One-time scheduling** at specific timestamps
- **Recurring campaigns** using intervals or cron expressions
- **Auto-start capabilities** with intelligent shutdown
- **Job persistence** with BoltDB storage
- **Distributed locking** for safe concurrent access

---

## One-Time Scheduling

Schedule a campaign to run at a specific time using RFC3339 timestamps.

### Basic Example

```bash
mailgrid --env config.json \
         --csv subscribers.csv \
         --template newsletter.html \
         --subject "Weekly Newsletter" \
         --schedule-at "2025-01-15T09:00:00Z"
```

### Time Zone Support

```bash
# UTC time
--schedule-at "2025-01-15T09:00:00Z"

# With timezone offset
--schedule-at "2025-01-15T09:00:00-05:00"  # EST
--schedule-at "2025-01-15T09:00:00+01:00"  # CET
```

### Single Email Scheduling

```bash
mailgrid --env config.json \
         --to "manager@company.com" \
         --subject "Monthly Report" \
         --text "Please find the attached monthly report." \
         --attach report.pdf \
         --schedule-at "2025-01-01T10:00:00Z"
```

---

## Recurring Campaigns

### Interval-Based Scheduling

Use Go duration syntax for simple recurring campaigns:

```bash
# Every 30 minutes
mailgrid --env config.json \
         --csv recipients.csv \
         --template reminder.html \
         --subject "Reminder: {{.task}}" \
         --interval "30m"

# Every 2 hours
mailgrid --env config.json \
         --to "admin@company.com" \
         --subject "System Status" \
         --text "All systems operational" \
         --interval "2h"

# Daily campaigns
mailgrid --env config.json \
         --csv daily_subscribers.csv \
         --template daily_digest.html \
         --subject "Daily Digest" \
         --interval "24h"
```

**Common intervals:**
- `30s` - Every 30 seconds
- `5m` - Every 5 minutes
- `1h` - Every hour
- `24h` - Every 24 hours
- `168h` - Every week

### Cron-Based Scheduling

Use 5-field cron expressions for precise timing:

```bash
# Daily at 9:00 AM
mailgrid --env config.json \
         --csv subscribers.csv \
         --template daily_newsletter.html \
         --subject "Daily Newsletter" \
         --cron "0 9 * * *"

# Weekly on Monday at 9:00 AM
mailgrid --env config.json \
         --csv weekly_subscribers.csv \
         --template weekly_report.html \
         --subject "Weekly Report" \
         --cron "0 9 * * 1"

# Monthly on the 1st at 10:00 AM
mailgrid --env config.json \
         --csv monthly_subscribers.csv \
         --template monthly_summary.html \
         --subject "Monthly Summary" \
         --cron "0 10 1 * *"
```

**Cron format:** `minute hour day month weekday`

| Field | Values | Special |
|-------|--------|---------|
| minute | 0-59 | * , - / |
| hour | 0-23 | * , - / |
| day | 1-31 | * , - / |
| month | 1-12 | * , - / |
| weekday | 0-7 (0=Sunday) | * , - / |

**Common cron expressions:**
- `0 9 * * *` - Daily at 9:00 AM
- `0 9 * * 1` - Weekly on Monday at 9:00 AM
- `0 9 1 * *` - Monthly on 1st at 9:00 AM
- `0 9 * * 1-5` - Weekdays at 9:00 AM
- `*/30 * * * *` - Every 30 minutes

---

## Scheduler Management

### Auto-Start Behavior

The scheduler automatically starts when jobs are scheduled:

```bash
# This automatically starts the scheduler
mailgrid --env config.json \
         --csv recipients.csv \
         --template email.html \
         --cron "0 9 * * *"

# Scheduler runs in background and processes the job
# No manual intervention required
```

### Manual Scheduler Control

Run the scheduler in foreground mode:

```bash
mailgrid --scheduler-run --env config.json
```

This is useful for:
- Debugging scheduled jobs
- Running as a system service
- Monitoring scheduler activity

Press `Ctrl+C` to stop the scheduler.

### Custom Database Location

Specify a custom database file for job storage:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template email.html \
         --cron "0 9 * * *" \
         --scheduler-db "/path/to/custom-jobs.db"
```

**Default:** `mailgrid.db` in current directory

---

## Job Management

### List Scheduled Jobs

View all scheduled jobs and their status:

```bash
mailgrid --jobs-list --env config.json
```

**Output includes:**
- Job ID and name
- Schedule type (one-time, interval, cron)
- Next execution time
- Status (pending, running, completed, failed)
- Creation time

### Cancel Jobs

Remove a scheduled job by its ID:

```bash
mailgrid --jobs-cancel "job-id-123" --env config.json
```

Get job IDs from the `--jobs-list` command.

### Job Status Monitoring

Jobs can have the following statuses:
- **Pending**: Waiting for execution time
- **Running**: Currently executing
- **Completed**: Successfully finished (one-time jobs only)
- **Failed**: Execution failed (after all retries)
- **Cancelled**: Manually cancelled

---

## Advanced Configuration

### Retry Configuration

Configure scheduler-level retries for job execution:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template email.html \
         --cron "0 9 * * *" \
         --job-retries 5 \
         --job-backoff 5s
```

**Parameters:**
- `--job-retries`: Number of retry attempts (default: 3)
- `--job-backoff`: Base backoff duration (default: 2s)

### Retry Behavior

- **Exponential backoff**: Delay increases with each retry
- **Jitter**: Random component to avoid thundering herd
- **Maximum backoff**: Capped at 5 minutes
- **Separate from SMTP retries**: These are scheduler-level retries

### Integration with Monitoring

Combine scheduling with real-time monitoring:

```bash
mailgrid --env config.json \
         --csv large_list.csv \
         --template campaign.html \
         --subject "Special Offer" \
         --cron "0 9 * * 1" \
         --monitor --monitor-port 9092 \
         --concurrency 10
```

---

## Production Deployment

### System Service Setup

**Linux (systemd):**

Create `/etc/systemd/system/mailgrid.service`:
```ini
[Unit]
Description=MailGrid Scheduler
After=network.target

[Service]
Type=simple
User=mailgrid
WorkingDirectory=/opt/mailgrid
ExecStart=/usr/local/bin/mailgrid --scheduler-run --env config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable mailgrid
sudo systemctl start mailgrid
```

**Windows (Service):**

Use tools like [NSSM](https://nssm.cc/) to run as a Windows service:
```cmd
nssm install MailGrid "C:\mailgrid\mailgrid.exe"
nssm set MailGrid Parameters "--scheduler-run --env C:\mailgrid\config.json"
nssm set MailGrid AppDirectory "C:\mailgrid"
nssm start MailGrid
```

### Docker Deployment

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  mailgrid-scheduler:
    image: ghcr.io/bravo1goingdark/mailgrid:latest
    command: ["--scheduler-run", "--env", "/app/config.json"]
    volumes:
      - ./config.json:/app/config.json
      - ./data:/app/data
      - ./templates:/app/templates
      - mailgrid-db:/app/db
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "mailgrid", "--jobs-list"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  mailgrid-db:
```

---

## Use Cases

### Newsletter Campaigns

Weekly newsletter every Monday at 9 AM:
```bash
mailgrid --env config.json \
         --csv subscribers.csv \
         --template newsletter.html \
         --subject "Weekly Newsletter - {{.date}}" \
         --cron "0 9 * * 1" \
         --monitor
```

### Reminder System

Daily reminders at 9 AM on weekdays:
```bash
mailgrid --env config.json \
         --csv team_members.csv \
         --template daily_standup.html \
         --subject "Daily Standup Reminder" \
         --cron "0 9 * * 1-5"
```

### Report Distribution

Monthly reports on the first day of each month:
```bash
mailgrid --env config.json \
         --to "executives@company.com" \
         --template monthly_report.html \
         --subject "Monthly Business Report" \
         --attach report.pdf \
         --cron "0 8 1 * *"
```

### Follow-up Campaigns

Send follow-up emails 3 days after signup:
```bash
# This would typically be triggered by an external system
mailgrid --env config.json \
         --to "{{.user_email}}" \
         --template followup.html \
         --subject "How are you finding our service?" \
         --schedule-at "2025-01-18T10:00:00Z"
```

---

## Monitoring & Metrics

### Health Check Endpoint

When the scheduler is running, access health information:
```bash
curl http://localhost:8090/health
```

### Metrics Endpoint

Get detailed scheduler metrics:
```bash
curl http://localhost:8090/metrics
```

**Metrics include:**
- Active job count
- Execution success/failure rates
- Average execution time
- Queue depth
- Last execution times

---

## Troubleshooting

### Common Issues

**Jobs not executing:**
1. Check system time and timezone
2. Verify scheduler is running: `--jobs-list`
3. Check database permissions
4. Review log output for errors

**Scheduler won't start:**
1. Check database file permissions
2. Verify port 8090 isn't in use
3. Check config file validity
4. Ensure SMTP credentials are correct

**Jobs failing:**
1. Test email configuration manually
2. Check recipient CSV file format
3. Verify template file exists and is valid
4. Review SMTP server logs

### Debug Mode

Run scheduler with verbose logging:
```bash
MAILGRID_LOG_LEVEL=debug mailgrid --scheduler-run --env config.json
```

---

## Next Steps

- [Real-Time Monitoring](monitoring) - Track scheduled campaigns
- [CLI Reference](cli-reference) - Complete scheduling flags
- [Advanced Usage](advanced) - Complex scheduling scenarios