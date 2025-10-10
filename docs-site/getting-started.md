---
layout: default
title: Getting Started
nav_order: 2
has_children: false
---

# Getting Started
{: .no_toc }

Get up and running with MailGrid in minutes.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Installation

### Quick Install (Recommended)

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex
```

**Linux & macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash
```

**Docker:**
```bash
docker run --rm ghcr.io/bravo1goingdark/mailgrid:latest --help
```

For detailed installation instructions, see the [Installation Guide](installation).

---

## Configuration

Create your SMTP configuration file (`config.json`):

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

### Gmail Setup

For Gmail, you'll need an App Password:
1. Enable 2-Factor Authentication on your Google Account
2. Go to [App Passwords](https://myaccount.google.com/apppasswords)
3. Generate a new app password
4. Use this password in your `config.json`

---

## Your First Email

### Single Email

Send a quick test email:

```bash
mailgrid --env config.json \
         --to test@example.com \
         --subject "MailGrid Test" \
         --text "Hello from MailGrid!" \
         --dry-run
```

Remove `--dry-run` to actually send the email.

### Bulk Emails

1. **Create a CSV file** (`recipients.csv`):
   ```csv
   email,name,company
   john@example.com,John Doe,Acme Corp
   jane@example.com,Jane Smith,Tech Inc
   ```

2. **Create an HTML template** (`email.html`):
   ```html
   <!DOCTYPE html>
   <html>
   <head>
       <title>Welcome</title>
   </head>
   <body>
       <h1>Hello {{.name}}!</h1>
       <p>Welcome to {{.company}}.</p>
       <p>We're excited to have you on board.</p>
   </body>
   </html>
   ```

3. **Send bulk emails**:
   ```bash
   mailgrid --env config.json \
            --csv recipients.csv \
            --template email.html \
            --subject "Welcome {{.name}}!" \
            --dry-run
   ```

---

## Preview Your Emails

Before sending, preview how your emails will look:

```bash
mailgrid --preview \
         --csv recipients.csv \
         --template email.html \
         --port 8080
```

Open `http://localhost:8080` in your browser to see the rendered emails.

---

## Monitor Campaign Progress

For bulk campaigns, enable real-time monitoring:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template email.html \
         --subject "Welcome {{.name}}!" \
         --monitor \
         --concurrency 5
```

This opens a monitoring dashboard at `http://localhost:9091` showing:
- Live recipient status
- Campaign metrics
- Performance analytics
- Error tracking

---

## Performance Optimization

### Concurrency

Use multiple workers for faster sending:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template email.html \
         --concurrency 5 \
         --retries 3
```

{: .warning }
> **Rate Limits**: Keep concurrency ≤ 5 for consumer email providers (Gmail, Yahoo, Outlook) to avoid rate limiting.

### Batch Processing

Group emails for better throughput:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template email.html \
         --batch-size 5 \
         --concurrency 3
```

---

## Next Steps

- [CLI Reference](cli-reference) - Complete flag documentation
- [Monitoring](monitoring) - Real-time campaign tracking
- [Scheduling](scheduling) - Automated email campaigns
- [Advanced Usage](advanced) - Filters, webhooks, and more

---

## Need Help?

- Check the [FAQ](faq) for common questions
- Browse [Examples](examples) for real-world scenarios
- Report issues on [GitHub](https://github.com/bravo1goingdark/mailgrid/issues)