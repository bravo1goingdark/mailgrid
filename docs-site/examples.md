---
layout: default
title: Examples
nav_order: 8
has_children: false
---

# Real-World Examples
{: .no_toc }

Production-ready examples for common use cases and scenarios.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Marketing Campaigns

### Newsletter Distribution

**Scenario**: Weekly newsletter to 10,000 subscribers with monitoring and webhooks.

```bash
mailgrid --env config.json \
         --csv subscribers.csv \
         --template newsletter.html \
         --subject "Weekly Tech Digest - {{.date}}" \
         --concurrency 10 \
         --batch-size 5 \
         --retries 3 \
         --monitor --monitor-port 9091 \
         --webhook "https://api.company.com/webhooks/newsletter"
```

**Template** (`newsletter.html`):
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Weekly Tech Digest</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .article { margin-bottom: 30px; border-bottom: 1px solid #eee; padding-bottom: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Weekly Tech Digest</h1>
        <p>{{.date}}</p>
    </div>

    <div class="content">
        <h2>Hello {{.first_name}}!</h2>

        <p>Welcome to your personalized tech digest. Based on your interests in <strong>{{.interests}}</strong>, here are this week's top stories:</p>

        <div class="article">
            <h3>{{.featured_title}}</h3>
            <p>{{.featured_summary}}</p>
            <a href="{{.featured_link}}">Read More →</a>
        </div>

        <p>Thanks for being a valued subscriber since {{.signup_date}}!</p>

        <p>Best regards,<br>The Tech Team</p>
    </div>
</body>
</html>
```

**CSV** (`subscribers.csv`):
```csv
email,first_name,interests,signup_date,featured_title,featured_summary,featured_link,date
john@example.com,John,AI & Machine Learning,2023-01-15,GPT-4 Breakthrough,Latest developments in AI...,https://...,December 15 2024
jane@example.com,Jane,Cloud Computing,2023-03-22,Kubernetes 1.30,New features in container orchestration...,https://...,December 15 2024
```

---

## Transactional Emails

### User Onboarding Sequence

**Scenario**: Welcome email with personalized content and attachments.

```bash
mailgrid --env config.json \
         --to "{{.user_email}}" \
         --template onboarding/welcome.html \
         --subject "Welcome to {{.company_name}}, {{.first_name}}!" \
         --attach "welcome-guide.pdf" \
         --attach "quick-start.pdf" \
         --bcc "onboarding@company.com" \
         --monitor
```

**Advanced Onboarding** with scheduling:
```bash
# Welcome email (immediate)
mailgrid --env config.json \
         --to "user@example.com" \
         --template onboarding/welcome.html \
         --subject "Welcome to TechCorp, John!" \
         --attach "welcome-guide.pdf"

# Follow-up email (3 days later)
mailgrid --env config.json \
         --to "user@example.com" \
         --template onboarding/tips.html \
         --subject "Getting the most out of TechCorp" \
         --schedule-at "2024-12-18T10:00:00Z"

# Feature highlight (1 week later)
mailgrid --env config.json \
         --to "user@example.com" \
         --template onboarding/features.html \
         --subject "Unlock powerful features" \
         --schedule-at "2024-12-22T10:00:00Z"
```

---

## Event Management

### Conference Attendee Communications

**Pre-Event Reminder**:
```bash
mailgrid --env config.json \
         --csv attendees.csv \
         --template events/reminder.html \
         --subject "TechConf 2024 starts tomorrow - {{.first_name}}!" \
         --attach "venue-map.pdf" \
         --attach "schedule.pdf" \
         --schedule-at "2024-12-19T18:00:00Z" \
         --monitor
```

**Post-Event Follow-up**:
```bash
mailgrid --env config.json \
         --csv attendees.csv \
         --template events/followup.html \
         --subject "Thank you for attending TechConf 2024!" \
         --attach "presentation-slides.zip" \
         --attach "networking-contacts.csv" \
         --webhook "https://api.events.com/webhooks/followup"
```

---

## E-commerce Scenarios

### Order Confirmations

```bash
mailgrid --env config.json \
         --to "{{.customer_email}}" \
         --template orders/confirmation.html \
         --subject "Order Confirmation #{{.order_id}}" \
         --attach "invoice-{{.order_id}}.pdf" \
         --bcc "orders@store.com" \
         --webhook "https://api.store.com/webhooks/order-sent"
```

### Abandoned Cart Recovery

**Series of 3 emails** with different timing:
```bash
# 1 hour after abandonment
mailgrid --env config.json \
         --to "customer@example.com" \
         --template cart/reminder-1h.html \
         --subject "You left something in your cart" \
         --schedule-at "2024-12-15T15:00:00Z"

# 24 hours with discount
mailgrid --env config.json \
         --to "customer@example.com" \
         --template cart/discount-24h.html \
         --subject "10% off your cart items!" \
         --schedule-at "2024-12-16T14:00:00Z"

# 3 days final reminder
mailgrid --env config.json \
         --to "customer@example.com" \
         --template cart/final-3d.html \
         --subject "Last chance - items going fast!" \
         --schedule-at "2024-12-18T14:00:00Z"
```

---

## Enterprise Communications

### Employee Newsletter

**Monthly company-wide update**:
```bash
mailgrid --env config.json \
         --csv employees.csv \
         --template internal/monthly-update.html \
         --subject "{{.company}} Monthly Update - {{.month}} {{.year}}" \
         --filter 'status = "active" and department != "contractor"' \
         --concurrency 5 \
         --monitor --monitor-port 9092 \
         --cron "0 9 1 * *"  # First day of each month at 9 AM
```

### System Maintenance Notifications

```bash
mailgrid --env config.json \
         --csv stakeholders.csv \
         --template notifications/maintenance.html \
         --subject "Scheduled Maintenance - {{.system_name}}" \
         --schedule-at "2024-12-20T17:00:00Z" \
         --cc "ops@company.com" \
         --attach "maintenance-plan.pdf"
```

---

## Customer Support

### Bulk Support Responses

**Mass response to support tickets**:
```bash
mailgrid --env config.json \
         --csv support-tickets.csv \
         --template support/update.html \
         --subject "Update on Support Ticket #{{.ticket_id}}" \
         --filter 'priority = "high" and status = "in_progress"' \
         --concurrency 3 \
         --retries 2 \
         --monitor
```

**CSV** (`support-tickets.csv`):
```csv
email,ticket_id,customer_name,issue_type,priority,status,estimated_resolution
customer1@example.com,TK-2024-001,John Smith,Login Issue,high,in_progress,2024-12-16
customer2@example.com,TK-2024-002,Jane Doe,Billing Question,medium,resolved,2024-12-15
```

---

## Educational Institutions

### Course Announcements

**Weekly course updates**:
```bash
mailgrid --env config.json \
         --csv students.csv \
         --template education/course-update.html \
         --subject "{{.course_code}}: Week {{.week_number}} Materials" \
         --filter 'enrollment_status = "active"' \
         --attach "lecture-slides-week{{.week_number}}.pdf" \
         --attach "assignment-{{.week_number}}.pdf" \
         --cron "0 8 * * 1"  # Every Monday at 8 AM
```

### Graduation Invitations

```bash
mailgrid --env config.json \
         --csv graduates.csv \
         --template events/graduation.html \
         --subject "You're invited: {{.degree_program}} Graduation Ceremony" \
         --attach "ceremony-details.pdf" \
         --attach "parking-info.pdf" \
         --cc "families@university.edu" \
         --schedule-at "2024-05-01T10:00:00Z"
```

---

## Non-Profit Organizations

### Donor Communications

**Fundraising campaign**:
```bash
mailgrid --env config.json \
         --csv donors.csv \
         --template fundraising/campaign.html \
         --subject "Help us reach our {{.campaign_goal}} goal, {{.first_name}}" \
         --filter 'last_donation_date >= "2023-01-01" and preferred_contact = "email"' \
         --attach "impact-report.pdf" \
         --webhook "https://api.nonprofit.org/webhooks/campaign-sent"
```

**Thank you series**:
```bash
mailgrid --env config.json \
         --csv recent-donors.csv \
         --template thankyou/immediate.html \
         --subject "Thank you for your generous donation!" \
         --attach "tax-receipt.pdf" \
         --bcc "development@nonprofit.org"
```

---

## Advanced Filtering Examples

### Complex Recipient Filtering

**Geographic targeting**:
```bash
mailgrid --env config.json \
         --csv customers.csv \
         --template regional/west-coast.html \
         --subject "West Coast Exclusive Event" \
         --filter 'state in ["CA", "OR", "WA"] and customer_tier = "premium"'
```

**Behavioral targeting**:
```bash
mailgrid --env config.json \
         --csv users.csv \
         --template engagement/reactivation.html \
         --subject "We miss you, {{.first_name}}!" \
         --filter 'last_login_date < "2024-11-01" and signup_date > "2024-06-01"'
```

**Purchase history targeting**:
```bash
mailgrid --env config.json \
         --csv customers.csv \
         --template offers/premium-upgrade.html \
         --subject "Exclusive upgrade offer for valued customers" \
         --filter 'total_purchases > 1000 and account_type = "standard"'
```

---

## Monitoring & Analytics Examples

### High-Volume Campaign with Full Monitoring

```bash
mailgrid --env config.json \
         --csv large-subscriber-list.csv \
         --template campaigns/black-friday.html \
         --subject "🔥 Black Friday: {{.discount_percent}}% off everything!" \
         --concurrency 20 \
         --batch-size 10 \
         --retries 3 \
         --monitor --monitor-port 9090 \
         --webhook "https://api.analytics.com/webhooks/campaign" \
         --attach "exclusive-deals.pdf"
```

**Monitor the campaign**:
```bash
# Real-time statistics
curl http://localhost:9090/api/stats | jq

# Export recipient data
curl http://localhost:9090/api/recipients > campaign-results.json
```

---

## Integration Examples

### CI/CD Pipeline Integration

**GitHub Actions workflow**:
```yaml
name: Send Release Notification
on:
  release:
    types: [published]

jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Send release notification
        run: |
          mailgrid --env config.json \
                   --csv subscribers.csv \
                   --template releases/notification.html \
                   --subject "🚀 New Release: ${{ github.event.release.tag_name }}" \
                   --webhook "${{ secrets.WEBHOOK_URL }}"
```

### Docker Compose Deployment

```yaml
version: '3.8'
services:
  mailgrid-scheduler:
    image: ghcr.io/bravo1goingdark/mailgrid:latest
    command: ["--scheduler-run", "--env", "/app/config.json"]
    volumes:
      - ./config.json:/app/config.json
      - ./templates:/app/templates
      - ./data:/app/data
    environment:
      - MAILGRID_LOG_LEVEL=info
    restart: unless-stopped

  mailgrid-monitor:
    image: ghcr.io/bravo1goingdark/mailgrid:latest
    command: ["--monitor", "--monitor-port", "9091"]
    ports:
      - "9091:9091"
    depends_on:
      - mailgrid-scheduler
```

---

## Testing & Development

### Template Testing

**Dry run with preview**:
```bash
# Test template rendering
mailgrid --env config.json \
         --csv test-data.csv \
         --template new-campaign.html \
         --subject "Test: {{.campaign_name}}" \
         --dry-run

# Preview in browser
mailgrid --preview \
         --csv test-data.csv \
         --template new-campaign.html \
         --port 8080
```

### A/B Testing Setup

**Version A**:
```bash
mailgrid --env config.json \
         --csv segment-a.csv \
         --template campaigns/version-a.html \
         --subject "Subject Line A: {{.offer}}" \
         --webhook "https://api.analytics.com/ab-test/version-a"
```

**Version B**:
```bash
mailgrid --env config.json \
         --csv segment-b.csv \
         --template campaigns/version-b.html \
         --subject "Subject Line B: {{.offer}}" \
         --webhook "https://api.analytics.com/ab-test/version-b"
```

---

## Performance Optimization

### High-Throughput Configuration

For maximum performance with reliable providers:
```bash
mailgrid --env config.json \
         --csv large-list.csv \
         --template high-volume.html \
         --subject "{{.personalized_subject}}" \
         --concurrency 25 \
         --batch-size 20 \
         --retries 2 \
         --monitor --monitor-port 9091
```

### Conservative Configuration

For consumer email providers (Gmail, Yahoo, Outlook):
```bash
mailgrid --env config.json \
         --csv subscribers.csv \
         --template newsletter.html \
         --subject "Weekly Update" \
         --concurrency 3 \
         --batch-size 1 \
         --retries 3 \
         --monitor
```

---

## Next Steps

- [CLI Reference](cli-reference) - Complete command documentation
- [Real-Time Monitoring](monitoring) - Dashboard usage and API
- [Features Overview](features) - Comprehensive capabilities guide
- [Installation Guide](installation) - Setup and configuration