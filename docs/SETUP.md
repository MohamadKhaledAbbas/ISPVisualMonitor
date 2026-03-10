# ISP Visual Monitor - Setup Guide

This guide will walk you through setting up the ISP Visual Monitor platform from scratch.

## Prerequisites

Ensure you have the following installed:

- **Docker** (20.10+) and **Docker Compose** (v2.0+)
- **Go** (1.21+) - only for local development
- **PostgreSQL** (15+) with PostGIS - only if not using Docker
- **Git**

## Quick Setup (Docker)

This is the fastest way to get the platform running.

### Step 1: Clone and Configure

```bash
# Clone the repository
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor.git
cd ISPVisualMonitor

# Create environment file
cp .env.example .env
```

### Step 2: Configure Environment

Edit `.env` file and set at minimum:

```bash
JWT_SECRET=your-very-secure-random-secret-key-here-change-this
```

Generate a secure JWT secret:
```bash
openssl rand -base64 32
```

### Step 3: Start Services

```bash
# Start all services in background
docker-compose up -d

# Check status
docker-compose ps
```

You should see the following services running:
- `ispmonitor-postgres` - Database
- `ispmonitor-redis` - Cache
- `ispmonitor-app` - API server
- `ispmonitor-nginx` - Web server
- `ispmonitor-prometheus` - Metrics
- `ispmonitor-grafana` - Dashboards

### Step 4: Initialize Database

The database schema is automatically loaded on first start from the `db/migrations/` directory.

Verify the database:
```bash
docker-compose exec postgres psql -U ispmonitor -d ispmonitor -c "\dt"
```

### Step 5: Verify Installation

Check the health endpoint:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "isp-visual-monitor"
}
```

### Step 6: Access Services

- **API**: http://localhost:8080
- **Frontend**: http://localhost (when frontend is built)
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

## Local Development Setup

For active development, you may want to run the Go application directly.

### Step 1: Install Dependencies

```bash
# Install Go dependencies
go mod download

# Verify build
go build ./...
```

### Step 2: Start Infrastructure

Start only the database and cache:
```bash
docker-compose up -d postgres redis
```

### Step 3: Configure Environment

Create `.env` file with local database settings:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=ispmonitor
DB_PASSWORD=ispmonitor
DB_NAME=ispmonitor
API_PORT=8080
JWT_SECRET=your-secret-key
```

### Step 4: Run Migrations

```bash
# Run database migrations
psql -h localhost -U ispmonitor -d ispmonitor -f db/migrations/001_initial_schema.sql
```

Or using make:
```bash
make migrate
```

### Step 5: Run the Application

```bash
# Run directly
go run cmd/ispmonitor/main.go

# Or use make
make run

# Or build and run
make build
./bin/ispmonitor
```

## Production Deployment

### Docker Compose (Small Scale)

For small deployments, the included `docker-compose.yml` works well:

1. Configure environment variables for production
2. Set strong passwords and secrets
3. Configure TLS certificates in Nginx
4. Set up backup strategy for PostgreSQL
5. Configure log rotation

### Kubernetes (Large Scale)

For production Kubernetes deployment:

1. **Database**: Use managed PostgreSQL (AWS RDS, Google Cloud SQL, Azure Database)
2. **Cache**: Use managed Redis (AWS ElastiCache, Google Memorystore)
3. **Application**: Deploy as Kubernetes Deployment with:
   - HorizontalPodAutoscaler for API servers
   - Separate deployment for poller workers
   - ConfigMaps for configuration
   - Secrets for sensitive data

Example Kubernetes structure:
```
k8s/
├── namespace.yaml
├── configmap.yaml
├── secrets.yaml
├── postgres.yaml (or use managed service)
├── redis.yaml (or use managed service)
├── api-deployment.yaml
├── api-service.yaml
├── poller-deployment.yaml
├── ingress.yaml
└── hpa.yaml
```

## Database Setup

### Manual PostgreSQL Setup

If you're not using Docker for PostgreSQL:

```bash
# Create database and user
sudo -u postgres psql << EOF
CREATE USER ispmonitor WITH PASSWORD 'your-password';
CREATE DATABASE ispmonitor OWNER ispmonitor;
\c ispmonitor
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
EOF

# Run migrations
psql -U ispmonitor -d ispmonitor -f db/migrations/001_initial_schema.sql
```

### Database Backup

Setup automated backups:

```bash
# Backup script
pg_dump -U ispmonitor -d ispmonitor -F c -f backup_$(date +%Y%m%d).dump

# Restore script
pg_restore -U ispmonitor -d ispmonitor -c backup_20240116.dump
```

## Map Tiles Setup

The platform uses self-hosted map tiles with PMTiles format.

### Option 1: Download Pre-built Tiles

Download from [Protomaps](https://protomaps.com/):

```bash
# Download world tiles (example)
cd docker/tiles/
wget https://build.protomaps.com/latest.pmtiles -O world.pmtiles
```

### Option 2: Generate Custom Tiles

Using Planetiler:

```bash
# Install planetiler
# Download OSM data for your region
# Generate PMTiles

java -jar planetiler.jar \
  --download \
  --area=your-region \
  --output=docker/tiles/region.pmtiles
```

See [docker/tiles/README.md](docker/tiles/README.md) for more details.

## Initial Data Setup

### Create First Tenant

```sql
INSERT INTO tenants (name, slug, contact_email, subscription_tier)
VALUES ('Demo ISP', 'demo-isp', 'admin@demo-isp.com', 'enterprise');
```

### Create First User

```sql
-- Get tenant ID
SELECT id FROM tenants WHERE slug = 'demo-isp';

-- Create user (use bcrypt hash for password)
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, email_verified)
VALUES (
  '<tenant-id>',
  'admin@demo-isp.com',
  '<bcrypt-hash>',
  'Admin',
  'User',
  true
);
```

Generate bcrypt hash in Go:
```go
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
    fmt.Println(string(hash))
}
```

### Assign Admin Role

```sql
-- Create admin role
INSERT INTO roles (tenant_id, name, description, is_system)
VALUES ('<tenant-id>', 'Admin', 'Administrator with full access', true);

-- Assign all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '<role-id>', id FROM permissions;

-- Assign admin role to user
INSERT INTO user_roles (user_id, role_id)
VALUES ('<user-id>', '<role-id>');
```

## Testing the Setup

### Test API Authentication

```bash
# Login (once authentication is implemented)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@demo-isp.com",
    "password": "password"
  }'
```

### Add Test Router

```bash
# Create a test router (with valid JWT token)
curl -X POST http://localhost:8080/api/v1/routers \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "NYC-CORE-01",
    "hostname": "nyc-core-01.demo-isp.com",
    "management_ip": "10.0.1.1",
    "router_type": "core",
    "vendor": "Cisco",
    "model": "ASR9000"
  }'
```

## Monitoring Setup

### Configure Grafana

1. Access Grafana at http://localhost:3000
2. Login with admin/admin
3. Add Prometheus data source: http://prometheus:9090
4. Import dashboards for:
   - System metrics
   - Application metrics
   - Database metrics

### View Prometheus Metrics

Access http://localhost:9090 and query:
- `up` - Check service health
- `go_goroutines` - Application goroutines
- `process_cpu_seconds_total` - CPU usage

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U ispmonitor -d ispmonitor -c "SELECT 1"
```

### Application Won't Start

```bash
# Check application logs
docker-compose logs app

# Common issues:
# - JWT_SECRET not set
# - Database not ready (wait for health check)
# - Port already in use
```

### Cannot Access Services

```bash
# Check all services are running
docker-compose ps

# Check port bindings
docker-compose port app 8080
docker-compose port nginx 80

# Check firewall rules
sudo ufw status
```

## Next Steps

After successful setup:

1. **Implement Authentication** - Complete the auth handlers in `internal/api/`
2. **Add Routers** - Populate your network topology
3. **Configure Polling** - Set up SNMP credentials for routers
4. **Create Alerts** - Define alert rules for your network
5. **Build Frontend** - Develop the web interface with MapLibre
6. **Add Users** - Create user accounts for your team

## Additional Resources

- [System Architecture](docs/architecture/SYSTEM_ARCHITECTURE.md)
- [Multi-Tenant Design](docs/architecture/MULTI_TENANT.md)
- [RBAC Documentation](docs/architecture/RBAC.md)
- [Topology Architecture](docs/architecture/TOPOLOGY_AWARE.md)
- [API Documentation](README.md#api-documentation)

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- Review documentation in `docs/`
- Open an issue on GitHub
