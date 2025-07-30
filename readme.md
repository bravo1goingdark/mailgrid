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

Mailgrid is a fast, minimal CLI tool for sending personalized emails from CSV files via SMTP — no web UI, just powerful automation.

---

### 📬 Email Capabilities
- **Bulk email sending** from any CSV file
- **Dynamic templating** using Go’s native `text/template`
    - Supports placeholders like `{{.name}}`, `{{.company}}`, etc.
- **Custom subject lines** via CLI flag or config

---

### ⚙️ Configuration & Control
- **SMTP support** for Gmail, Zoho, Outlook, and custom servers
- **Lightweight config** via a simple `config.json`
- **Dry-run mode** (`--dry-run`) to preview rendered emails without sending
- **Missing field warnings** for incomplete CSV rows

---

### 🛠️ Developer Experience
- **Built with Go** — fast, static binary with zero dependencies
- **Cross-platform support** — runs on Linux, macOS, and Windows
- **Live CLI logs** for each email: success ✅ or failure ❌
- **Production-ready directory structure** with modular packages

---

### 🔜 Coming Soon
- 🚦 rate-limiting
- 📊 Delivery summary metrics (sent, failed, skipped)

> 📄 Licensed under BSD-3-Clause — see [LICENSE](./LICENSE)





