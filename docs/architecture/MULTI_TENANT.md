# Multi-Tenant Architecture

## Overview

The ISP Visual Monitor platform implements a shared-database, shared-schema multi-tenancy model with row-level tenant isolation. This approach provides strong isolation while maximizing resource efficiency.

## Tenant Model

### Tenant Definition
A tenant represents a single ISP organization using the platform. Each tenant has:
- Unique tenant identifier (UUID)
- Organization metadata (name, contact, subscription tier)
- Independent user base with role assignments
- Isolated network topology and monitoring data

### Tenant Identification
- **Primary Key**: `tenant_id` (UUID) in all tenant-scoped tables
- **Context Propagation**: Tenant ID extracted from JWT token and propagated through request context
- **Database Enforcement**: PostgreSQL Row-Level Security (RLS) policies enforce isolation

## Database Design

### Tenant-Scoped Tables
All tables containing tenant-specific data include a `tenant_id` column:

```sql
CREATE TABLE routers (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    -- other columns
    CONSTRAINT unique_router_per_tenant UNIQUE(tenant_id, name)
);
```

### Global Tables
Some tables are global and not tenant-scoped:
- System configuration
- Audit logs (include tenant_id but not scoped)
- Background job status

### Row-Level Security

PostgreSQL RLS policies automatically filter queries:

```sql
ALTER TABLE routers ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON routers
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

## Application Layer

### Tenant Context Middleware

```go
func TenantContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant_id from JWT claims
        tenantID := extractTenantFromJWT(r)
        
        // Add to request context
        ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
        
        // Set PostgreSQL session variable
        db.Exec("SET app.current_tenant = $1", tenantID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Tenant-Scoped Queries

All database operations automatically include tenant context:

```go
func GetRouters(ctx context.Context, db *sql.DB) ([]Router, error) {
    tenantID := ctx.Value(TenantIDKey).(uuid.UUID)
    
    // Query automatically filtered by RLS policy
    rows, err := db.QueryContext(ctx, 
        "SELECT * FROM routers WHERE tenant_id = $1", 
        tenantID)
    // ...
}
```

## Tenant Onboarding

### Provisioning Flow

1. **Tenant Creation**
   - Admin creates tenant record
   - Generate API keys and initial credentials
   - Set resource quotas and limits

2. **User Setup**
   - Create admin user for tenant
   - Assign initial roles and permissions
   - Send welcome email with credentials

3. **Configuration**
   - Tenant-specific settings (polling intervals, alert thresholds)
   - Custom branding (optional)
   - Integration webhooks

### Data Migration
For tenants migrating from existing systems:
- Import API for bulk device upload
- Validation and deduplication
- Historical data import (optional)

## Resource Isolation

### Computational Limits
- **API Rate Limiting**: Per-tenant request quotas
- **Polling Limits**: Maximum devices per tenant
- **Storage Quotas**: Data retention policies based on subscription tier

### Fair Resource Sharing
- **Worker Pool Distribution**: Polling workers distributed across tenants
- **Priority Queues**: Higher-tier tenants get priority during high load
- **Backpressure**: Graceful degradation when resources are constrained

## Tenant Administration

### Self-Service Portal
Each tenant admin can:
- Manage users and roles
- Configure devices and topology
- View usage statistics and billing
- Configure alert rules and webhooks

### Platform Administration
Platform admins can:
- Create and manage tenants
- Monitor resource usage across all tenants
- Configure global system settings
- Review audit logs

## Security Considerations

### Tenant Isolation Guarantees
- **Database Level**: RLS policies prevent cross-tenant data access
- **Application Level**: All queries include tenant_id validation
- **API Level**: JWT tokens are tenant-scoped and validated on every request

### Audit & Compliance
- All tenant operations logged with tenant context
- Tenant data export for compliance (GDPR, etc.)
- Tenant data deletion workflow for account closure

## Performance Optimization

### Indexing Strategy
All tenant-scoped tables have composite indexes:
```sql
CREATE INDEX idx_routers_tenant_id ON routers(tenant_id);
CREATE INDEX idx_routers_tenant_created ON routers(tenant_id, created_at);
```

### Query Optimization
- Partition large tables by tenant_id for improved query performance
- Tenant-specific table statistics for query planner optimization

### Caching Strategy
- Cache keys include tenant_id prefix
- Tenant-specific cache eviction policies
- Shared caches for global data (map tiles, system config)

## Monitoring

### Per-Tenant Metrics
- API request volume and latency
- Polling success rate and error count
- Storage usage and growth rate
- Active user sessions

### Platform-wide Metrics
- Total tenant count and distribution
- Resource utilization by tenant tier
- Cross-tenant performance comparison

## Future Enhancements

### Advanced Isolation
- **Database-per-Tenant**: For enterprise customers requiring complete isolation
- **Schema-per-Tenant**: Middle ground between shared and isolated databases

### Tenant Marketplace
- Plugin system for tenant-specific extensions
- Third-party integration marketplace
- Custom visualization components
