# Mailgrid — Complete Reference

> Full flag reference, configuration guide, and operational patterns for production use

---

## Table of Contents

- [Quick Start](#quick-start)
- [SMTP Configuration](#smtp-configuration)
  - [Required Fields](#required-fields)
  - [TLS Options](#tls-options)
  - [Provider Configs](#provider-configs)
- [Recipient Source](#recipient-source)
  - [--csv](#--csv---f)
  - [--sheet-url](#--sheet-url---u)
  - [--to](#--to)
- [Email Content](#email-content)
  - [--template](#--template---t)
  - [--text](#--text)
  - [--subject](#--subject---s)
  - [--attach](#--attach---a)
  - [--cc / --bcc](#--cc----bcc)
- [Delivery Options](#delivery-options)
  - [--concurrency](#--concurrency---c)
  - [--batch-size](#--batch-size---b)
  - [--retries](#--retries---r)
  - [--smtp-timeout](#--smtp-timeout)
- [Monitoring](#monitoring)
  - [--monitor](#--monitor---m)
  - [--monitor-port](#--monitor-port)
  - [--monitor-client-timeout](#--monitor-client-timeout)
  - [Endpoints](#endpoints)
  - [Prometheus Metrics](#prometheus-metrics)
- [Scheduling](#scheduling)
  - [--schedule-at](#--schedule-at---a)
  - [--interval](#--interval---i)
  - [--cron](#--cron---c)
  - [--scheduler-run](#--scheduler-run---r)
  - [--job-retries](#--job-retries---j)
  - [--db-path](#--db-path)
  - [--jobs-list / --jobs-cancel](#--jobs-list----jobs-cancel)
- [Notifications](#notifications)
  - [--webhook](#--webhook---w)
  - [--webhook-secret](#--webhook-secret)
- [Logging](#logging)
  - [--log-level](#--log-level)
  - [--log-format](#--log-format)
- [Offset Tracking](#offset-tracking)
  - [--resume](#--resume)
  - [--reset-offset](#--reset-offset)
- [Testing & Debug](#testing--debug)
  - [--dry-run](#--dry-run---d)
  - [--preview](#--preview---p)
- [Advanced Patterns](#advanced-patterns)
- [Delivery Logs](#delivery-logs)
- [Exit Codes](#exit-codes)
- [Quick Reference Table](#quick-reference-table)

---

## Quick Start

```bash
# 1. Minimal SMTP config
cat > config.json <<'EOF'
{
  "smtp": {
    "host": "smtp.gmail.com", "port": 587,
    "username": "you@gmail.com", "password": "app-password",
    "from": "You <you@gmail.com>"
  }
}
EOF

# 2. Send one email
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Hi!"

# 3. Bulk send with live dashboard
mailgrid --env config.json \
  --csv recipients.csv \
  --template email.html \
  --subject "Hi {{.name}}!" \
  --concurrency 5 --monitor
```

---

## SMTP Configuration

Pass with `--env` (`-e`). All five required fields are validated at startup — Mailgrid exits with an error before loading any recipients if any are missing.

### Required Fields

| Field | Type | Description |
|---|---|---|
| `host` | string | SMTP server hostname |
| `port` | int | Port — `587` (STARTTLS), `465` (TLS), `25` (unencrypted) |
| `username` | string | SMTP auth username |
| `password` | string | SMTP auth password |
| `from` | string | Envelope and header `From`; may include display name |

```json
{
  "smtp": {
    "host":     "smtp.gmail.com",
    "port":     587,
    "username": "you@gmail.com",
    "password": "your-app-password",
    "from":     "You <you@gmail.com>"
  }
}
```

### TLS Options

| Field | Type | Default | Description |
|---|---|---|---|
| `tls_cert_file` | string | — | Custom CA certificate path (PEM). Required for private-CA SMTP servers. |
| `tls_key_file` | string | — | Client certificate key path (PEM). Provide alongside `tls_cert_file` for mutual TLS. |
| `insecure_tls` | bool | `false` | Disable certificate verification. Emits a security warning. **Never use in production.** |

**Behavior:**
- TLS 1.2+ is enforced on all connections.
- STARTTLS is negotiated when the server advertises it; the connection continues without it if not offered.
- Misconfigured `tls_cert_file` or `tls_key_file` paths are a hard error — Mailgrid never silently falls back to system certificates.
- When `insecure_tls: true`, the following warning is printed to stderr before every run:
  ```
  SECURITY WARNING: TLS certificate verification is disabled (insecure_tls=true).
  This connection is vulnerable to man-in-the-middle attacks.
  ```

### Provider Configs

**Gmail** — requires a [Google App Password](https://support.google.com/accounts/answer/185833):

```json
{ "smtp": { "host": "smtp.gmail.com", "port": 587,
            "username": "you@gmail.com", "password": "app-password",
            "from": "You <you@gmail.com>" } }
```

**Amazon SES (SMTP interface):**

```json
{ "smtp": { "host": "email-smtp.us-east-1.amazonaws.com", "port": 587,
            "username": "AKIAIOSFODNN7EXAMPLE",
            "password": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
            "from": "noreply@verified-domain.com" } }
```

**SendGrid:**

```json
{ "smtp": { "host": "smtp.sendgrid.net", "port": 587,
            "username": "apikey", "password": "SG.your-api-key",
            "from": "noreply@yourdomain.com" } }
```

**Mailgun:**

```json
{ "smtp": { "host": "smtp.mailgun.org", "port": 587,
            "username": "postmaster@mg.yourdomain.com",
            "password": "your-mailgun-smtp-password",
            "from": "noreply@yourdomain.com" } }
```

---

## Recipient Source

Provide exactly one of `--csv`, `--sheet-url`, or `--to`.

---

### `--csv` / `-f`

```
--csv <path>
```

Path to a CSV file. The `email` column is required; all other columns become template variables. Column names are lowercased before use.

**Behavior:**
- Rows with a missing or invalid `email` are skipped and logged.
- Duplicate email addresses are deduplicated (case-insensitive) before sending. A count of removed duplicates is logged.

**Example:**

```csv
email,name,company,tier
alice@example.com,Alice,Acme,premium
bob@example.com,Bob,Beta,free
alice@example.com,Alice,Acme,premium  ← duplicate, removed
```

```bash
mailgrid --env config.json --csv recipients.csv --template email.html
```

---

### `--sheet-url` / `-u`

```
--sheet-url <url>
```

URL of a public Google Sheet. The sheet must have "Anyone with the link can view" access.

**Behavior:**
- The URL hostname is validated against `docs.google.com` before any HTTP request — arbitrary URLs are rejected to prevent SSRF.
- Redirects to non-Google hosts are blocked.
- The first row is used as the header; column names become template variables.

**Example:**

```bash
mailgrid --env config.json \
  --sheet-url "https://docs.google.com/spreadsheets/d/SHEET_ID/edit?gid=0" \
  --template email.html --subject "Newsletter"
```

---

### `--to`

```
--to <email>
```

Single recipient address. Mutually exclusive with `--csv` and `--sheet-url`.

**Example:**

```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Hi!"
```

---

## Email Content

---

### `--template` / `-t`

```
--template <path>
```

Path to an HTML email template using Go `html/template` syntax. Every CSV column is available as `{{ .column_name }}`. The variable `{{ .email }}` is always available regardless of CSV columns.

**Behavior:**
- Templates are parsed once and cached in-process (LRU, 1-hour TTL).
- Paths are canonicalized — `./email.html` and `email.html` share the same cache entry.
- A recipient is skipped if template rendering fails (missing required variable, etc.). The skip is logged.

**Example:**

```html
<!-- email.html -->
<p>Hello {{ .name }},</p>
<p>Your <strong>{{ .tier }}</strong> plan is active on {{ .company }}.</p>
<p>Reply to this email or contact us at {{ .email }}.</p>
```

```bash
mailgrid --env config.json --csv recipients.csv --template email.html
```

---

### `--text`

```
--text <string | path>
```

Plain-text body. If the value is a path to an existing file, the file contents are used; otherwise the value itself is treated as the inline message.

**Behavior:**
- **With `--template`** — Mailgrid sends a `multipart/alternative` message: `text/plain` first (fallback), then `text/html`. Email clients that cannot render HTML display the plain-text version.
- **Without `--template`** — a plain `text/plain` message is sent.

**Message structure when both are provided:**

```
multipart/alternative
  ├── text/plain   (--text)
  └── text/html    (--template)
```

When attachments are also present (`--attach`), this becomes:

```
multipart/mixed
  ├── multipart/alternative
  │     ├── text/plain
  │     └── text/html
  └── attachment(s)
```

**Example:**

```bash
# Multipart: HTML + plain-text fallback
mailgrid --env config.json --csv recipients.csv \
  --template email.html --text fallback.txt \
  --subject "Your invoice"

# Plain-text only
mailgrid --env config.json --to user@example.com \
  --text "Hello, this is a plain-text message." \
  --subject "Hello"
```

---

### `--subject` / `-s`

```
--subject <string>     default: "Test Email from Mailgrid"
```

Subject line. Supports Go template syntax — recipient fields are available as variables.

**Behavior:**
- Non-ASCII characters are RFC 2047-encoded automatically.
- A recipient is skipped if subject template execution fails. The skip is logged.

**Example:**

```bash
--subject "Hi {{.name}}, your order from {{.company}} is ready"
--subject "Invoice #{{.invoice_number}} — due {{.due_date}}"
```

---

### `--attach` / `-a`

```
--attach <path>    (repeatable)
```

File to attach. Repeat the flag for each attachment. All attachments are sent to every recipient.

**Parameters:**
- Maximum file size: **10 MB** per attachment.
- MIME type is detected from the file extension; falls back to `application/octet-stream` for unknown extensions.

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --attach invoice.pdf \
  --attach terms.pdf
```

---

### `--cc` / `--bcc`

```
--cc  <addresses | path>
--bcc <addresses | path>
```

Carbon copy and blind carbon copy. Accepts a comma-separated address list or a path to a file with one address per line.

**Behavior:**
- CC recipients appear in the `CC` header and receive a copy.
- BCC addresses are added to the SMTP envelope only — never exposed in headers.
- Addresses that duplicate the primary recipient or each other are deduplicated. Each address receives exactly one RCPT command.

**Example:**

```bash
# Inline list
--cc "manager@example.com,audit@example.com" --bcc archive@example.com

# From file
--cc ./cc-list.txt --bcc ./bcc-list.txt
```

---

## Delivery Options

---

### `--concurrency` / `-c`

```
--concurrency <int>    default: 1
```

Number of parallel SMTP worker goroutines. Each worker maintains its own persistent SMTP connection with automatic reconnection on failure.

**Guidelines:**

| Provider | Recommended |
|---|---|
| Gmail | 1–2 |
| SendGrid / Mailgun | 5–10 |
| Amazon SES | 10–20 |
| Self-hosted Postfix | 2–5 |

**Example:**

```bash
--concurrency 10
```

---

### `--batch-size` / `-b`

```
--batch-size <int>    default: 1
```

Number of emails sent per SMTP DATA transaction. Higher values reduce connection overhead at the cost of larger failure blast radius per transaction.

**Guidelines:**
- Consumer inboxes (Gmail, Yahoo): `1`
- Warmed corporate IPs or transactional services: `5–10`
- Test with `--dry-run` before increasing.

**Example:**

```bash
--batch-size 5
```

---

### `--retries` / `-r`

```
--retries <int>    default: 1
```

Per-email retry attempts on failure.

**Backoff formula:** `min(2^attempt, 256)` seconds, plus up to 1 second of random jitter.

| Attempt | Min wait | Max wait |
|---|---|---|
| 1 | 2s | 3s |
| 2 | 4s | 5s |
| 3 | 8s | 9s |
| 8 | 256s | 257s |
| 9+ | 256s (capped) | 257s |

**Example:**

```bash
--retries 3
```

---

### `--smtp-timeout`

```
--smtp-timeout <seconds>    default: 10
```

TCP dial timeout for the initial SMTP connection. Does not affect the TLS handshake or AUTH exchange duration.

**Note:** Increase for high-latency SMTP relays. Decrease for fail-fast behaviour on flaky networks.

**Example:**

```bash
--smtp-timeout 30    # long-distance relay
--smtp-timeout 5     # fail fast on local network
```

---

## Monitoring

---

### `--monitor` / `-m`

```
--monitor
```

Starts an HTTP server alongside the campaign on `127.0.0.1`. See [Endpoints](#endpoints) for all available routes.

**Note:** The server binds to loopback only and is not reachable from other hosts.

---

### `--monitor-port`

```
--monitor-port <int>    default: 9091
```

Port for the monitoring HTTP server.

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --monitor --monitor-port 9091
# Dashboard: http://localhost:9091
# Metrics:   http://localhost:9091/metrics
```

---

### `--monitor-client-timeout`

```
--monitor-client-timeout <seconds>    default: 300
```

How long an idle SSE connection is kept open. Connections with no activity beyond this threshold are closed and removed.

**Note:** Maximum concurrent SSE connections is **50**. Excess connections receive `429 Too Many Requests`.

**Example:**

```bash
--monitor-client-timeout 600    # keep dashboard open for 10 minutes
```

---

### Endpoints

| Path | Description |
|---|---|
| `/` | Real-time SSE dashboard (browser UI) |
| `/api/status` | Current `CampaignStats` as JSON |
| `/api/stream` | Raw SSE event stream |
| `/metrics` | Prometheus text format |
| `/health` | Liveness check — `200 OK` |
| `/ready` | Readiness check — `200 OK` after 5s warmup |

---

### Prometheus Metrics

Available at `/metrics` when `--monitor` is enabled. Compatible with any Prometheus-compatible scraper.

```
# HELP mailgrid_emails_sent_total Total emails successfully sent
# TYPE mailgrid_emails_sent_total counter
mailgrid_emails_sent_total 1450

# HELP mailgrid_emails_failed_total Total emails that permanently failed
# TYPE mailgrid_emails_failed_total counter
mailgrid_emails_failed_total 12

# HELP mailgrid_emails_pending Total emails pending
# TYPE mailgrid_emails_pending gauge
mailgrid_emails_pending 538

# HELP mailgrid_emails_total Total recipients in campaign
# TYPE mailgrid_emails_total gauge
mailgrid_emails_total 2000

# HELP mailgrid_campaign_duration_seconds Elapsed seconds since campaign start
# TYPE mailgrid_campaign_duration_seconds gauge
mailgrid_campaign_duration_seconds 47.832
```

**Prometheus scrape config:**

```yaml
scrape_configs:
  - job_name: mailgrid
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9091']
```

---

## Scheduling

Jobs are persisted in BoltDB and survive process restarts. Each job holds a distributed lock to prevent duplicate execution when multiple Mailgrid processes share the same database. Panics in the job handler are recovered — the job is marked `failed` and the lock is released.

---

### `--schedule-at` / `-A`

```
--schedule-at <RFC3339>
```

One-time execution at a specific moment.

**Behavior:** If the time is already past, the job runs immediately and logs a warning.

**Example:**

```bash
--schedule-at "2025-06-15T09:00:00Z"
--schedule-at "2025-12-25T08:00:00+05:30"    # with timezone offset
```

---

### `--interval` / `-i`

```
--interval <duration>
```

Run the campaign repeatedly at this interval. Accepts any Go duration string.

**Example:**

```bash
--interval "30m"     # every 30 minutes
--interval "6h"      # every 6 hours
--interval "24h"     # daily
```

---

### `--cron` / `-C`

```
--cron <expression>
```

Standard 5-field cron expression. All times are UTC.

```
┌───── minute (0–59)
│ ┌───── hour (0–23)
│ │ ┌───── day of month (1–31)
│ │ │ ┌───── month (1–12)
│ │ │ │ ┌───── day of week (0–6, Sun=0)
│ │ │ │ │
* * * * *
```

**Example:**

```bash
--cron "0 9 * * 1-5"    # weekdays at 9 AM UTC
--cron "0 0 1 * *"      # first of every month at midnight
--cron "*/15 * * * *"   # every 15 minutes
```

---

### `--scheduler-run` / `-R`

```
--scheduler-run
```

Run the scheduler dispatcher as a foreground blocking daemon. Blocks until SIGINT or SIGTERM, then shuts down gracefully.

**Example:**

```bash
# As a background service
nohup mailgrid --env config.json --scheduler-run \
  --db-path /var/lib/mailgrid/jobs.db &

# systemd — see the Operational Runbook below
```

---

### `--job-retries` / `-J`

```
--job-retries <int>    default: 3
```

Number of times the scheduler retries a failed job handler before marking the job permanently failed.

**Note:** This is distinct from `--retries`, which controls per-email SMTP retries within a single job run.

---

### `--db-path`

```
--db-path <path>    default: "mailgrid.db"
```

Path to the BoltDB database file for job persistence. Created automatically on first use.

**Example:**

```bash
--db-path /var/lib/mailgrid/jobs.db
```

---

### `--jobs-list` / `--jobs-cancel`

```
--jobs-list
--jobs-cancel <job-id>
```

Manage scheduled jobs.

**Example:**

```bash
# List all jobs
mailgrid --env config.json --jobs-list

# Output:
# JOB ID                         STATUS          RUN AT               CREATED AT
# mailgrid-1718445600            pending   2025-06-15 09:00:00  2025-06-14 10:00:00

# Cancel a job
mailgrid --env config.json --jobs-cancel "mailgrid-1718445600"
```

---

## Notifications

---

### `--webhook` / `-w`

```
--webhook <url>
```

HTTP POST sent to this URL when the campaign finishes.

**Request:**
- Method: `POST`
- Content-Type: `application/json`
- Body:

```json
{
  "job_id":                "mailgrid-1718445600",
  "status":                "completed",
  "total_recipients":      2000,
  "successful_deliveries": 1988,
  "failed_deliveries":     12,
  "start_time":            "2025-06-15T10:00:00Z",
  "end_time":              "2025-06-15T10:05:47Z",
  "duration_seconds":      347,
  "concurrent_workers":    5,
  "csv_file":              "recipients.csv",
  "template_file":         "email.html"
}
```

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --webhook "https://hooks.example.com/mailgrid"
```

---

### `--webhook-secret`

```
--webhook-secret <string>
```

When set, Mailgrid adds an `X-Mailgrid-Signature: sha256=<hmac>` header to every webhook POST using HMAC-SHA256 over the request body. This lets receivers verify the request is from Mailgrid and has not been tampered with.

**Signature scheme:** identical to GitHub webhooks — `sha256=` prefix followed by the lowercase hex HMAC.

**Verifying in Python:**

```python
import hmac, hashlib

def verify(secret: str, body: bytes, header: str) -> bool:
    expected = "sha256=" + hmac.new(
        secret.encode(), body, hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, header)
```

**Verifying in Go:**

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func verify(secret, body []byte, header string) bool {
    mac := hmac.New(sha256.New, secret)
    mac.Write(body)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(header))
}
```

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --webhook "https://hooks.example.com/mailgrid" \
  --webhook-secret "$(cat /etc/mailgrid/webhook-secret)"
```

---

## Logging

All log output goes to **stderr**. CSV delivery records are written to `success.csv` / `failed.csv` separately — see [Delivery Logs](#delivery-logs).

---

### `--log-level`

```
--log-level <level>    default: "info"
```

Minimum log severity to emit.

| Level | What you see |
|---|---|
| `debug` | Every retry attempt, SMTP response, template render, connection event |
| `info` | Sent / failed per recipient, campaign summary, webhook result |
| `warn` | Soft failures (BCC RCPT error, skipped recipient) |
| `error` | Fatal failures only |

**Example:**

```bash
--log-level debug    # verbose — use for troubleshooting
--log-level warn     # quiet — use in production with a log shipper
```

---

### `--log-format`

```
--log-format <format>    default: "text"
```

Output format.

| Value | Description |
|---|---|
| `text` | Human-readable with timestamp, level, message |
| `json` | Structured JSON — one object per line, for log aggregators |

**JSON output example:**

```json
{"component":"scheduler","level":"info","msg":"Job mailgrid-123 scheduled successfully","time":"2025-06-15T10:00:00"}
{"level":"info","msg":"Sent to alice@example.com","time":"2025-06-15T10:00:01"}
{"level":"warn","msg":"Failed permanently: bob@example.com","time":"2025-06-15T10:00:02"}
```

**Example:**

```bash
# For Datadog, Loki, Splunk, Vector, etc.
mailgrid --env config.json --csv recipients.csv --template email.html \
  --log-format json --log-level info 2>> /var/log/mailgrid/campaign.log
```

---

## Offset Tracking

Mailgrid writes `.mailgrid.offset` in the working directory after every successful send during a bulk campaign. On `--resume`, this file is read and already-sent recipients are skipped — guaranteeing no duplicate sends on restart.

---

### `--resume`

```
--resume
```

Read the offset file and skip the corresponding number of recipients from the start of the task list.

**Behavior:**
- If no offset file exists, the campaign starts from the beginning.
- The offset is the count of successfully completed sends from the previous run.
- Combine with an identical `--filter` to resume a filtered campaign correctly.

**Example:**

```bash
# Original run (interrupted)
mailgrid --env config.json --csv recipients.csv --template email.html \
  --concurrency 5

# Resume
mailgrid --env config.json --csv recipients.csv --template email.html \
  --concurrency 5 --resume
```

---

### `--reset-offset`

```
--reset-offset
```

Delete the offset file and start from the beginning.

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --reset-offset
```

---

## Testing & Debug

---

### `--dry-run` / `-d`

```
--dry-run
```

Render every email to stdout without opening any SMTP connections. Attachments are listed but not read. Subject templates and HTML templates are fully evaluated.

**Use for:**
- Validating template variable references before a live send
- Counting how many recipients pass a `--filter`
- Checking subject rendering on a sample of rows

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium"' --dry-run 2>&1 | head -60

# Count matched recipients
mailgrid ... --filter '...' --dry-run 2>&1 | grep -c "Email #"
```

---

### `--preview` / `-p`

```
--preview
--port <int>    default: 8080
```

Render the first matched recipient's email and serve it on a local HTTP server. Opens immediately — no SMTP connection is made.

**Example:**

```bash
mailgrid --env config.json --csv recipients.csv --template email.html --preview
# Visit http://localhost:8080

mailgrid ... --preview --port 9000
```

---

## Advanced Patterns

### Validate before sending

```bash
# Step 1: count matches and inspect rendering
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium"' --dry-run

# Step 2: visually inspect in browser
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium"' --preview

# Step 3: send
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium"' --concurrency 5 --retries 3
```

### Resume a failed campaign

```bash
# Run with monitoring so you can see exactly where it stopped
mailgrid --env config.json --csv recipients.csv --template email.html \
  --concurrency 10 --retries 3 --monitor

# If interrupted, resume with the same arguments
mailgrid --env config.json --csv recipients.csv --template email.html \
  --concurrency 10 --retries 3 --resume
```

### Run as a systemd service

```ini
# /etc/systemd/system/mailgrid-scheduler.service
[Unit]
Description=Mailgrid Scheduler Daemon
After=network.target

[Service]
ExecStart=/usr/local/bin/mailgrid \
          --env /etc/mailgrid/config.json \
          --db-path /var/lib/mailgrid/jobs.db \
          --scheduler-run \
          --log-format json \
          --log-level info
Restart=on-failure
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now mailgrid-scheduler
```

### Avoid re-sending on retry

**Avoid:**
```bash
# Running without --resume after an interruption
mailgrid --env config.json --csv recipients.csv --template email.html
# ← Starts from the beginning, re-sends to recipients already delivered
```

**Prefer:**
```bash
# Always use --resume when restarting an interrupted campaign
mailgrid --env config.json --csv recipients.csv --template email.html --resume
# ← Picks up exactly where it left off
```

### JSON logs + monitoring together

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --concurrency 10 \
  --monitor --monitor-port 9091 \
  --log-format json --log-level info \
  2>> /var/log/mailgrid/campaign.json
```

---

## Delivery Logs

After each campaign, two CSV files are written (appended) to the working directory:

| File | Row format | Contents |
|---|---|---|
| `success.csv` | `address,subject,OK` | One row per successfully delivered email |
| `failed.csv` | `address,subject,Failed` | One row per permanent failure (retries exhausted) |

**Behavior:**
- Both files are **appended** across runs — rotate them between campaigns if per-run records are needed.
- Writes are buffered (64 KB) and flushed to disk on clean exit. A crash may lose the last buffer — check `success.csv` after a resume to identify any gap.

**Example `success.csv`:**

```
alice@example.com,Hi Alice! Your order is ready,OK
bob@example.com,Hi Bob! Your order is ready,OK
```

---

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Campaign finished (all emails sent or retried to exhaustion) |
| `1` | Fatal error — config invalid, no recipients found, attachment not readable, invalid flag value, SMTP auth failed |

**Note:** Exit code `0` does not mean every email was delivered. Permanent failures are logged and written to `failed.csv`. Check both files for a full picture of the run.

---

## Quick Reference Table

| Flag | Short | Default | Description |
|---|---|---|---|
| `--env` | `-e` | — | SMTP config JSON **(required)** |
| `--csv` | `-f` | — | CSV file path |
| `--sheet-url` | `-u` | — | Public Google Sheet URL |
| `--to` | — | — | Single recipient |
| `--template` | `-t` | — | HTML template path |
| `--text` | — | — | Plain-text body or `.txt` file |
| `--subject` | `-s` | `"Test Email from Mailgrid"` | Subject (Go template) |
| `--attach` | `-a` | — | Attachment path (repeatable) |
| `--cc` | — | — | CC addresses (comma-sep or file) |
| `--bcc` | — | — | BCC addresses (comma-sep or file) |
| `--filter` | `-F` | — | Recipient filter expression |
| `--concurrency` | `-c` | `1` | Parallel SMTP workers |
| `--batch-size` | `-b` | `1` | Emails per SMTP batch |
| `--retries` | `-r` | `1` | Per-email retry attempts |
| `--smtp-timeout` | — | `10` | SMTP dial timeout (seconds) |
| `--dry-run` | `-d` | `false` | Render without sending |
| `--preview` | `-p` | `false` | Local preview server |
| `--port` | — | `8080` | Preview server port |
| `--monitor` | `-m` | `false` | Dashboard + `/metrics` |
| `--monitor-port` | — | `9091` | Dashboard port |
| `--monitor-client-timeout` | — | `300` | Idle SSE timeout (seconds) |
| `--schedule-at` | `-A` | — | One-time schedule (RFC3339) |
| `--interval` | `-i` | — | Repeat interval (`1h`, `30m`) |
| `--cron` | `-C` | — | Cron expression (5-field) |
| `--job-retries` | `-J` | `3` | Scheduler handler retries |
| `--jobs-list` | `-L` | `false` | List scheduled jobs |
| `--jobs-cancel` | `-X` | — | Cancel job by ID |
| `--scheduler-run` | `-R` | `false` | Run scheduler as daemon |
| `--db-path` | — | `mailgrid.db` | BoltDB path |
| `--resume` | — | `false` | Resume from saved offset |
| `--reset-offset` | — | `false` | Clear offset, start fresh |
| `--webhook` | `-w` | — | Webhook URL |
| `--webhook-secret` | — | — | HMAC-SHA256 signing secret |
| `--log-level` | — | `info` | `debug` / `info` / `warn` / `error` |
| `--log-format` | — | `text` | `text` / `json` |
| `--version` | — | — | Print version and exit |
| `--help` | `-h` | — | Show help |
