# ISP Visual Monitor - System Architecture

## Overview

ISP Visual Monitor is a production-grade, multi-tenant ISP monitoring platform that provides real-time network visibility through an interactive map-based interface. The system is designed for cloud deployment with support for multiple ISPs in a single instance.

## Core Principles

### 1. Multi-Tenancy
- **Tenant Isolation**: Each ISP operates in its own logical namespace with complete data isolation
- **Shared Infrastructure**: All tenants share the same application and database infrastructure
- **Tenant Context**: All database queries and operations are scoped to the authenticated tenant

### 2. Role-Based Access Control (RBAC)
- **Hierarchical Roles**: Admin, Manager, Engineer, Viewer
- **Permission Model**: Fine-grained permissions for network operations
- **Tenant-Scoped**: Roles and permissions are scoped within each tenant

### 3. Topology-Aware Architecture
- **Network Modeling**: Routers, links, and network devices represented as a graph
- **Geographic Awareness**: All network elements have geographic coordinates
- **Hierarchical Structure**: Support for regions, POPs, and edge locations

## System Components

### Frontend
- **Technology**: Modern web framework with MapLibre GL JS
- **Maps**: Self-hosted maps using PMTiles for offline capability
- **Real-time Updates**: WebSocket connections for live monitoring data

### Backend (Go)
- **API Server**: RESTful API with JWT authentication
- **Router Poller**: Concurrent SNMP/API polling of network devices
- **Data Processor**: Real-time metrics aggregation and alerting
- **Webhook Handler**: Integration with external monitoring systems

### Database (PostgreSQL)
- **Multi-tenant Schema**: Row-level security and tenant isolation
- **Time-series Data**: Optimized tables for metrics storage
- **Full-text Search**: Device and topology search capabilities

### Infrastructure
- **Container Orchestration**: Docker Compose for development, Kubernetes for production
- **Caching**: Redis for session management and metric caching
- **Message Queue**: For asynchronous job processing

## Data Flow

```
External Router (SNMP/API)
    ↓
Router Poller (Go concurrent workers)
    ↓
Data Processor & Validation
    ↓
PostgreSQL (time-series + topology)
    ↓
API Server (tenant-scoped queries)
    ↓
Frontend (map visualization)
```

## Scalability Considerations

### Horizontal Scaling
- **API Servers**: Stateless design allows multiple instances behind load balancer
- **Poller Workers**: Distributed polling across multiple worker nodes
- **Database**: Read replicas for query scaling

### Performance Optimization
- **Connection Pooling**: Efficient database connection management
- **Caching Strategy**: Multi-layer caching (in-memory, Redis, CDN)
- **Batch Processing**: Aggregate polling results before database writes

## Security Architecture

### Authentication & Authorization
- **JWT Tokens**: Stateless authentication with refresh token rotation
- **API Keys**: For programmatic access and integrations
- **Multi-Factor Authentication**: Optional 2FA for admin users

### Network Security
- **TLS Everywhere**: End-to-end encryption for all communications
- **SNMP v3**: Secure polling of network devices
- **Network Segmentation**: Isolated polling network from public API

### Data Security
- **Encryption at Rest**: Database and backup encryption
- **Audit Logging**: All administrative actions logged
- **Secrets Management**: Environment-based configuration

## Monitoring & Observability

### Application Metrics
- **Prometheus**: Metrics collection and alerting
- **Custom Metrics**: Polling success rate, API latency, tenant usage

### Logging
- **Structured Logging**: JSON format for easy parsing
- **Log Aggregation**: Centralized logging with retention policies

### Tracing
- **Distributed Tracing**: Request tracing across services
- **Performance Profiling**: Go pprof for performance analysis

## Deployment Architecture

### Development Environment
- Docker Compose with hot-reload for rapid development

### Production Environment
- Kubernetes deployment with:
  - Auto-scaling for API and poller services
  - High-availability PostgreSQL with replication
  - Redis cluster for distributed caching
  - Ingress with TLS termination

## Technology Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15+ with PostGIS extension
- **Cache**: Redis 7+
- **Frontend**: MapLibre GL JS, PMTiles
- **Infrastructure**: Docker, Kubernetes
- **Monitoring**: Prometheus, Grafana
