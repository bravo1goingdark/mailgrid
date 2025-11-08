## 🏁 CLI Flags

Mailgrid now supports high-throughput dispatch and automatic retry handling.  
Below is the complete, production-ready flag reference with **`--concurrency`** and **`--retries`** added.

---

### ⚙️ Basic Usage — Production Sends

```bash
mailgrid \
  --env cfg/prod.json \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Welcome!" \
  --concurrency 5 \
  --retries 3
```

### 📁 Available Flags

| Flag               | Shorthand | Default Value              | Description                                                                                 |
|--------------------|-----------|----------------------------|---------------------------------------------------------------------------------------------|
| `--env`            | —         | `""`                       | Path to the SMTP config JSON file (required for sending).                                   |
| `--csv`            | —         | `""`                       | Path to the recipient CSV file. Must include headers like `email`, `name`.                  |
| `--sheet-url`      | —         | `""`                       | Google Sheet CSV URL as an alternative to local `--csv` file.                               |
| `--template`       | `-t`      | `example/welcome.html`     | Path to the HTML email template with Go-style placeholders.                                 |
| `--subject`        | `-s`      | `Test Email from Mailgrid` | The subject line of the email. Can be overridden per run.                                   |
| `--cc`             | —         | `""`                       | Comma-separated list or file (`@file.txt`) of CC email addresses (visible recipients).      |
| `--bcc`            | —         | `""`                       | Comma-separated list or file (`@file.txt`) of BCC addresses (hidden from recipients).       |
| `--to`             | -         | `""`                       | The email address of the single recipient. Cannot be used with --csv.                       |
| `--text`           | -         | `""`                       | Inline plain-text body or path to a .txt file. Cannot be used with --template.              |
| `--dry-run`        | —         | `false`                    | If set, renders the emails to console without sending them via SMTP.                        |
| `--preview`        | `-p`      | `false`                    | Start a local server to preview the rendered email in browser.                              |
| `--port`           | `--port`  | `8080`                     | Port for the preview server when using `--preview` flag.                                    |
| `--concurrency`    | `-c`      | `1`                        | Number of parallel worker goroutines that send emails concurrently.                         |
| `--retries`        | `-r`      | `1`                        | Maximum retry attempts per email on transient errors (exponential backoff).                 |
| `--batch-size`     | —         | `1`                        | Number of emails to send per SMTP connection (helps avoid throttling).                      |
| `--filter`         | —         | `""`                       | Filter rows using a conditional expression (e.g. `tier = "pro" and age > 25`).              |
| `--attach`         | -         | `[]`                       | File attachments to include with every email. Repeat flag for multiple files. (MAX = 10MB)  |
| `--webhook`        | —         | `""`                       | HTTP URL to send POST request with campaign results after completion.                        |
| `--monitor`        | —         | `false`                    | Enable real-time monitoring dashboard to track email sending progress.                       |
| `--monitor-port`   | —         | `9091`                     | Port for the real-time monitoring dashboard server.                                          |
| `--schedule-at`    | `-A`      | `""`                       | Schedule send at an RFC3339 time (e.g. `2025-09-08T09:00:00Z`).                             |
| `--interval`       | `-i`      | `""`                       | Recurring schedule using Go duration (e.g. `1h`, `30m`).                                    |
| `--cron`           | `-C`      | `""`                       | Recurring schedule using 5-field cron (minute hour dom month dow).                          |
| `--job-retries`    | `-J`      | `3`                        | Scheduler-level max attempts on handler failure (separate from SMTP `--retries`).           |
| `--job-backoff`    | `-B`      | `2s`                       | Base backoff duration for scheduler retries (exponential with jitter, capped at 5m).        |
| `--jobs-list`      | `-L`      | `false`                    | List scheduled jobs in the scheduler database.                                              |
| `--jobs-cancel`    | `-X`      | `""`                       | Cancel job by ID.                                                                           |
| `--scheduler-run`  | `-R`      | `false`                    | Run the scheduler dispatcher in the foreground (press Ctrl+C to stop).                      |
| `--scheduler-db`   | `-D`      | `mailgrid.db`              | Path to BoltDB for schedules. Default is `mailgrid.db` in current working directory.        |

---

### 📌 Flag Descriptions

#### `--env`

Path to a required SMTP config file in JSON format:

```json
{
  "host": "smtp.zoho.com",
  "port": 587,
  "username": "you@example.com",
  "password": "your_smtp_password",
  "from": "you@example.com"
}
```

---

#### `--csv`

Path to the `.csv` file containing recipients.

- **Required column:** `email` (case-insensitive).
- Optional columns (e.g. `name`, `company`) can be referenced from the template.

Each row becomes one email.

---

---

#### `--sheet-url`

Fetch recipients from a **public Google Sheet** instead of a local CSV.

- **Required column:** `email` (case-insensitive).
- Optional columns (e.g. `name`, `company`) can be used in the email template.
- Each row becomes one email.
- ⚠️ Currently works **only for public Google Sheets** (set to "Anyone with the link can view").

**Example:**

```bash
mailgrid --env example/config.json \
  --sheet-url "https://docs.google.com/spreadsheets/d/1EUh5VWlSNtrlEIJ6SjJAQ9kYAcf4XrlsIIwXtYjImKc/edit?gid=1980978683#gid=1980978683" \
  -t example/welcome.html \
  -s "Welcome {{.name}}" \
  -c 5 \
  --batch-size 5
```

---

#### `--template` / `-t`

Path to an HTML (or plain-text) email template rendered with Go’s `text/template`.

**Interpolation**

- Use `{{ .ColumnName }}` to inject values from each CSV row—e.g. `{{ .email }}`, `{{ .name }}`, `{{ .company }}`.

Example:

```html
<p>Hello {{ .name }},</p>
<p>Welcome to {{ .company }}!</p>
```

---

#### `--subject` / `-s`

Define the **subject line** for each outgoing email.

- Accepts **plain text** or Go `text/template` placeholders—e.g. `Welcome, {{ .name }}!`.
- Overrides the default subject (`Test Email from Mailgrid`) if one isn’t already set.
- Placeholders are resolved with the same CSV columns available to your template.

Example:

```bash
mailgrid \
  --subject "Monthly update for {{ .company }}" \
  --csv contacts.csv \
  --template newsletter.html
```
---
#### `--cc`

Define one or more CC (carbon copy) recipients for the outgoing email.

- These addresses will appear in the Cc: header and be visible to all recipients.
- Accepts a comma-separated string or a file reference using the @ symbol.
- Useful when you want to transparently include teammates, managers, or collaborators.

Example:

```bash
mailgrid \
  --cc "team@example.com,manager@example.com" \
  --csv contacts.csv \
  --template newsletter.html
```
---
#### `--bcc`

Define one or more BCC (blind carbon copy) recipients for each email.

- These addresses receive the email silently—they don’t appear in the To: or Cc: headers.
- Accepts a comma-separated string or a file reference with @.
- Great for logging, supervisors, or invisible monitoring.

Example:

```bash
mailgrid \
  --bcc "admin@example.com" \
  --csv contacts.csv \
  --template newsletter.html
```
---
### `--to`

Used to send an email to a single recipient without a CSV or Google Sheet.
This flag is mutually exclusive with --csv and --sheet-url.

Example:
```bash
--to test@example.com
```
Useful for sending quick one-off messages without uploading recipient lists.

---
### `--text`
Provides a plain-text body for the email, either inline or via a .txt file path.
This flag is mutually exclusive with --template.

Example:
```bash
# Inline text
--text "This is a test email body"

# OR from a file
--text ./body.txt
```
Ideal for simple messages or debugging without using HTML templates.

---

#### `--dry-run`

If enabled, Mailgrid **renders the emails but does not send them via SMTP**.

- Print the fully rendered output for each recipient to the console.
- Helpful for **debugging templates**, verifying CSV mapping, and checking final email content before a live sending.
- Can be combined with `--concurrency` to speed up rendering.

Example:

```bash
mailgrid \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  --dry-run
```

---

### 📬 Email Preview Server

```bash
# Preview using default example CSV and HTML template
mailgrid --preview

# Shorthand flag with defaults
mailgrid -p

# Provide custom CSV and HTML template
mailgrid --preview --csv example/test_contacts.csv --template example/welcome.html

# Shorthand with custom port
mailgrid -p --port 7070 --csv data/contacts.csv --template templates/offer.html



```

The preview server can be stopped by pressing Ctrl+C in your terminal.

---

#### `--concurrency` / `-c`

Set the number of parallel SMTP workers to use when sending emails.

- Each worker maintains a **persistent SMTP connection**.
- Improves speed by sending multiple emails at once.
- 🛑 **Recommended: Keep ≤ 5** unless you're confident about your SMTP provider's rate limits.
- 📤 **Outputs:**
    - `success.csv`: all emails sent successfully
    - `failed.csv`: emails that failed after all retries

**Example:**

```bash
mailgrid \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  --concurrency 5
```

or using shorthand:

```bash
mailgrid \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  -c 5
```

---

#### `--retries` / `-r`

Set how many times a failed email will be retried before being marked as a failure.

- Mailgrid performs **1 retry by default**, so each message gets the initial send plus one automatic follow-up attempt. Increase the limit with `--retries <n>` (or `-r <n>`) when you expect transient issues.
- Set `--retries 0` if you want to disable automatic retries entirely.

- Retries are spaced using **exponential backoff**:
  Delay = `2^n seconds` between each retry attempt.
- A small **jitter (random delay)** is added to each retry to avoid **thundering herd** problems when multiple failures
  occur at once.
- `total delay = 2^n + rand(0,1)`

#### \* Retries help recover from:

- 🔌 Temporary network drops
- 🧱 SMTP 4xx soft errors (e.g. greylisting)
- 🕒 Provider-imposed rate limits or slow responses

### ⚠️ Best Practices

- Use `--retries 2` or `3` for most production scenarios
- Use alongside `--concurrency` and `--dry-run` for safe testing and debugging-
- 🚫 Avoid exceeding `3` retries unless you're handling high-stakes or critical messages

Example:

```bash
mailgrid \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  --retries 3
```

or using shorthand:

```bash
mailgrid \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  -r 3
```

---

#### `--batch-size`

Controls how many emails are grouped and sent together in one flush by each worker.

A higher batch size reduces SMTP overhead and improves throughput, especially for bulk sends to **enterprise or
transactional mail providers**.  
However, it comes with trade-offs depending on the target inbox provider.

---

### 🚫 When Not to Use Large Batch Sizes

Avoid large batch sizes when targeting **consumer inboxes** like:

- 📬 Gmail
- 📬 Yahoo
- 📬 Outlook/Hotmail

These providers:

- Enforce **aggressive rate limits**
- Detect batched emails as potential **spam bursts**
- May delay, throttle, or **block SMTP sessions** that deliver too many messages in one shot

### ⚠️ Best Practices

- For Gmail/Yahoo/Outlook: use `--batch-size 1` <- **default**
- For trusted corporate domains or warmed-up IPs: `--batch-size 5–10`
- Always test with `--dry-run` before scaling batch sizes

---

### 💡 Tip

Each batch is flushed per worker.  
So with `--concurrency 4` and `--batch-size 5`, up to **20 emails** can be processed and sent in parallel.

---

### `--filter`

- You can filter rows before sending emails using the `--filter` flag.
- Want advanced filters like `contains`, `!=`, or grouped conditions?
    - 👉 See [Filter Documentation](filter.md) for full syntax and supported operators.
- For instance, to only email users who are **Pro tier** and **older than 25**:

```bash
mailgrid \
  --env config.json \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Welcome!" \
  --filter 'tier = "pro" and age > 25' \
  --concurrency 5
```

---
#### `--attach`

- Include one or more file attachments with every email you send. Provide the flag multiple times for multiple files, e.g. `--attach brochure.pdf --attach terms.pdf`.
- Max of 10 MB is allowed collectively.

Example:

```bash
mailgrid \
  --csv contacts.csv \
  --template invoice.html \
  --attach invoice.pdf \
  --attach receipt.pdf
```

---

#### `--webhook`

Send HTTP POST notifications with campaign results after email completion.

- **URL validation**: Only HTTP and HTTPS URLs are accepted
- **Automatic notifications**: Sent after both bulk and single email campaigns
- **Rich payload**: Includes metrics like total recipients, success/failure counts, duration, and file paths
- **Non-blocking**: Webhook delivery runs asynchronously and won't delay email sending

The webhook payload structure:

```json
{
  "job_id": "mailgrid-1633024800",
  "status": "completed",
  "total_recipients": 150,
  "successful_deliveries": 148,
  "failed_deliveries": 2,
  "start_time": "2023-10-01T10:00:00Z",
  "end_time": "2023-10-01T10:05:30Z",
  "duration_seconds": 330,
  "concurrent_workers": 5,
  "csv_file": "subscribers.csv",
  "template_file": "newsletter.html"
}
```

**Examples:**

```bash
# Webhook with bulk email campaign
mailgrid --env config.json \
  --csv subscribers.csv \
  --template newsletter.html \
  --subject "Newsletter {{.name}}" \
  --webhook "https://api.example.com/webhooks/mailgrid"

# Webhook with single email
mailgrid --env config.json \
  --to "user@example.com" \
  --subject "Welcome" \
  --text "Thanks for signing up!" \
  --webhook "https://api.example.com/webhooks/mailgrid"
```

---

#### `--monitor` & `--monitor-port`

Enable real-time monitoring dashboard to track email sending progress with live metrics and recipient status.

**Key Features:**

- **Live Campaign Stats**: Total recipients, sent/failed counts, throughput (emails/sec), estimated time remaining
- **Real-time Recipient Tracking**: Individual email status, retry attempts, duration, error messages
- **SMTP Response Monitoring**: Track response codes (250, 421, 550, etc.) for debugging
- **Domain Analytics**: Breakdown by email provider (Gmail, Outlook, etc.)
- **Live Log Stream**: Real-time activity feed of send events
- **Progress Visualization**: Progress bars and status indicators

**Dashboard Interface:**

The monitoring dashboard provides a modern, responsive web interface accessible at `http://localhost:<port>` (default port 9091). The dashboard includes:

- **Campaign Overview**: Job ID, start time, configuration summary
- **Statistics Grid**: Key metrics with progress visualization
- **Recipients Table**: Live status updates for up to 20 recent recipients
- **Activity Logs**: Rolling log of the latest send events

**Examples:**

```bash
# Enable monitoring with default port (9091)
mailgrid --env config.json \
  --csv subscribers.csv \
  --template newsletter.html \
  --subject "Newsletter {{.name}}" \
  --monitor

# Use custom port for monitoring dashboard
mailgrid --env config.json \
  --to "user@example.com" \
  --subject "Test Email" \
  --text "Hello world!" \
  --monitor --monitor-port 8080

# Monitor high-throughput campaign
mailgrid --env config.json \
  --csv large_list.csv \
  --template campaign.html \
  --subject "Special Offer" \
  --concurrency 10 \
  --monitor --monitor-port 9092
```

**Notes:**
- Dashboard automatically starts when `--monitor` is enabled
- Server stops automatically 5 seconds after email sending completes
- Works with both bulk campaigns and single email sending
- Real-time updates via Server-Sent Events (no page refresh needed)
- Compatible with all other CLI flags and features

---

### 🧪 Example

```bash
mailgrid \
  --csv contacts.csv \
  --template invite.html \
  --subject "You're Invited!" \
  --batch-size 1 \
  --concurrency 4 \
  --retries 3 \
  --batch-size 5 \
  --filter 'name = ashutosh && email contains @gmail.com' \
  --attach brochure.pdf
```

---

## ⏱️ Scheduling and Job Management

You can schedule one-off or recurring sends. Schedules are persisted in a local BoltDB file (default: `mailgrid.db` in your current working directory). Use listing/cancel commands to manage jobs, and optionally run the dispatcher in the foreground.

Short forms: -A (schedule-at), -i (interval), -C (cron), -J (job-retries), -B (job-backoff), -L (jobs-list), -X (jobs-cancel), -R (scheduler-run), -D (scheduler-db)

### Short-form examples

- One-off at specific time:
```bash
mailgrid -A 2025-09-08T09:00:00Z --env example/config.json --csv example/test_contacts.csv -t example/welcome.html -s "Welcome {{.name}}"
```
- Every 2 minutes:
```bash
mailgrid -i 2m --env example/config.json --csv example/test_contacts.csv -t example/welcome.html -s "Welcome {{.name}}"
```
- Cron daily 09:00:
```bash
mailgrid -C "0 9 * * *" --env example/config.json --csv example/test_contacts.csv -t example/welcome.html -s "Morning {{.name}}"
```
- List / cancel / run scheduler / custom DB:
```bash
mailgrid -L
mailgrid -X <job_id>
mailgrid -R -D mailgrid.db
```

- One-off scheduled CSV send (RFC3339 time):

```bash
mailgrid \
  --env example/config.json \
  --csv example/test_contacts.csv \
  --template example/welcome.html \
  --subject "Welcome {{.name}}" \
  --schedule-at 2025-09-08T09:00:00Z
```

- Recurring by interval:

```bash
mailgrid \
  --env example/config.json \
  --csv example/test_contacts.csv \
  --template example/welcome.html \
  --subject "Welcome {{.name}}" \
  --interval 1h
```

- Recurring by cron (every day at 09:00):

```bash
mailgrid \
  --env example/config.json \
  --csv example/test_contacts.csv \
  --template example/welcome.html \
  --subject "Welcome {{.name}}" \
  --cron "0 9 * * *"
```

- Scheduler database path (optional):

```bash
# Uses mailgrid.db by default; override when needed
--scheduler-db custom-schedules.db
```

- Scheduler-level retry/backoff (separate from SMTP `--retries`):

```bash
--job-retries 3 --job-backoff 2s
```

- List and cancel jobs:

```bash
mailgrid --jobs-list
mailgrid --jobs-cancel <job_id>
```

- Run scheduler dispatcher in the foreground (reattaches handlers and processes due jobs):

```bash
mailgrid --scheduler-run
# Press Ctrl+C to stop
```
