# MailGrid CLI Reference

Complete command-line interface reference for MailGrid's advanced email automation and scheduling features.

## Quick Reference

```bash
mailgrid [OPTIONS] [EMAIL_CONTENT] [DATA_SOURCE] [SCHEDULING]
```

## Core Email Options

### Recipients
- `--to <email>` - Single recipient email address
- `--cc <emails|file>` - CC recipients (comma-separated or file path)
- `--bcc <emails|file>` - BCC recipients (comma-separated or file path)

### Content
- `--subject <text>` - Email subject line (supports templating)
- `--text <message|file>` - Plain text message or path to .txt file
- `--template <file>` - Path to HTML template file
- `--attach <files...>` - File attachments (repeat flag for multiple)

## Data Sources

### CSV Files
```bash
--csv <path>              # Path to CSV file with recipient data
```

**CSV Format:**
```csv
email,name,company,custom_field
user@example.com,John Doe,Acme Corp,premium
```

### Google Sheets
```bash
--sheet-url <url>         # Public Google Sheets URL
```

**Supported formats:**
- `https://docs.google.com/spreadsheets/d/{id}/edit#gid={gid}`
- `https://docs.google.com/spreadsheets/d/{id}/edit`

## Advanced Scheduling

### One-Time Scheduling
```bash
--schedule-at <timestamp>  # RFC3339 timestamp
```

**Examples:**
```bash
# Schedule for specific date/time
mailgrid --to "user@example.com" \
         --subject "Reminder" \
         --text "Meeting tomorrow!" \
         --schedule-at "2025-01-15T14:30:00Z" \
         --env config.json

# Schedule for 1 hour from now
mailgrid --to "team@company.com" \
         --subject "Status Update" \
         --text "Project status..." \
         --schedule-at "$(date -d '+1 hour' -Iseconds)" \
         --env config.json
```

### Recurring Schedules

#### Interval-Based
```bash
--interval <duration>      # Go duration format
```

**Supported durations:**
- `30s` - 30 seconds  
- `5m` - 5 minutes
- `1h` - 1 hour
- `24h` - 24 hours
- `168h` - 1 week

**Example:**
```bash
# Send newsletter every 6 hours
mailgrid --csv subscribers.csv \
         --template newsletter.html \
         --subject "News Update - {{.timestamp}}" \
         --interval "6h" \
         --env config.json
```

#### Cron Expressions
```bash
--cron <expression>        # 5-field cron format
```

**Cron format:** `minute hour day month weekday`

**Common patterns:**
```bash
"0 9 * * *"        # Daily at 9:00 AM
"30 8 * * 1-5"     # Weekdays at 8:30 AM
"0 0 1 * *"        # Monthly on 1st at midnight
"0 */4 * * *"      # Every 4 hours
"15 14 1 * *"      # Monthly on 1st at 2:15 PM
```

**Example:**
```bash
# Daily report at 9 AM
mailgrid --to "admin@company.com" \
         --subject "Daily Report - {{.date}}" \
         --template daily_report.html \
         --cron "0 9 * * *" \
         --env config.json
```

### Job Management

#### List Jobs
```bash
--jobs-list             # List all scheduled jobs
-L                      # Short form
```

**Output format:**
```
📋 Found 3 job(s):

JOB ID               STATUS     RUN AT               NEXT RUN             ATTEMPTS  
-------------------------------------------------------------------------------------
175933162200481900... pending    20:50:00 10/01       -                    0/3
175933181200481901... running    09:00:00 10/02       09:00:00 10/03       1/3
175933201200481902... done       15:30:00 10/01       -                    3/3

💡 Use --jobs-cancel <JOB_ID> to cancel a specific job
```

#### Cancel Jobs
```bash
--jobs-cancel <job_id>  # Cancel specific job by ID
-X <job_id>             # Short form
```

**Example:**
```bash
# Cancel a specific job
mailgrid --jobs-cancel "175933162200481900-8094455869530906480" \
         --env config.json
```

#### Scheduler Daemon
```bash
--scheduler-run         # Run scheduler as daemon
-R                      # Short form
```

**Example:**
```bash
# Run as background service
mailgrid --scheduler-run --env config.json
```

**Features:**
- Processes all scheduled jobs automatically
- Graceful shutdown with Ctrl+C
- Metrics server on port 8090
- Automatic job retry and error handling

## Performance & Reliability

### Concurrency Control
```bash
--concurrency <number>     # Number of concurrent workers (default: 1)
-c <number>               # Short form
```

**Guidelines:**
- Gmail: 1-2 workers
- SendGrid/Mailgun: 5-10 workers
- Amazon SES: 10-20 workers

### Batch Processing
```bash
--batch-size <number>      # Emails per batch (default: 1)
```

**Recommended sizes:**
- Single/transactional: 1
- Newsletter: 10-50
- Bulk campaigns: 50-100

### Retry Configuration
```bash
--retries <number>         # Per-email retry attempts (default: 1)
--job-retries <number>     # Scheduler-level retries (default: 3)
--job-backoff <duration>   # Retry backoff duration (default: 2s)
-J <number>               # Short form for job-retries
-B <duration>             # Short form for job-backoff
```

## Template System

### Template Syntax
MailGrid uses Go templates with the following data available:

**Standard fields from CSV:**
- `{{.email}}` - Recipient email
- `{{.name}}` - Recipient name  
- `{{.company}}` - Company name
- `{{.custom_field}}` - Any CSV column

**System variables:**
- `{{.timestamp}}` - Current timestamp
- `{{.date}}` - Current date
- `{{.time}}` - Current time

### Template Examples

**HTML Template:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>Welcome {{.name}}!</title>
</head>
<body>
    <h1>Hello {{.name}},</h1>
    <p>Welcome to {{.company}}!</p>
    
    {{if .premium}}
    <div class="premium-content">
        🎉 Premium features unlocked!
    </div>
    {{end}}
    
    <p>Best regards,<br>
    The Team</p>
</body>
</html>
```

**Subject Templates:**
```bash
--subject "Welcome to {{.company}}, {{.name}}!"
--subject "{{.name}}, your {{.plan}} account is ready"
--subject "Monthly report for {{.department}}"
```

## Filtering & Selection

### Logical Filters
```bash
--filter <expression>      # Logical filter for recipients
```

**Operators:**
- `==` - Equals
- `!=` - Not equals  
- `contains` - String contains
- `startswith` - String starts with
- `endswith` - String ends with
- `>`, `<` - Numeric comparison
- `AND`, `OR` - Logical operators
- `NOT` - Negation
- `()` - Grouping

**Examples:**
```bash
# Premium users only
--filter "plan == 'premium'"

# Specific companies
--filter "company contains 'Tech' OR company == 'StartupCorp'"

# Complex filtering
--filter "(plan == 'premium' AND region != 'EU') OR vip == true"

# Numeric filtering  
--filter "age > 25 AND purchase_count >= 5"
```

## Configuration & Environment

### Configuration File
```bash
--env <path>              # Path to JSON config file
```

**Config format:**
```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from_email": "your-email@gmail.com", 
    "from_name": "Your Name",
    "use_tls": true
  }
}
```

### Database Location
```bash
--scheduler-db <path>     # Path to scheduler database (default: mailgrid.db)
-D <path>                 # Short form
```

## Debugging & Testing

### Preview Mode
```bash
--preview                 # Start preview server
--port <number>           # Preview server port (default: 8080)
-p                        # Short form for preview
```

### Dry Run
```bash
--dry-run                 # Render emails without sending
```

**Output example:**
```
=== DRY RUN MODE ===
TO: john@example.com
CC: manager@company.com
BCC: 
SUBJECT: Welcome John!
BODY:
---
<html>
<body>
<h1>Hello John,</h1>
<p>Welcome to Acme Corp!</p>
</body>
</html>
---
ATTACHMENTS: contract.pdf, welcome.txt
===================
```

## Monitoring & Metrics

### Metrics Endpoints
When scheduler is active, monitoring endpoints are available:

```bash
curl http://localhost:8090/metrics   # Performance metrics
curl http://localhost:8090/health    # Health status
```

**Metrics response:**
```json
{
  "uptime": 3600000,
  "emails_sent": 1250,
  "emails_failed": 23,
  "avg_delivery_time": 847,
  "active_connections": 5,
  "batches_processed": 125,
  "avg_batch_size": 10,
  "batch_success_rate": 0.98,
  "error_counts": {
    "connection_timeout": 15,
    "smtp_error": 8
  },
  "template_cache_hits": 1180,
  "throttle_events": 2
}
```

## Complete Examples

### Immediate Bulk Email
```bash
mailgrid --csv contacts.csv \
         --template newsletter.html \
         --subject "Newsletter - {{.name}}" \
         --attach newsletter.pdf \
         --cc "manager@company.com" \
         --concurrency 5 \
         --batch-size 20 \
         --retries 3 \
         --filter "subscribed == true AND region != 'GDPR'" \
         --env production.json
```

### Scheduled Campaign
```bash
mailgrid --sheet-url "https://docs.google.com/spreadsheets/d/abc123/edit" \
         --template campaign.html \
         --subject "Special Offer - {{.name}}!" \
         --attach terms.pdf \
         --schedule-at "2025-01-15T10:00:00Z" \
         --job-retries 5 \
         --job-backoff "30s" \
         --env config.json
```

### Recurring Newsletter
```bash
mailgrid --csv subscribers.csv \
         --template weekly_newsletter.html \
         --subject "Weekly Update - Week of {{.date}}" \
         --cron "0 9 * * 1" \
         --concurrency 10 \
         --batch-size 50 \
         --env newsletter.json
```

### Job Management Session
```bash
# Schedule a job
mailgrid --to "test@example.com" \
         --subject "Test" \
         --text "Testing scheduler" \
         --schedule-at "2025-01-01T12:00:00Z" \
         --env config.json

# List all jobs  
mailgrid --jobs-list --env config.json

# Cancel a job
mailgrid --jobs-cancel "job-id-123" --env config.json

# Run daemon
mailgrid --scheduler-run --env config.json
```

## Exit Codes

- `0` - Success
- `1` - Configuration error
- `2` - Invalid arguments  
- `3` - SMTP connection error
- `4` - Template parsing error
- `5` - File I/O error

## Support & Troubleshooting

For detailed troubleshooting guides and best practices, see:
- [README.md](../README.md) - Main documentation
- [examples/](../examples/) - Usage examples
- [Issues](https://github.com/bravo1goingdark/mailgrid/issues) - Bug reports and support

---

**MailGrid CLI Reference v2.0** - Updated for optimized scheduler features