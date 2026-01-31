# ISP Visual Monitor - Deployment Guide

This document provides comprehensive instructions for deploying ISP Visual Monitor in various environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Development Setup](#development-setup)
- [Production Deployment](#production-deployment)
  - [Docker Compose](#docker-compose-production)
  - [Kubernetes](#kubernetes-deployment)
- [Environment Variables](#environment-variables)
- [Secrets Management](#secrets-management)
- [Health Checks](#health-checks)
- [Upgrading](#upgrading)

## Quick Start

### Prerequisites

- Docker and Docker Compose (for containerized deployment)
- Go 1.21+ (for local development)
- PostgreSQL 15+ with PostGIS extension
- Redis 7+

### Fastest Way to Start

```bash
# Clone the repository
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor.git
cd ISPVisualMonitor

# Start with Docker Compose
docker-compose up -d

# Access the application
open http://localhost:8080
```

## Development Setup

### Local Development with Docker

1. **Clone and configure:**
   ```bash
   git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor.git
   cd ISPVisualMonitor
   cp .env.example .env
   ```

2. **Edit `.env`** with your development settings

3. **Start services:**
   ```bash
   docker-compose up -d
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f api
   ```

### Local Development without Docker

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Start PostgreSQL and Redis** (manually or via Docker)

3. **Run the application:**
   ```bash
   export JWT_SECRET="your-dev-secret"
   export DB_HOST="localhost"
   go run cmd/ispmonitor/main.go
   ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linters
make lint
```

## Production Deployment

### Docker Compose Production

1. **Prepare environment:**
   ```bash
   # Copy production compose file
   cp docker-compose.prod.yml docker-compose.override.yml
   
   # Create production environment file
   cp .env.example .env
   ```

2. **Configure secrets:**
   ```bash
   # Edit .env with production values
   vim .env
   ```

   Required production values:
   - `JWT_SECRET` - Strong random secret (32+ characters)
   - `DATABASE_URL` - Production PostgreSQL connection string
   - `REDIS_URL` - Production Redis connection string
   - `LICENSE_KEY` - Your license key (if applicable)

3. **Deploy:**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

4. **Verify deployment:**
   ```bash
   curl http://localhost:8080/health
   ```

### Kubernetes Deployment

#### Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm (optional, for cert-manager)

#### Using Kustomize

1. **Create namespace and secrets:**
   ```bash
   # Create namespace
   kubectl apply -f deploy/k8s/base/namespace.yaml
   
   # Create secrets (edit first!)
   kubectl apply -f deploy/k8s/base/secrets.yaml
   ```

2. **Deploy using overlays:**

   **Development:**
   ```bash
   kubectl apply -k deploy/k8s/overlays/development
   ```

   **Staging:**
   ```bash
   kubectl apply -k deploy/k8s/overlays/staging
   ```

   **Production:**
   ```bash
   kubectl apply -k deploy/k8s/overlays/production
   ```

3. **Verify deployment:**
   ```bash
   kubectl -n ispmonitor get pods
   kubectl -n ispmonitor get svc
   ```

#### Setting up Ingress with TLS

1. **Install cert-manager:**
   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   ```

2. **Create ClusterIssuer:**
   ```yaml
   apiVersion: cert-manager.io/v1
   kind: ClusterIssuer
   metadata:
     name: letsencrypt-prod
   spec:
     acme:
       server: https://acme-v02.api.letsencrypt.org/directory
       email: your-email@example.com
       privateKeySecretRef:
         name: letsencrypt-prod
       solvers:
         - http01:
             ingress:
               class: nginx
   ```

3. **Apply the ingress** (already configured in base manifests)

## Environment Variables

### Core Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DEPLOYMENT_MODE` | Deployment mode (development/production/on-premise) | development |
| `API_PORT` | API server port | 8080 |
| `JWT_SECRET` | JWT signing secret | - (required) |

### Database Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | PostgreSQL user | ispmonitor |
| `DB_PASSWORD` | PostgreSQL password | ispmonitor |
| `DB_NAME` | Database name | ispmonitor |
| `DB_SSLMODE` | SSL mode | disable |
| `DATABASE_URL` | Full connection string (overrides individual vars) | - |

### Redis Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_URL` | Redis connection URL | redis://localhost:6379 |

### License Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `LICENSE_KEY` | License key for production | - |
| `LICENSE_SERVER_URL` | License validation server | https://license.ispmonitor.com/v1 |

### Observability

| Variable | Description | Default |
|----------|-------------|---------|
| `ENABLE_METRICS` | Enable Prometheus metrics | true |
| `METRICS_PORT` | Metrics endpoint port | 9090 |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | info |
| `LOG_FORMAT` | Log format (json/text) | json |

## Secrets Management

### Recommended Approaches

1. **Kubernetes Secrets** (basic):
   ```bash
   kubectl create secret generic ispmonitor-secrets \
     --from-literal=JWT_SECRET="your-secret" \
     --from-literal=DATABASE_URL="postgres://..." \
     -n ispmonitor
   ```

2. **Sealed Secrets** (recommended for GitOps):
   ```bash
   kubeseal --format yaml < secrets.yaml > sealed-secrets.yaml
   ```

3. **External Secrets Operator** (enterprise):
   Configure with AWS Secrets Manager, HashiCorp Vault, etc.

### Never Do

- Commit secrets to Git
- Use default/example secrets in production
- Share secrets in plaintext

## Health Checks

### Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Full health check (database, redis) |
| `GET /ready` | Kubernetes readiness probe |
| `GET /live` | Kubernetes liveness probe |

### Example Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "deployment": "production",
  "checks": {
    "database": {
      "status": "healthy",
      "duration": "2ms"
    },
    "redis": {
      "status": "healthy",
      "duration": "1ms"
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Upgrading

### Docker Compose

```bash
# Pull latest images
docker-compose pull

# Restart with new images
docker-compose up -d

# Verify
curl http://localhost:8080/health
```

### Kubernetes

```bash
# Update image tag in kustomization.yaml
# Then apply
kubectl apply -k deploy/k8s/overlays/production

# Monitor rollout
kubectl -n ispmonitor rollout status deployment/ispmonitor-api
```

### Database Migrations

Migrations are automatically applied on startup. For manual execution:

```bash
# Docker
docker-compose exec api ./ispmonitor migrate

# Kubernetes
kubectl -n ispmonitor exec deployment/ispmonitor-api -- ./ispmonitor migrate
```

## Troubleshooting

### Common Issues

1. **Database connection failed**
   - Verify DATABASE_URL or DB_* variables
   - Check network connectivity
   - Ensure PostGIS extension is installed

2. **License validation failed**
   - Verify LICENSE_KEY is correct
   - Check connectivity to license server
   - Review offline grace period status

3. **Health check failing**
   - Check container logs: `docker-compose logs api`
   - Verify all dependencies are running
   - Check resource limits

### Getting Help

- GitHub Issues: https://github.com/MohamadKhaledAbbas/ISPVisualMonitor/issues
- Documentation: https://github.com/MohamadKhaledAbbas/ISPVisualMonitor/tree/main/docs
