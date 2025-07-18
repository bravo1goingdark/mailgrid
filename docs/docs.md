## 🏁 CLI Flags

Mailgrid offers flexible command-line options for sending personalized emails using CSV and HTML templates. Here's how it works:

---

### ⚙️ Basic Usage

```bash
mailgrid send \
  --env config.json \
  --csv contacts.csv \
  --template welcome.html \
  --subject "Welcome!" \
  --dry-run
```

---

### 📁 Available Flags

| Flag            | Shorthand | Default Value               | Description                                                                |
| --------------- | --------- | --------------------------- | -------------------------------------------------------------------------- |
| `--env`         | —         | `example/config.json`       | Path to the SMTP config JSON file (required for sending).                  |
| `--csv`         | —         | `example/test_contacts.csv` | Path to the recipient CSV file. Must include headers like `email`, `name`. |
| `--template`    | `-t`      | `example/welcome.html`      | Path to the HTML email template with Go-style placeholders.                |
| `--subject`     | `-s`      | `Test Email from Mailgrid`  | The subject line of the email. Can be overridden per run.                 |
| `--dry-run`     | —         | `false`                     | If set, renders the emails to console without sending them via SMTP.       |
| `--preview`     | `-p`       | `false`                     | Start a local server to preview the rendered email in browser.             |
| `--preview-port`| `--port`   | `8080`                      | Port for the preview server when using --preview flag.                     |

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
Each row becomes one email. First row must include headers like `email,name,company`.

---

#### `--template` / `-t`

Path to an HTML email template using Go's `text/template` syntax.
Example:

```html
<p>Hello {{.name}},</p>
<p>Welcome to {{.company}}!</p>
```

---

#### `--subject` / `-s`

Sets the subject line for the outgoing email.
Overrides the default subject.

---

#### `--dry-run`

If enabled, Mailgrid **renders emails but does not send them**.
Useful for previewing the email output and debugging templates.

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

### ✅ Example: Dry Run Mode

```bash
mailgrid send \
  --csv team.csv \
  --template onboarding.html \
  --subject "Welcome!" \
  --dry-run
```

This will print the fully rendered emails to your terminal without sending them.
