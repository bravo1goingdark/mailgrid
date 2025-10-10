---
layout: default
title: Features
nav_order: 7
has_children: false
---

# Features Overview
{: .no_toc }

Comprehensive overview of MailGrid's enterprise-grade capabilities.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Email Delivery Engine

### High-Performance SMTP
- **Connection Pooling**: Reuse SMTP connections for optimal throughput
- **Concurrent Workers**: Configurable parallel processing (1-50 workers)
- **Batch Processing**: Group emails for efficient delivery
- **Rate Limiting**: Respect provider limits with intelligent throttling

### Template System
- **Go Templates**: Powerful templating with `{{.field}}` syntax
- **Dynamic Content**: Personalize subject lines and email bodies
- **Template Caching**: Lightning-fast rendering with 1-hour cache
- **HTML & Text**: Support for rich HTML and plain-text emails

### Data Sources
- **CSV Files**: Import recipients from local CSV files
- **Google Sheets**: Direct integration with public Google Sheets
- **Custom Fields**: Map any CSV column to template variables
- **Data Validation**: Automatic email format validation

---

## Monitoring & Analytics

### Real-Time Dashboard
- **Live Metrics**: Campaign progress with real-time updates
- **Recipient Tracking**: Individual email status and delivery times
- **Performance Charts**: Visual representation of sending rates
- **Error Analysis**: Detailed SMTP response code tracking

### Campaign Insights
- **Success Rates**: Track delivery and failure percentages
- **Domain Analytics**: Performance breakdown by email provider
- **Throughput Metrics**: Emails per second/minute statistics
- **Duration Tracking**: Average delivery time per email

### API Access
```bash
# Get real-time statistics
curl http://localhost:9091/api/stats

# List all recipients
curl http://localhost:9091/api/recipients
```

---

## Scheduling & Automation

### Flexible Scheduling
- **One-Time Jobs**: Schedule emails for specific times
- **Recurring Campaigns**: Cron expressions and intervals
- **Auto-Start Scheduler**: Intelligent background processing
- **Job Persistence**: BoltDB storage with crash recovery

### Cron Support
```bash
# Daily at 9 AM
--cron "0 9 * * *"

# Weekly on Monday
--cron "0 9 * * 1"

# Monthly on 1st
--cron "0 9 1 * *"

# Weekdays only
--cron "0 9 * * 1-5"
```

### Job Management
- **List Jobs**: View all scheduled campaigns
- **Cancel Jobs**: Remove scheduled jobs by ID
- **Status Monitoring**: Track job execution and failures
- **Retry Logic**: Automatic retry with exponential backoff

---

## Security & Reliability

### Built-in Security
- **Credential Protection**: Secure SMTP configuration handling
- **Input Validation**: Comprehensive data sanitization
- **Error Handling**: Graceful failure management
- **Audit Logging**: Detailed operation logs

### Resilience Features
- **Circuit Breakers**: Automatic failover during provider issues
- **Exponential Backoff**: Smart retry timing with jitter
- **Connection Recovery**: Auto-reconnect on network failures
- **Graceful Shutdown**: Clean termination with signal handling

### Enterprise Controls
- **Rate Limiting**: Configurable sending limits
- **Timeout Management**: Prevent hanging connections
- **Memory Management**: Efficient resource utilization
- **Error Classification**: Distinguish temporary vs permanent failures

---

## Developer Experience

### CLI Interface
- **Intuitive Commands**: Simple, memorable command structure
- **Short Flags**: Quick access with single-letter options
- **Help System**: Comprehensive built-in documentation
- **Exit Codes**: Proper status codes for scripting

### Integration Ready
- **Webhook Support**: HTTP notifications on completion
- **JSON Output**: Structured data for parsing
- **Environment Variables**: Flexible configuration options
- **Docker Support**: Containerized deployment ready

### Debugging Tools
- **Dry Run Mode**: Test templates without sending
- **Preview Server**: Browser-based email preview
- **Verbose Logging**: Detailed operation information
- **Error Details**: Clear error messages and suggestions

---

## Performance Specifications

### Throughput Capabilities
{: .label .label-green }
High Performance

| Metric | Specification |
|--------|---------------|
| **Maximum Concurrency** | 50 parallel workers |
| **Emails per Minute** | 10,000+ (provider dependent) |
| **Batch Size** | Configurable 1-100 emails |
| **Memory Usage** | <50MB typical |
| **Binary Size** | 3.7MB (Windows x64) |

### Scalability Features
- **Horizontal Scaling**: Multiple instances support
- **Load Distribution**: Intelligent work distribution
- **Resource Monitoring**: Built-in performance metrics
- **Auto-tuning**: Adaptive batch sizing

---

## File Format Support

### Attachment Types
- **Documents**: PDF, DOC, DOCX, TXT
- **Images**: JPG, PNG, GIF, SVG
- **Archives**: ZIP, TAR, GZ
- **Data**: CSV, JSON, XML
- **Maximum Size**: 10MB per email total

### Template Formats
- **HTML Templates**: Rich formatting with CSS
- **Plain Text**: Simple text-based emails
- **Mixed Content**: HTML with text fallback
- **Custom Headers**: Full MIME support

---

## Platform Support

### Operating Systems
- **Windows**: x64, ARM64
- **macOS**: Intel, Apple Silicon
- **Linux**: x64, ARM64
- **FreeBSD**: x64

### Distribution Methods
- **Direct Download**: Pre-built binaries
- **Package Managers**: Homebrew, Chocolatey, Scoop
- **Container Images**: Docker, Podman
- **Build from Source**: Go toolchain

---

## Compliance & Standards

### Email Standards
- **RFC 5321**: SMTP Protocol compliance
- **RFC 5322**: Internet Message Format
- **RFC 2045-2049**: MIME support
- **UTF-8 Encoding**: Full Unicode support

### Security Standards
- **TLS/SSL**: Encrypted SMTP connections
- **STARTTLS**: Opportunistic encryption
- **Authentication**: PLAIN, LOGIN, CRAM-MD5
- **SPF/DKIM Ready**: Compatible with email authentication

---

## Monitoring Integration

### Metrics Export
- **Prometheus**: Native metrics endpoint
- **JSON API**: RESTful statistics access
- **Webhooks**: Real-time notifications
- **Log Formats**: Structured JSON logging

### Health Checks
```bash
# Check scheduler health
curl http://localhost:8090/health

# Get performance metrics
curl http://localhost:8090/metrics
```

---

## Next Steps

- [Installation Guide](installation) - Get MailGrid installed
- [Getting Started](getting-started) - Your first campaign
- [CLI Reference](cli-reference) - Complete command documentation
- [Real-Time Monitoring](monitoring) - Dashboard setup and usage