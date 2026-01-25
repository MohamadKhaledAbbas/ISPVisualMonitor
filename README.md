# ISP Visual Monitor

A production-grade, multi-tenant ISP monitoring platform with Role-Based Access Control (RBAC) and Topology-Aware architecture. Monitor network infrastructure visually using interactive maps powered by MapLibre GL JS and self-hosted PMTiles.

## Features

- 🌐 **Multi-Tenant Architecture** - Support multiple ISPs in a single deployment
- 🔐 **Role-Based Access Control** - Fine-grained permissions and role management
- 🗺️ **Topology-Aware Monitoring** - Network devices and links visualized on maps
- 🚀 **High-Performance Polling** - Concurrent Go workers for SNMP/API polling
- 🗃️ **PostgreSQL + PostGIS** - Robust storage with geographic capabilities
- 🐳 **Docker-Ready** - Complete Docker Compose setup for easy deployment
- 📊 **Real-Time Metrics** - Interface and router metrics with time-series storage
- 🔔 **Alert Management** - Configurable alert rules and notification system
- 🗺️ **Self-Hosted Maps** - MapLibre GL JS with PMTiles (no external dependencies)
- 🔄 **Multi-Role Routers** - Routers can have multiple roles (core, edge, PPPoE, NAT, etc.)
- 🔌 **Vendor Abstraction** - Pluggable adapter pattern supporting MikroTik, Cisco, Juniper, and more
- 📡 **Multiple Connection Methods** - API, SNMP, SSH, NETCONF with automatic fallback
- 👥 **Session Tracking** - Monitor PPPoE, NAT, and DHCP sessions in real-time
- 🏗️ **Router Dependencies** - Model upstream/downstream/peer relationships for impact analysis

## Architecture

The system follows a modern microservices-inspired architecture with a **pluggable adapter pattern** for vendor abstraction:

- **API Server (Go)** - RESTful API with JWT authentication
- **Enhanced Router Poller (Go)** - Concurrent worker pool with vendor-agnostic polling
- **Adapter Framework** - Pluggable adapters for MikroTik, Cisco, Juniper, and more
- **PostgreSQL + PostGIS** - Primary database with geographic extensions and multi-role support
- **Redis** - Caching and session management
- **Nginx** - Frontend and tile serving
- **MapLibre GL JS** - Self-hosted map visualization

### Key Architectural Improvements

- **Multi-Role Router Support**: Routers can have multiple roles (e.g., PPPoE server + NAT gateway)
- **Vendor Abstraction**: Pluggable adapter pattern allows easy addition of new vendors
- **Connection Flexibility**: Support for API, SNMP, SSH, NETCONF with automatic fallback
- **Session Tracking**: Track PPPoE, NAT, and DHCP sessions across the network
- **Router Dependencies**: Model and visualize router relationships for impact analysis

For detailed architecture documentation, see:
- [System Architecture](docs/architecture/SYSTEM_ARCHITECTURE.md)
- [Multi-Tenant Design](docs/architecture/MULTI_TENANT.md)
- [Role-Based Access Control](docs/architecture/RBAC.md)
- [Topology-Aware Architecture](docs/architecture/TOPOLOGY_AWARE.md)
- **NEW:** [CHR Lab Setup](docs/CHR_SETUP.md) - MikroTik test environment
- **NEW:** [Adapter Development Guide](docs/ADAPTER_DEVELOPMENT.md) - Create custom adapters

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 15+ with PostGIS (if not using Docker)

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor.git
cd ISPVisualMonitor
```

2. Create environment file:
```bash
cp .env.example .env
# Edit .env and set JWT_SECRET and other configuration
```

3. Start all services:
```bash
docker-compose up -d
```

4. Initialize the database:
```bash
# Run initial schema
docker-compose exec postgres psql -U ispmonitor -d ispmonitor -f /docker-entrypoint-initdb.d/001_initial_schema.sql

# Run enhanced schema (multi-role support)
docker-compose exec postgres psql -U ispmonitor -d ispmonitor -f /docker-entrypoint-initdb.d/002_enhanced_router_schema.sql
```

5. Access the application:
- API: http://localhost:8080
- Frontend: http://localhost
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL with PostGIS:
```bash
# Using Docker for database only
docker-compose up -d postgres redis
```

3. Run database migrations:
```bash
psql -U ispmonitor -d ispmonitor -f db/migrations/001_initial_schema.sql
psql -U ispmonitor -d ispmonitor -f db/migrations/002_enhanced_router_schema.sql
```

4. Configure environment:
```bash
cp .env.example .env
# Edit .env with your settings
```

5. Run the application:
```bash
go run cmd/ispmonitor/main.go
```

## Configuration

Configuration is done via environment variables. See [.env.example](.env.example) for all available options.

Key configuration areas:
- **Database**: PostgreSQL connection settings
- **API**: Server port, JWT secret, CORS settings
- **Poller**: Worker count, polling intervals, concurrency
- **Auth**: Token expiry, password hashing cost

## API Documentation

### Authentication

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# Response includes JWT token
```

### Router Management

```bash
# List routers (requires authentication)
curl http://localhost:8080/api/v1/routers \
  -H "Authorization: Bearer <token>"

# Get router details
curl http://localhost:8080/api/v1/routers/{id} \
  -H "Authorization: Bearer <token>"
```

### Topology

```bash
# Get topology as GeoJSON
curl http://localhost:8080/api/v1/topology/geojson \
  -H "Authorization: Bearer <token>"
```

## Database Schema

The database schema includes:

- **Tenant Management** - Tenants, users, roles, permissions
- **Network Topology** - Regions, POPs, routers, interfaces, links
- **Monitoring Data** - Interface metrics, router metrics (time-series)
- **Alert System** - Alert rules and active alerts
- **Audit Logging** - Complete audit trail
- **Enhanced Router Capabilities** (NEW):
  - **Router Roles** - Multi-role support (core, edge, PPPoE, NAT, etc.)
  - **Router Capabilities** - Connection methods (API, SNMP, SSH, NETCONF)
  - **Router Dependencies** - Upstream/downstream/peer relationships
  - **Session Tracking** - PPPoE sessions, NAT sessions, DHCP leases
  - **Polling History** - Track polling attempts and results

See migration files for complete schema:
- [Initial Schema](db/migrations/001_initial_schema.sql)
- [Enhanced Router Schema](db/migrations/002_enhanced_router_schema.sql)

## Map Setup

The platform uses self-hosted maps with MapLibre GL JS and PMTiles:

1. Download or generate PMTiles files
2. Place them in `docker/tiles/` directory
3. Configure MapLibre to use the tiles

See [docker/tiles/README.md](docker/tiles/README.md) for detailed instructions.

## Development

### Project Structure

```
ISPVisualMonitor/
├── cmd/
│   └── ispmonitor/          # Main application entry point
├── internal/
│   ├── api/                 # HTTP API handlers
│   ├── auth/                # Authentication logic
│   ├── database/            # Database connection
│   ├── middleware/          # HTTP middleware
│   └── poller/              # Router polling service
│       ├── adapter/         # Polling adapters (NEW)
│       │   ├── adapter.go          # Adapter interface
│       │   ├── mikrotik_adapter.go # MikroTik RouterOS API
│       │   ├── snmp_adapter.go     # Generic SNMP
│       │   └── registry.go         # Adapter registry
│       ├── poller.go               # Legacy poller
│       └── poller_enhanced.go      # Enhanced adapter-based poller (NEW)
├── pkg/
│   ├── config/              # Configuration management
│   └── models/              # Data models
│       ├── capabilities.go        # Router capabilities (NEW)
│       ├── role.go                # Router roles (NEW)
│       ├── router_enhanced.go     # Enhanced router model (NEW)
│       ├── sessions.go            # Session tracking (NEW)
│       └── models.go              # Base models
├── db/
│   └── migrations/          # Database migration scripts
│       ├── 001_initial_schema.sql
│       └── 002_enhanced_router_schema.sql (NEW)
├── docs/
│   ├── architecture/        # Architecture documentation
│   ├── CHR_SETUP.md         # CHR lab setup guide (NEW)
│   └── ADAPTER_DEVELOPMENT.md  # Adapter dev guide (NEW)
├── scripts/
│   └── chr/                 # CHR lab scripts (NEW)
│       ├── setup-chr-monitoring.rsc
│       └── simulate-pppoe-activity.sh
├── docker/                  # Docker configuration files
├── Dockerfile               # Application Dockerfile
├── docker-compose.yml       # Standard deployment
└── docker-compose.chr.yml   # CHR lab environment (NEW)
```

### Building

```bash
# Build the binary
go build -o bin/ispmonitor cmd/ispmonitor/main.go

# Build Docker image
docker build -t ispmonitor:latest .
```

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Deployment

### Docker Compose (Development/Small Deployments)

The included `docker-compose.yml` is suitable for:
- Development environments
- Small production deployments
- Testing and evaluation

### Kubernetes (Production)

For production deployments, consider:
- Kubernetes deployment with auto-scaling
- High-availability PostgreSQL (e.g., Patroni, CloudSQL)
- Redis cluster for distributed caching
- Load balancer for API servers
- Separate poller worker deployments

## Security

- **Authentication**: JWT-based with refresh tokens
- **Authorization**: Role-based access control (RBAC)
- **Multi-Tenancy**: Row-level security in PostgreSQL
- **SNMP**: Support for SNMPv3 with encryption
- **TLS**: Configure reverse proxy (Nginx) with TLS certificates
- **Secrets**: Use environment variables, never commit secrets

## Monitoring & Observability

The platform includes:
- **Prometheus** - Metrics collection
- **Grafana** - Visualization and dashboards
- **Structured Logging** - JSON logs for easy parsing
- **Health Checks** - `/health` endpoint for monitoring

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license here]

## Support

For issues and questions:
- Open an issue on GitHub
- Check the documentation in `docs/`
- Review architecture documents

## Roadmap

- [x] **Multi-Role Router Support** - Routers with multiple roles (PPPoE, NAT, etc.)
- [x] **Vendor Abstraction Layer** - Pluggable adapter pattern for different vendors
- [x] **MikroTik RouterOS Support** - Native API adapter for MikroTik devices
- [x] **Session Tracking** - PPPoE, NAT, and DHCP session monitoring
- [x] **Router Dependencies** - Model upstream/downstream relationships
- [x] **CHR Test Lab** - Docker-based MikroTik lab for testing
- [ ] Complete API handler implementations
- [ ] Frontend development with React/Vue
- [ ] Advanced SNMP polling with vendor MIBs
- [ ] Cisco IOS/XE adapter (RESTCONF)
- [ ] Juniper JUNOS adapter (NETCONF)
- [ ] Advanced alerting with webhooks
- [ ] User management UI
- [ ] Topology auto-discovery
- [ ] Historical data analysis
- [ ] Mobile application
- [ ] Integration with external monitoring systems

## Acknowledgments

- [MapLibre GL JS](https://maplibre.org/) - Open-source mapping library
- [PMTiles](https://github.com/protomaps/PMTiles) - Cloud-optimized map tiles
- [PostgreSQL](https://www.postgresql.org/) - Powerful open-source database
- [PostGIS](https://postgis.net/) - Spatial database extender
