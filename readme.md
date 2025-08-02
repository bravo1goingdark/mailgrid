<p align="center">
  <img src="./assets/readme-banner-mailgrid.svg" alt="Mailgrid Logo" width="100%" height="100%"/>
</p>

<p align="center">
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

### 🔜 Coming Soon
- 🚦 rate-limiting
- 📊 Delivery summary metrics (sent, failed, skipped)

> 📄 Licensed under BSD-3-Clause — see [LICENSE](./LICENSE)





