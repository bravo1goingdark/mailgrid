## 🏁 CLI Flags

Mailgrid now supports high-throughput dispatch and automatic retry handling.  
Below is the complete, production-ready flag reference with **`--concurrency`** and **`--retry-limit`** added.

---

### ⚙️ Basic Usage — Production Sends

```bash
mailgrid send \
  --env cfg/prod.json \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Welcome!" \
  --concurrency 5 \
  --retry-limit 3
  
```

### 📁 Available Flags

| Flag             | Shorthand | Default Value               | Description                                                                 |
|------------------|-----------|-----------------------------|-----------------------------------------------------------------------------|
| `--env`          | —         | `example/config.json`       | Path to the SMTP config JSON file (required for sending).                   |
| `--csv`          | —         | `example/test_contacts.csv` | Path to the recipient CSV file. Must include headers like `email`, `name`.  |
| `--template`     | `-t`      | `example/welcome.html`      | Path to the HTML email template with Go-style placeholders.                 |
| `--subject`      | `-s`      | `Test Email from Mailgrid`  | The subject line of the email. Can be overridden per run.                   |
| `--dry-run`      | —         | `false`                     | If set, renders the emails to console without sending them via SMTP.        |
| `--preview`      | `-p`      | `false`                     | Start a local server to preview the rendered email in browser.              |
| `--preview-port` | `--port`  | `8080`                      | Port for the preview server when using `--preview` flag.                    |
| `--concurrency`  | `-c`      | `1`                         | Number of parallel worker goroutines that send emails concurrently.         |
| `--retries`      | `-r`      | `2`                         | Maximum retry attempts per email on transient errors (exponential backoff). |
| `--batch-size`   | _         | `1`                         | Number of emails per SMTP batch                                             |

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

* **Required column:** `email` (case-insensitive).
* Optional columns (e.g. `name`, `company`) can be referenced from the template.

Each row becomes one email.

---

#### `--template` / `-t`

Path to an HTML (or plain-text) email template rendered with Go’s `text/template`.

**Interpolation**

* Use `{{ .ColumnName }}` to inject values from each CSV row—e.g. `{{ .email }}`, `{{ .name }}`, `{{ .company }}`.

Example:

```html
<p>Hello {{ .name }},</p>
<p>Welcome to {{ .company }}!</p>
```

---

#### `--subject` / `-s`

Define the **subject line** for each outgoing email.

* Accepts **plain text** or Go `text/template` placeholders—e.g. `Welcome, {{ .name }}!`.
* Overrides the default subject (`Test Email from Mailgrid`) if one isn’t already set.
* Placeholders are resolved with the same CSV columns available to your template.

Example:

```bash
mailgrid send \
  --subject "Monthly update for {{ .company }}" \
  --csv contacts.csv \
  --template newsletter.html
```

---

#### `--dry-run`

If enabled, Mailgrid **renders the emails but does not send them via SMTP**.

* Print the fully rendered output for each recipient to the console.
* Helpful for **debugging templates**, verifying CSV mapping, and checking final email content before a live sending.
* Can be combined with `--concurrency` to speed up rendering.

Example:

```bash
mailgrid send \
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

* Each worker maintains a **persistent SMTP connection**.
* Improves speed by sending multiple emails at once.
* 🛑 **Recommended: Keep ≤ 5** unless you're confident about your SMTP provider's rate limits.

Example:

```bash
mailgrid send \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  --concurrency 5 
```

or using shorthand:

```bash
mailgrid send \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  -c 5
```

---
#### `--retries` / `-r`

Set how many times a failed email will be retried before being marked as a failure.

* Retries are spaced using **exponential backoff**:  
Delay = `2^n seconds` between each retry attempt.
* A small **jitter (random delay)** is added to each retry to avoid **thundering herd** problems when multiple failures occur at once.
* `total delay = 2^n + rand(0,1)`

#### * Retries help recover from:
- 🔌 Temporary network drops
- 🧱 SMTP 4xx soft errors (e.g. greylisting)
- 🕒 Provider-imposed rate limits or slow responses

### ⚠️ Best Practices

-  Use `--retries 2` or `3` for most production scenarios
- Use alongside `--concurrency` and `--dry-run` for safe testing and debugging- 
- 🚫 Avoid exceeding `3` retries unless you're handling high-stakes or critical messages

Example:

```bash
mailgrid send \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  --retries 3
```

or using shorthand:

```bash
mailgrid send \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  -r 3
```

---

#### `--batch-size`

Controls how many emails are grouped and sent together in one flush by each worker.

A higher batch size reduces SMTP overhead and improves throughput, especially for bulk sends to **enterprise or transactional mail providers**.  
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

---

### ⚠️ Best Practices

- For Gmail/Yahoo/Outlook: use `--batch-size 1` <- **default**
- For trusted corporate domains or warmed-up IPs: `--batch-size 5–10`
- Always test with `--dry-run` before scaling batch sizes

---

### 💡 Tip

Each batch is flushed per worker.  
So with `--concurrency 4` and `--batch-size 5`, up to **20 emails** can be processed and sent in parallel.

---

### 🧪 Example

```bash
mailgrid send \
  --csv contacts.csv \
  --template invite.html \
  --subject "You're Invited!" \
  --batch-size 1 \
  --concurrency 4 \
  --retries 3 \
  --batch-size 5
  
```