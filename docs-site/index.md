---
layout: default
title: Home
nav_order: 1
description: "MailGrid - Enterprise-grade CLI for high-performance bulk email delivery"
permalink: /
---

<div class="hero-section">
  <h1 class="hero-title">MailGrid</h1>
  <p class="hero-subtitle">Enterprise-grade CLI for high-performance bulk email delivery via SMTP. Built for developers, marketers, and enterprises who demand speed, reliability, and precision.</p>

  <div class="hero-buttons">
    <a href="getting-started" class="btn btn-primary">Get Started</a>
    <a href="https://github.com/bravo1goingdark/mailgrid/releases/latest" class="btn btn-secondary">Download Latest</a>
  </div>
</div>

<div class="stats-section">
  <div class="stats-grid">
    <div class="stat-item">
      <span class="stat-number">10,000+</span>
      <span class="stat-label">Emails per minute</span>
    </div>
    <div class="stat-item">
      <span class="stat-number">99.9%</span>
      <span class="stat-label">Delivery success rate</span>
    </div>
    <div class="stat-item">
      <span class="stat-number">0</span>
      <span class="stat-label">Dependencies required</span>
    </div>
    <div class="stat-item">
      <span class="stat-number">3.7MB</span>
      <span class="stat-label">Binary size</span>
    </div>
  </div>
</div>

## Why Choose MailGrid?

<div class="feature-grid">
  <div class="feature-card">
    <span class="feature-icon">⚡</span>
    <h3 class="feature-title">Lightning Fast</h3>
    <p class="feature-description">Send thousands of emails per minute with concurrent workers and optimized SMTP connection pooling.</p>
  </div>

  <div class="feature-card">
    <span class="feature-icon">🎯</span>
    <h3 class="feature-title">Enterprise Ready</h3>
    <p class="feature-description">Production-ready scheduler, real-time monitoring, webhook integration, and comprehensive error handling.</p>
  </div>

  <div class="feature-card">
    <span class="feature-icon">🔒</span>
    <h3 class="feature-title">Secure & Reliable</h3>
    <p class="feature-description">Built-in security best practices, automatic retries, circuit breakers, and detailed audit logging.</p>
  </div>

  <div class="feature-card">
    <span class="feature-icon">🛠️</span>
    <h3 class="feature-title">Developer Friendly</h3>
    <p class="feature-description">Single binary, zero dependencies, comprehensive CLI, and seamless integration with existing workflows.</p>
  </div>

  <div class="feature-card">
    <span class="feature-icon">📊</span>
    <h3 class="feature-title">Real-time Monitoring</h3>
    <p class="feature-description">Live dashboard with campaign metrics, recipient tracking, performance analytics, and error diagnostics.</p>
  </div>

  <div class="feature-card">
    <span class="feature-icon">⏰</span>
    <h3 class="feature-title">Smart Scheduling</h3>
    <p class="feature-description">Cron-based recurring campaigns, one-time scheduling, and intelligent auto-start/shutdown capabilities.</p>
  </div>
</div>

## Quick Examples

### Single Email
```bash
mailgrid --env config.json --to user@example.com --subject "Hello" --text "Welcome!"
```

### Bulk Campaign with Monitoring
```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
         --subject "Hi {{.name}}!" --monitor --concurrency 10
```

### Scheduled Newsletter
```bash
mailgrid --env config.json --csv subscribers.csv --template newsletter.html \
         --subject "Weekly Update" --cron "0 9 * * 1" --monitor
```

## Core Capabilities

### 📧 **Email Delivery**
- **CSV & Google Sheets** - Import recipients from files or public sheets
- **Dynamic Templates** - Go template engine with custom field substitution
- **Rich Content** - HTML emails, attachments, CC/BCC support
- **Bulk Processing** - High-throughput delivery with intelligent batching

### ⚙️ **Enterprise Features**
- **SMTP Pools** - Connection pooling for optimal performance
- **Auto Retries** - Exponential backoff with jitter for failed deliveries
- **Circuit Breakers** - Automatic failover during provider issues
- **Rate Limiting** - Respect provider limits with configurable throttling

### 📊 **Monitoring & Analytics**
- **Live Dashboard** - Real-time campaign tracking and metrics
- **Recipient Status** - Individual email delivery status and timing
- **Performance Metrics** - Throughput, success rates, and error analysis
- **SMTP Diagnostics** - Response code tracking and provider insights

### ⏰ **Automation & Scheduling**
- **Cron Scheduling** - Flexible timing with 5-field cron expressions
- **Interval Campaigns** - Simple recurring sends with Go durations
- **Job Management** - List, cancel, and monitor scheduled campaigns
- **Auto-Start Scheduler** - Intelligent background processing

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