<p align="center">
  <img src="./assets/readme-banner-mailgrid.svg" alt="Mailgrid Logo" width="100%" height="100%"/>
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/bravo1goingdark/mailgrid">
    <img src="https://img.shields.io/badge/Go%20Report-C-yellow?style=for-the-badge&logo=go" alt="Go Report Card"/>
  </a>
  <a href="https://github.com/bravo1goingdark/mailgrid/tree/main/docs/docs.md">
    <img src="https://img.shields.io/badge/📘 Documentation-blue?style=for-the-badge" alt="Docs Badge"/>
  </a>
  <a href="https://github.com/bravo1goingdark/blipmq">
    <img src="https://img.shields.io/badge/Built%20by-BlipMQ-8E44AD?style=for-the-badge&logo=github" alt="Built By BlipMQ"/>
  </a>
</p>



**Mailgrid** is a high-performance, ultra-lightweight CLI tool written in Go for sending bulk emails via SMTP from CSV or Google Sheets. Built for speed, reliability, and minimalism — no bloated web UIs, just powerful automation.



---

## 🚀 Features

Mailgrid is a fast, minimal CLI tool for sending personalized emails from CSV files or Google Sheets via SMTP — no web UI, just powerful automation.

---

### 📬 Email Capabilities
- **Bulk email sending** from CSV files **or public Google Sheets**
- **Dynamic templating** for subject lines and HTML body using Go’s `text/template`
- **File attachments** (up to 10 MB each)
- **CC/BCC support** via inline lists or files

---

### ⚙️ Configuration & Control
- **SMTP support** with simple `config.json`
- **Concurrency, batching, and automatic retries** for high throughput
- **Preview server** (`--preview`) to view rendered emails in the browser
- **Dry-run mode** (`--dry-run`) to render without sending
- **Logical recipient filtering** using `--filter`
- **Success and failure logs** written to CSV

---

### 🛠️ Developer Experience
- **Built with Go** — fast, static binary with zero dependencies
- **Cross-platform support** — runs on Linux, macOS, and Windows
- **Live CLI logs** for each email: success ✅ or failure ❌
- **Missing field warnings** for incomplete CSV rows

---

### ⏱️ Advanced Scheduling & Automation
Mailgrid features a high-performance, production-ready scheduler with auto-start capabilities, monitoring, and intelligent lifecycle management.

#### 🚀 Auto-Start Scheduler
- **Automatic activation**: Scheduler starts automatically when jobs are scheduled
- **Auto-shutdown**: Intelligent shutdown after configurable idle periods (default: 5 minutes)
- **Background operation**: Jobs execute seamlessly without manual intervention
- **Persistent storage**: BoltDB-backed job persistence with distributed locking
- **Metrics & monitoring**: Built-in HTTP endpoints for performance tracking

#### 📅 Flexible Scheduling Options

**One-time scheduling:**
```bash
mailgrid \
  --env config.json \
  --to "user@example.com" \
  --subject "Reminder" \
  --text "Don't forget the meeting!" \
  --schedule-at "2025-01-01T10:00:00Z"
```

**Recurring schedules:**
```bash
# Every 30 minutes
mailgrid \
  --env config.json \
  --csv subscribers.csv \
  --template newsletter.html \
  --subject "Updates - {{.name}}" \
  --interval "30m"

# Daily at 9 AM (cron)
mailgrid \
  --env config.json \
  --to "admin@company.com" \
  --subject "Daily Report" \
  --text "Here's today's summary..." \
  --cron "0 9 * * *"
```

#### 🎛️ Job Management
```bash
# List all scheduled jobs with status
mailgrid --jobs-list --env config.json

# Cancel a specific job
mailgrid --jobs-cancel "job-id-123" --env config.json

# Run scheduler as a daemon service
mailgrid --scheduler-run --env config.json
```

#### 📊 Monitoring & Metrics
When active, the scheduler provides real-time metrics:
- **Metrics endpoint**: `http://localhost:8090/metrics`
- **Health check**: `http://localhost:8090/health`
- **Performance data**: Delivery times, success rates, connection status
- **Error tracking**: Detailed error classification and counts

See the full flag reference and examples in [docs/docs.md](./docs/docs.md).

---

### 🏗️ Performance Features
- **Connection pooling**: Optimized SMTP connection management
- **Batch processing**: Intelligent email batching with adaptive sizing
- **Template caching**: Lightning-fast template rendering (1-hour cache)
- **Circuit breaking**: Automatic failover during SMTP issues
- **Concurrent execution**: Multi-threaded job processing
- **Adaptive polling**: Dynamic interval adjustment based on workload
- **Resilience management**: Retry logic with exponential backoff

### 📈 Production Ready
- **Zero-downtime operation**: Graceful shutdown with signal handling
- **Resource efficiency**: Automatic cleanup and memory management
- **Error recovery**: Intelligent retry mechanisms and circuit breaking
- **Monitoring integration**: JSON metrics for external monitoring tools
- **Database persistence**: Reliable job storage with crash recovery

> 📄 Licensed under BSD-3-Clause — see [LICENSE](./LICENSE)






