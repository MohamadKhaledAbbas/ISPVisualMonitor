# ISP Visual Monitor - Operations Guide

This guide covers day-to-day operations, monitoring, maintenance, and troubleshooting for ISP Visual Monitor deployments.

## Table of Contents

- [Monitoring](#monitoring)
- [Alerting](#alerting)
- [Backup and Restore](#backup-and-restore)
- [Scaling](#scaling)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

## Monitoring

### Prometheus Metrics

ISP Visual Monitor exposes Prometheus metrics on port 9090 by default.

#### Key Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ispmonitor_http_requests_total` | Counter | Total HTTP requests by method, endpoint, status |
| `ispmonitor_http_request_duration_seconds` | Histogram | HTTP request latency |
| `ispmonitor_routers_total` | Gauge | Total routers by status |
| `ispmonitor_active_pollers` | Gauge | Number of active polling workers |
| `ispmonitor_routers_polled_total` | Counter | Total routers polled |
| `ispmonitor_poll_duration_seconds` | Histogram | Polling duration |
| `ispmonitor_alerts_active` | Gauge | Active alerts by severity |
| `ispmonitor_license_valid` | Gauge | License validity (1=valid, 0=invalid) |

#### Example Prometheus Queries

```promql
# Request rate per second
rate(ispmonitor_http_requests_total[5m])

# 99th percentile latency
histogram_quantile(0.99, rate(ispmonitor_http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(ispmonitor_http_requests_total{status=~"5.."}[5m])) / sum(rate(ispmonitor_http_requests_total[5m]))

# Routers offline
ispmonitor_routers_total{status="offline"}

# Polling throughput
rate(ispmonitor_routers_polled_total[5m])
```

### Grafana Dashboards

Pre-built dashboards are included:

1. **ISP Monitor Overview**: High-level metrics and status
2. **Router Health**: Individual router metrics
3. **API Performance**: HTTP request metrics
4. **Alerts**: Alert status and history

#### Importing Dashboards

Dashboards are automatically provisioned when using Docker Compose. For manual import:

```bash
# Copy dashboard JSON files
cp deploy/grafana/dashboards/*.json /var/lib/grafana/dashboards/
```

### Health Endpoints

| Endpoint | Purpose | Expected Response |
|----------|---------|-------------------|
| `GET /health` | Full health check | 200 OK with JSON status |
| `GET /ready` | Readiness probe | 200 OK when ready |
| `GET /live` | Liveness probe | 200 OK when alive |

#### Health Check Script

```bash
#!/bin/bash
# health-check.sh

HEALTH_URL="${1:-http://localhost:8080/health}"

response=$(curl -s -w "\n%{http_code}" "$HEALTH_URL")
body=$(echo "$response" | head -n -1)
status=$(echo "$response" | tail -n 1)

if [ "$status" -eq 200 ]; then
    echo "✓ Healthy"
    echo "$body" | jq .
    exit 0
else
    echo "✗ Unhealthy (HTTP $status)"
    echo "$body"
    exit 1
fi
```

## Alerting

### Prometheus Alerting Rules

Create alert rules in Prometheus:

```yaml
# alerting-rules.yml
groups:
  - name: ispmonitor
    rules:
      - alert: HighErrorRate
        expr: sum(rate(ispmonitor_http_requests_total{status=~"5.."}[5m])) / sum(rate(ispmonitor_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: RoutersOffline
        expr: ispmonitor_routers_total{status="offline"} > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Routers offline"
          description: "{{ $value }} routers are offline"

      - alert: LicenseExpiringSoon
        expr: ispmonitor_license_days_until_expiry < 30
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "License expiring soon"
          description: "License expires in {{ $value }} days"

      - alert: HighPollingLatency
        expr: histogram_quantile(0.99, rate(ispmonitor_poll_duration_seconds_bucket[5m])) > 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High polling latency"
          description: "99th percentile polling latency is {{ $value }}s"
```

### Alertmanager Configuration

```yaml
# alertmanager.yml
route:
  receiver: 'default'
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

receivers:
  - name: 'default'
    email_configs:
      - to: 'ops@example.com'
        from: 'alertmanager@example.com'
    webhook_configs:
      - url: 'https://hooks.slack.com/services/xxx'
```

## Backup and Restore

### Database Backup

#### Automated Backups (Recommended)

```bash
#!/bin/bash
# backup.sh - Run daily via cron

BACKUP_DIR="/backups/ispmonitor"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d ispmonitor | gzip > "${BACKUP_DIR}/backup_${DATE}.sql.gz"

# Upload to S3 (optional)
aws s3 cp "${BACKUP_DIR}/backup_${DATE}.sql.gz" "s3://your-bucket/backups/"

# Cleanup old backups
find ${BACKUP_DIR} -name "backup_*.sql.gz" -mtime +${RETENTION_DAYS} -delete
```

#### Docker Backup

```bash
# Backup
docker-compose exec -T postgres pg_dump -U ispmonitor ispmonitor | gzip > backup.sql.gz

# Restore
gunzip -c backup.sql.gz | docker-compose exec -T postgres psql -U ispmonitor ispmonitor
```

#### Kubernetes Backup

```bash
# Backup
kubectl -n ispmonitor exec deployment/postgres -- pg_dump -U ispmonitor ispmonitor | gzip > backup.sql.gz

# Restore
gunzip -c backup.sql.gz | kubectl -n ispmonitor exec -i deployment/postgres -- psql -U ispmonitor ispmonitor
```

### Configuration Backup

Always backup your configuration:

```bash
# Environment files
cp .env .env.backup

# Kubernetes secrets
kubectl -n ispmonitor get secret ispmonitor-secrets -o yaml > secrets-backup.yaml
```

## Scaling

### Horizontal Scaling

#### Docker Compose

```bash
# Scale API replicas
docker-compose up -d --scale api=3
```

#### Kubernetes

```bash
# Manual scaling
kubectl -n ispmonitor scale deployment ispmonitor-api --replicas=5

# Autoscaling is configured via HPA
kubectl -n ispmonitor get hpa
```

### Vertical Scaling

Update resource limits:

```yaml
# Kubernetes
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: "1"
    memory: 1Gi
```

### Database Scaling

For large deployments:

1. **Connection Pooling**: Use PgBouncer
2. **Read Replicas**: Configure PostgreSQL streaming replication
3. **Partitioning**: Enable TimescaleDB for time-series data

### Scaling Guidelines

| Routers | API Replicas | Poller Replicas | Database Resources |
|---------|--------------|-----------------|-------------------|
| 1-100 | 1 | 1 | 2 CPU, 4 GB RAM |
| 100-500 | 2 | 1-2 | 4 CPU, 8 GB RAM |
| 500-1000 | 3-4 | 2-3 | 8 CPU, 16 GB RAM |
| 1000+ | 5+ | 3+ | 16+ CPU, 32+ GB RAM |

## Maintenance

### Updating

#### Docker Compose

```bash
# Pull new images
docker-compose pull

# Apply updates (with zero downtime)
docker-compose up -d

# Verify
docker-compose ps
curl http://localhost:8080/health
```

#### Kubernetes

```bash
# Update image tag
kubectl -n ispmonitor set image deployment/ispmonitor-api api=ghcr.io/mohamadkhaledabbas/ispvisualmonitor:v1.1.0

# Monitor rollout
kubectl -n ispmonitor rollout status deployment/ispmonitor-api

# Rollback if needed
kubectl -n ispmonitor rollout undo deployment/ispmonitor-api
```

### Database Maintenance

```sql
-- Analyze tables for query optimization
ANALYZE;

-- Vacuum to reclaim space
VACUUM ANALYZE;

-- Reindex if needed
REINDEX DATABASE ispmonitor;
```

### Log Rotation

Configure log rotation for Docker:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

## Troubleshooting

### Common Issues

#### Application Won't Start

1. **Check logs:**
   ```bash
   docker-compose logs api
   kubectl -n ispmonitor logs deployment/ispmonitor-api
   ```

2. **Verify database connection:**
   ```bash
   docker-compose exec api ./ispmonitor healthcheck db
   ```

3. **Check configuration:**
   ```bash
   docker-compose config
   ```

#### High Memory Usage

1. **Check for memory leaks:**
   ```bash
   # Enable profiling
   ENABLE_PROFILING=true
   
   # Access pprof
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

2. **Review connection pools:**
   - Database connections
   - HTTP client connections

#### Slow Polling

1. **Check poller metrics:**
   ```promql
   histogram_quantile(0.99, rate(ispmonitor_poll_duration_seconds_bucket[5m]))
   ```

2. **Increase workers:**
   ```bash
   POLLER_WORKERS=20
   POLLER_CONCURRENT=100
   ```

3. **Check router connectivity:**
   - Network latency
   - Router response time
   - Firewall rules

#### License Issues

1. **Check license status:**
   ```bash
   curl http://localhost:8080/api/v1/license/status
   ```

2. **Verify connectivity:**
   ```bash
   curl -I https://license.ispmonitor.com/v1/health
   ```

3. **Check offline grace period:**
   - Default: 72 hours
   - Cached license location: `/app/data/license.cache`

### Debug Mode

Enable debug logging:

```bash
LOG_LEVEL=debug
```

### Getting Support

1. **Collect diagnostics:**
   ```bash
   # System info
   docker-compose exec api ./ispmonitor diagnostics > diagnostics.txt
   
   # Logs
   docker-compose logs --tail=1000 > logs.txt
   ```

2. **Contact support:**
   - Email: support@ispmonitor.com
   - Include: diagnostics, logs, and steps to reproduce

### Useful Commands

```bash
# Check all container statuses
docker-compose ps

# View real-time logs
docker-compose logs -f

# Execute command in container
docker-compose exec api sh

# Check database status
docker-compose exec postgres psql -U ispmonitor -c "SELECT version();"

# Check Redis status
docker-compose exec redis redis-cli ping

# Force recreation of containers
docker-compose up -d --force-recreate

# Remove all data (WARNING: destructive)
docker-compose down -v
```

---

For additional support, visit: https://github.com/MohamadKhaledAbbas/ISPVisualMonitor/issues
