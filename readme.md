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
  <a href="https://github.com/bravo1goingdark/mailgrid/blob/main/INSTALLATION.md">
    <img src="https://img.shields.io/badge/🚀 Installation-green?style=for-the-badge" alt="Installation Guide"/>
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
- **Real-time monitoring dashboard** (`--monitor`) with live metrics and recipient tracking
- **Preview server** (`--preview`) to view rendered emails in the browser
- **Dry-run mode** (`--dry-run`) to render without sending
- **Logical recipient filtering** using `--filter`
- **Webhook notifications** for campaign completion events
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

---

## 📦 Installation

MailGrid provides multiple installation methods for all major platforms with optimized, lightweight binaries.

### 🚀 Quick Install (Recommended)

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

### 📥 Direct Downloads

Download pre-built binaries from [GitHub Releases](https://github.com/bravo1goingdark/mailgrid/releases/latest):

| Platform | Architecture | Download | Size |
|----------|--------------|----------|------|
| **Windows** | x64 | [📥 mailgrid-windows-amd64.exe.zip](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-amd64.exe.zip) | ~3.7 MB |
| **Windows** | ARM64 | [📥 mailgrid-windows-arm64.exe.zip](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-arm64.exe.zip) | ~3.3 MB |
| **macOS** | Intel | [📥 mailgrid-macos-intel.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-macos-intel.tar.gz) | ~4.2 MB |
| **macOS** | Apple Silicon | [📥 mailgrid-macos-apple-silicon.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-macos-apple-silicon.tar.gz) | ~4.0 MB |
| **Linux** | x64 | [📥 mailgrid-linux-amd64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-amd64.tar.gz) | ~4.1 MB |
| **Linux** | ARM64 | [📥 mailgrid-linux-arm64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-arm64.tar.gz) | ~3.9 MB |
| **FreeBSD** | x64 | [📥 mailgrid-freebsd-amd64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-freebsd-amd64.tar.gz) | ~4.1 MB |

### 📋 Manual Installation Steps

**Windows:**
1. Download the appropriate `.zip` file for your architecture
2. Extract `mailgrid.exe` to a folder (e.g., `C:\tools\mailgrid\`)
3. Add the folder to your system PATH
4. Verify: `mailgrid --help`

**Linux/macOS:**
1. Download and extract the `.tar.gz` file:
   ```bash
   tar -xzf mailgrid-*.tar.gz
   sudo mv mailgrid /usr/local/bin/
   chmod +x /usr/local/bin/mailgrid
   ```
2. Verify: `mailgrid --help`

### 🐳 Docker Usage

**Quick test:**
```bash
docker run --rm -v $(pwd)/config.json:/app/config.json \
  ghcr.io/bravo1goingdark/mailgrid:latest \
  --env /app/config.json --to test@example.com --subject "Test" --text "Hello!"
```

**With Docker Compose:**
```bash
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/docker-compose.yml -o docker-compose.yml
docker-compose up -d
```

### ⚙️ Configuration Setup

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

### ✅ Quick Test

Verify your installation with a dry-run:
```bash
mailgrid --env config.json \
         --to test@example.com \
         --subject "MailGrid Test" \
         --text "Hello from MailGrid v1.0.0!" \
         --dry-run
```

> 📖 **For detailed installation instructions, troubleshooting, and advanced options, see [INSTALLATION.md](./INSTALLATION.md)**

---

## 🚦 Quick Start

**Single email:**
```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"
```

**Bulk emails from CSV:**
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --subject "Hi {{.name}}!"
```

**Schedule recurring emails:**
```bash
mailgrid --env config.json --to user@example.com --subject "Weekly Report" --text "Here's your report..." --cron "0 9 * * 1"
```

**Monitor campaign progress:**
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --monitor --concurrency 5
```

**Preview before sending:**
```bash
mailgrid --env config.json --csv recipients.csv --template email.html --preview --port 8080
```

> 📚 **For comprehensive usage examples and CLI reference, see [docs/docs.md](./docs/docs.md)**

---

> 📄 Licensed under BSD-3-Clause — see [LICENSE](./LICENSE)

