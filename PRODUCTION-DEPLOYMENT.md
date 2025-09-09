# Production Deployment Guide

This guide covers deploying Mailgrid in production environments with high performance, security, and reliability.

## 🚀 Quick Start

### 1. Build Production Binary

```bash
# Build optimized binary
make build

# Or build for all platforms
make build-all
```

### 2. Create Production Configuration

```bash
# Generate example production config
make example-config

# Edit the generated config
nano example/config-production.json
```

Example production configuration:

```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com",
    "use_tls": true,
    "insecure_skip_verify": false,
    "connection_timeout": "10s",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "rate_limit": 100,
  "burst_limit": 200,
  "log": {
    "level": "info",
    "format": "json",
    "file": "/var/log/mailgrid/mailgrid.log",
    "max_size": 100,
    "max_backups": 3,
    "max_age": 28
  },
  "metrics": {
    "enabled": true,
    "port": 8090
  },
  "max_attachment_size": 25165824,
  "max_concurrency": 50,
  "max_batch_size": 100,
  "max_retries": 5
}
```

### 3. Deploy with Docker

```bash
# Build Docker image
docker build -t mailgrid:production .

# Run with production config
docker run -d \
  --name mailgrid \
  -v $(pwd)/config-production.json:/app/config.json \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -p 8090:8090 \
  mailgrid:production
```

## 📊 Monitoring & Observability

### Health Checks

Mailgrid provides health and readiness endpoints:

- **Health**: `GET /health` - Basic health check
- **Readiness**: `GET /ready` - Checks if workers are active
- **Metrics**: `GET /metrics` - Prometheus-compatible metrics

### Metrics Available

- `emails_sent_total` - Total emails sent successfully
- `emails_failed_total` - Total emails that failed permanently
- `emails_retried_total` - Total email retry attempts
- `smtp_connections_active` - Active SMTP connections
- `workers_active` - Number of active workers
- `jobs_scheduled_total` - Total scheduled jobs
- `jobs_completed_total` - Total completed jobs
- `jobs_failed_total` - Total failed jobs
- `response_times_ms` - Response times by operation
- `error_counts` - Error counts by type
- `uptime_seconds` - Application uptime

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'mailgrid'
    static_configs:
      - targets: ['mailgrid:8090']
    scrape_interval: 30s
    metrics_path: '/metrics'
```

### Grafana Dashboard

Key metrics to monitor:

1. **Email Throughput**: Rate of emails sent per minute
2. **Success Rate**: Percentage of successful vs failed emails
3. **SMTP Connection Health**: Active connections and errors
4. **Worker Utilization**: Number of active workers
5. **Response Times**: P50, P95, P99 latencies
6. **Error Rates**: Error types and frequencies

## ⚡ Performance Tuning

### High-Volume Configuration

For sending millions of emails:

```json
{
  "smtp": {
    "connection_timeout": "5s",
    "read_timeout": "15s",
    "write_timeout": "15s"
  },
  "rate_limit": 500,
  "burst_limit": 1000,
  "max_concurrency": 100,
  "max_batch_size": 200,
  "max_retries": 3
}
```

### Resource Limits

**Recommended system resources:**

- **CPU**: 2-4 cores for moderate loads (10K-100K emails/hour)
- **Memory**: 2-4GB RAM for large recipient lists
- **Disk**: SSD recommended for database and logs
- **Network**: Stable connection with sufficient bandwidth

### Database Optimization

For scheduler persistence with BoltDB:

- Use SSD storage for better I/O performance
- Regular database compaction
- Monitor database size growth

```bash
# Check database size
ls -lh mailgrid.db

# Backup database
cp mailgrid.db mailgrid.db.backup.$(date +%Y%m%d)
```

## 🔐 Security Best Practices

### Configuration Security

1. **Environment Variables**: Store sensitive data in environment variables

```bash
export SMTP_PASSWORD="your-secure-password"
export SMTP_USERNAME="your-username"
```

2. **File Permissions**: Restrict config file access

```bash
chmod 600 config-production.json
chown mailgrid:mailgrid config-production.json
```

3. **TLS Configuration**: Always use TLS for SMTP

```json
{
  "smtp": {
    "use_tls": true,
    "insecure_skip_verify": false
  }
}
```

### Network Security

- Deploy behind a firewall
- Use VPN for remote access
- Restrict metrics endpoint access
- Monitor for unusual traffic patterns

### Container Security

```dockerfile
# Run as non-root user
USER mailgrid

# Read-only filesystem
--read-only
--tmpfs /tmp

# Resource limits
--memory=2g
--cpus=2.0
```

## 📋 Operational Procedures

### Deployment Checklist

- [ ] Configuration validated
- [ ] SMTP credentials tested
- [ ] Health endpoints accessible
- [ ] Metrics collection configured
- [ ] Log rotation configured
- [ ] Backup procedures in place
- [ ] Monitoring alerts configured

### Backup Strategy

1. **Configuration**: Version control + regular backups
2. **Database**: Daily backups of scheduler database
3. **Logs**: Centralized log aggregation
4. **Metrics**: Historical metrics storage

### Disaster Recovery

1. **Database Corruption**:
   ```bash
   # Restore from backup
   cp mailgrid.db.backup.latest mailgrid.db
   ```

2. **Configuration Loss**:
   - Restore from version control
   - Verify SMTP settings

3. **Service Recovery**:
   ```bash
   # Check service status
   systemctl status mailgrid
   
   # Restart service
   systemctl restart mailgrid
   
   # Check logs
   journalctl -u mailgrid -f
   ```

## 🔄 Scaling Strategies

### Horizontal Scaling

Deploy multiple instances with:

- Shared database (with distributed locking)
- Load balancer for job distribution
- Centralized metrics collection

### Vertical Scaling

Optimize single instance:

- Increase worker concurrency
- Larger batch sizes
- More memory for caching
- Faster storage (NVMe SSD)

### Queue-Based Architecture

For extremely high volumes, consider:

- External job queue (Redis, RabbitMQ)
- Separate scheduler and worker processes
- Auto-scaling based on queue depth

## 🚨 Troubleshooting

### Common Issues

1. **SMTP Authentication Failures**
   - Check credentials and app passwords
   - Verify TLS/SSL settings
   - Check firewall rules

2. **High Memory Usage**
   - Reduce batch sizes
   - Lower concurrency
   - Check for email content size

3. **Database Lock Contention**
   - Reduce scheduler polling frequency
   - Check for long-running jobs
   - Monitor database file size

4. **Rate Limiting Issues**
   - Adjust rate limits for your SMTP provider
   - Monitor SMTP provider quotas
   - Implement exponential backoff

### Debug Commands

```bash
# Check build info
./mailgrid --version

# Test configuration
./mailgrid --env config-production.json --dry-run

# List scheduled jobs
./mailgrid -L -D mailgrid.db

# Monitor real-time metrics
curl http://localhost:8090/metrics

# Check health status
curl http://localhost:8090/health
```

### Log Analysis

Key log patterns to monitor:

```bash
# Failed emails
grep "Failed permanently" mailgrid.log

# SMTP connection issues
grep "SMTP.*error" mailgrid.log

# Rate limiting
grep "Rate limiter" mailgrid.log

# Worker activity
grep "Worker.*Starting\|Stopping" mailgrid.log
```

## 📚 Additional Resources

- [Configuration Reference](docs/configuration.md)
- [API Documentation](docs/api.md)
- [Performance Benchmarks](docs/benchmarks.md)
- [Security Guidelines](docs/security.md)

For support and questions, please check the [GitHub Issues](https://github.com/bravo1goingdark/mailgrid/issues).
