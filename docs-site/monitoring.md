---
layout: default
title: Real-Time Monitoring
nav_order: 4
---

# Real-Time Monitoring
{: .no_toc }

Track your email campaigns with MailGrid's built-in monitoring dashboard.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

MailGrid includes a real-time monitoring dashboard that provides live insights into your email campaigns. The dashboard tracks individual recipient status, campaign metrics, and performance analytics.

![Monitoring Dashboard](assets/monitoring-dashboard.png)

---

## Enabling Monitoring

Add the `--monitor` flag to any email campaign:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template newsletter.html \
         --subject "Newsletter {{.name}}" \
         --monitor
```

The dashboard will be available at `http://localhost:9091` (default port).

### Custom Port

Use a different port for the monitoring dashboard:

```bash
mailgrid --env config.json \
         --csv recipients.csv \
         --template newsletter.html \
         --monitor --monitor-port 8080
```

---

## Dashboard Features

### Campaign Overview

The dashboard header displays:
- **Job ID**: Unique identifier for the campaign
- **Start Time**: When the campaign began
- **Configuration Summary**: Key settings and file paths

### Live Statistics

Real-time metrics including:
- **Total Recipients**: Number of emails to send
- **Sent Count**: Successfully delivered emails
- **Failed Count**: Emails that couldn't be delivered
- **Emails/Second**: Current throughput rate
- **Estimated Time Left**: Projected completion time
- **Average Duration**: Mean time per email delivery

### Recipient Tracking

Individual email status for each recipient:
- **Email Address**: Recipient's email
- **Status**: Current state (pending, sending, sent, failed, retry)
- **Attempts**: Number of delivery attempts
- **Duration**: Time taken for delivery
- **Error Message**: Details for failed deliveries

### Activity Logs

Live stream of campaign events:
- Email status changes
- Delivery confirmations
- Error notifications
- Performance milestones

---

## Status Indicators

| Status | Description | Color |
|--------|-------------|-------|
| **Pending** | Queued for sending | Yellow |
| **Sending** | Currently being delivered | Blue |
| **Sent** | Successfully delivered | Green |
| **Failed** | Delivery failed permanently | Red |
| **Retry** | Retrying after temporary failure | Orange |

---

## Advanced Features

### SMTP Response Tracking

Monitor SMTP response codes to diagnose delivery issues:
- **250**: Successful delivery
- **421**: Temporary failure (retry)
- **550**: Permanent failure (bounce)
- **Other codes**: Various SMTP responses

### Domain Analytics

Track delivery performance by email provider:
- Gmail delivery rates
- Outlook/Hotmail performance
- Corporate domain statistics
- Custom domain analysis

### Performance Charts

Visual representation of:
- Delivery rate over time
- Success/failure trends
- Throughput patterns
- Error distribution

---

## Integration Examples

### High-Throughput Campaigns

Monitor large campaigns with multiple workers:

```bash
mailgrid --env config.json \
         --csv large_list.csv \
         --template campaign.html \
         --subject "Special Offer" \
         --concurrency 10 \
         --batch-size 5 \
         --monitor --monitor-port 9092
```

### Single Email Monitoring

Even single emails can be monitored:

```bash
mailgrid --env config.json \
         --to "user@example.com" \
         --subject "Test Email" \
         --text "Hello world!" \
         --monitor
```

### Combined with Webhooks

Monitor campaigns and receive completion notifications:

```bash
mailgrid --env config.json \
         --csv subscribers.csv \
         --template newsletter.html \
         --subject "Newsletter {{.name}}" \
         --monitor \
         --webhook "https://api.example.com/webhooks/mailgrid"
```

---

## Technical Details

### Real-Time Updates

The dashboard uses Server-Sent Events (SSE) for real-time updates:
- No page refresh required
- Automatic reconnection on connection loss
- Minimal bandwidth usage
- Works in all modern browsers

### Auto-Shutdown

The monitoring server automatically:
- Starts when `--monitor` is enabled
- Stops 5 seconds after campaign completion
- Handles graceful shutdown on Ctrl+C
- Cleans up resources properly

### Performance Impact

Monitoring has minimal overhead:
- Lightweight data structures
- Efficient memory management
- Non-blocking operations
- Capped log storage (1000 entries)

---

## Security Considerations

{: .warning }
> **Production Deployments**: The monitoring dashboard has no authentication by default. For production use:

### Internal Networks Only

Bind to localhost or internal networks:
```bash
# Dashboard only accessible locally
mailgrid --monitor --monitor-port 9091 ...
```

### Firewall Protection

Use firewall rules to restrict access:
```bash
# Example: Only allow specific IPs
iptables -A INPUT -p tcp --dport 9091 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 9091 -j DROP
```

### Reverse Proxy

For external access, use a reverse proxy with authentication:
```nginx
location /mailgrid-monitor/ {
    auth_basic "MailGrid Monitor";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://localhost:9091/;
}
```

---

## Troubleshooting

### Dashboard Not Loading

1. **Check the port**: Ensure the port isn't already in use
2. **Verify URL**: Visit `http://localhost:<port>` in your browser
3. **Check firewall**: Ensure the port is not blocked
4. **Browser console**: Look for JavaScript errors

### Connection Issues

1. **Server logs**: Check terminal output for errors
2. **Port conflicts**: Try a different port with `--monitor-port`
3. **Network settings**: Verify localhost access isn't restricted

### Performance Issues

1. **Reduce update frequency**: Monitoring updates every 250ms by default
2. **Limit recipients shown**: Dashboard shows up to 20 recent recipients
3. **Close unused tabs**: Multiple browser tabs can impact performance

---

## API Access

For programmatic access to monitoring data:

### Statistics Endpoint

```bash
curl http://localhost:9091/api/stats
```

Returns JSON with current campaign statistics.

### Recipients Endpoint

```bash
curl http://localhost:9091/api/recipients
```

Returns JSON array of all recipient statuses.

---

## Next Steps

- [CLI Reference](cli-reference) - Complete flag documentation
- [Scheduling](scheduling) - Automated campaigns with monitoring
- [Advanced Usage](advanced) - Webhooks and integrations
- [Performance](performance) - Optimization guidelines