---
layout: default
title: Home
nav_order: 1
description: "MailGrid - High-performance CLI tool for bulk email sending via SMTP"
permalink: /
---

# MailGrid Documentation
{: .fs-9 }

High-performance, ultra-lightweight CLI tool for sending bulk emails via SMTP from CSV or Google Sheets. Built for speed, reliability, and minimalism.
{: .fs-6 .fw-300 }

[Get started now](getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/bravo1goingdark/mailgrid){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## Quick Start

Send a single email:
```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"
```

Send bulk emails from CSV:
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"
```

Monitor campaign progress:
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --monitor --concurrency 5
```

---

## Key Features

### 📬 Email Capabilities
- **Bulk email sending** from CSV files or public Google Sheets
- **Dynamic templating** for subject lines and HTML body using Go's `text/template`
- **File attachments** (up to 10 MB each)
- **CC/BCC support** via inline lists or files

### ⚙️ Configuration & Control
- **SMTP support** with simple `config.json`
- **Concurrency, batching, and automatic retries** for high throughput
- **Real-time monitoring dashboard** with live metrics and recipient tracking
- **Preview server** to view rendered emails in the browser
- **Dry-run mode** to render without sending
- **Logical recipient filtering** using expressions
- **Webhook notifications** for campaign completion events

### ⏱️ Advanced Scheduling
- **One-time scheduling** with RFC3339 timestamps
- **Recurring schedules** using intervals or cron expressions
- **Auto-start scheduler** with intelligent shutdown
- **Persistent storage** with BoltDB-backed job persistence
- **Job management** with list, cancel, and monitoring capabilities

### 🚀 Performance Features
- **Connection pooling** for optimized SMTP management
- **Batch processing** with adaptive sizing
- **Template caching** for lightning-fast rendering
- **Circuit breaking** for automatic failover
- **Concurrent execution** with multi-threaded processing
- **Resilience management** with exponential backoff

---

## Installation

### Quick Install

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

See the [Installation Guide](installation) for detailed instructions and all available methods.

---

## About

MailGrid is licensed under the [BSD-3-Clause License](https://github.com/bravo1goingdark/mailgrid/blob/main/LICENSE).