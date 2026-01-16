# Role-Based Access Control (RBAC)

## Overview

ISP Visual Monitor implements a flexible Role-Based Access Control system that provides fine-grained permissions for network monitoring operations while maintaining simplicity for common use cases.

## Role Hierarchy

### Predefined Roles

#### 1. Super Admin (Platform-Level)
- **Scope**: Cross-tenant, platform administration
- **Permissions**: 
  - Create and manage tenants
  - View all tenant data
  - System configuration
  - Platform-wide monitoring

#### 2. Tenant Admin
- **Scope**: Single tenant, full administrative access
- **Permissions**:
  - User management (create, update, delete users)
  - Role assignment within tenant
  - Device management (add, configure, delete)
  - Topology management
  - Alert configuration
  - API key management
  - Billing and subscription management

#### 3. Manager
- **Scope**: Single tenant, operational management
- **Permissions**:
  - View all devices and metrics
  - Add and configure devices
  - Create and modify alert rules
  - View and acknowledge alerts
  - Export reports
  - No user management

#### 4. Engineer
- **Scope**: Single tenant, network operations
- **Permissions**:
  - View devices and real-time metrics
  - Acknowledge alerts
  - Run diagnostic commands (ping, traceroute)
  - View topology
  - Create notes and annotations
  - No configuration changes

#### 5. Viewer
- **Scope**: Single tenant, read-only access
- **Permissions**:
  - View dashboard and maps
  - View device status and metrics
  - View alerts (read-only)
  - View topology
  - No write operations

### Custom Roles
Tenants can create custom roles by combining permissions:
- Start from a base role template
- Add or remove specific permissions
- Assign to users as needed

## Permission Model

### Permission Categories

#### Network Management
- `network.devices.read` - View devices
- `network.devices.create` - Add new devices
- `network.devices.update` - Modify device configuration
- `network.devices.delete` - Remove devices
- `network.topology.read` - View network topology
- `network.topology.update` - Modify topology

#### Monitoring & Alerts
- `monitoring.metrics.read` - View metrics and graphs
- `monitoring.alerts.read` - View alerts
- `monitoring.alerts.acknowledge` - Acknowledge alerts
- `monitoring.alerts.configure` - Create/modify alert rules

#### User & Access Management
- `users.read` - View user list
- `users.create` - Create new users
- `users.update` - Modify user details
- `users.delete` - Remove users
- `users.roles.assign` - Assign roles to users

#### Configuration
- `config.tenant.read` - View tenant settings
- `config.tenant.update` - Modify tenant settings
- `config.polling.update` - Configure polling parameters
- `config.integration.manage` - Manage webhooks and integrations

#### Data & Reporting
- `data.export` - Export data and reports
- `data.import` - Import device configurations
- `reports.generate` - Generate custom reports

## Implementation

### Database Schema

```sql
-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_custom BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_role_per_tenant UNIQUE(tenant_id, name)
);

-- Permissions table (global)
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT
);

-- Role-Permission mapping
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Role assignment
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);
```

### Authorization Middleware

```go
// Permission check middleware
func RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value(UserKey).(*User)
            
            if !user.HasPermission(permission) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Role check middleware
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value(UserKey).(*User)
            
            if !user.HasAnyRole(roles...) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Usage Example

```go
// API route with permission check
router.Handle("/api/devices", 
    RequirePermission("network.devices.create")(
        http.HandlerFunc(createDeviceHandler),
    ),
).Methods("POST")

// API route with role check
router.Handle("/api/admin/users",
    RequireRole("tenant_admin")(
        http.HandlerFunc(listUsersHandler),
    ),
).Methods("GET")
```

## Permission Evaluation

### Evaluation Flow

1. **Extract User Context**: Get user ID from JWT token
2. **Load User Roles**: Query all roles assigned to the user
3. **Aggregate Permissions**: Collect all permissions from all roles
4. **Check Permission**: Verify if required permission exists in set
5. **Grant/Deny**: Allow or reject the request

### Caching Strategy
- User permissions cached in Redis for performance
- Cache key: `user:{user_id}:permissions`
- TTL: 5 minutes
- Invalidate on role assignment changes

## Resource-Level Authorization

### Object Ownership
Some resources can be owned by specific users:
- Alert configurations
- Custom dashboards
- Saved searches

### Access Patterns
- **Owner**: Full access to their resources
- **Team**: Shared access within defined teams
- **Admin**: Admins can access all resources within tenant

## Audit Logging

All authorization decisions are logged:
- User attempting action
- Required permission
- Grant/Deny result
- Timestamp and request context

```go
func logAuthorizationDecision(userID, permission string, granted bool, ctx context.Context) {
    log.WithFields(log.Fields{
        "user_id": userID,
        "permission": permission,
        "granted": granted,
        "tenant_id": ctx.Value(TenantIDKey),
        "request_id": ctx.Value(RequestIDKey),
    }).Info("authorization decision")
}
```

## Best Practices

### Principle of Least Privilege
- Assign minimum necessary permissions
- Start with Viewer role, escalate as needed
- Regular permission audits

### Role Assignment
- Use predefined roles when possible
- Document custom role purposes
- Review role assignments quarterly

### Permission Naming
- Follow consistent naming convention: `category.resource.action`
- Use descriptive names
- Avoid overly broad permissions

## API Integration

### JWT Claims
```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "roles": ["engineer", "custom-role-uuid"],
  "permissions": [
    "network.devices.read",
    "monitoring.metrics.read",
    "monitoring.alerts.acknowledge"
  ]
}
```

### Permission Check in Frontend
Frontend receives permission list and can:
- Hide/show UI elements based on permissions
- Disable actions user cannot perform
- Provide better UX without unnecessary API calls

## Testing Authorization

### Unit Tests
```go
func TestUserHasPermission(t *testing.T) {
    user := &User{
        Roles: []Role{
            {Permissions: []string{"network.devices.read"}},
        },
    }
    
    assert.True(t, user.HasPermission("network.devices.read"))
    assert.False(t, user.HasPermission("network.devices.delete"))
}
```

### Integration Tests
- Test role assignment workflow
- Verify permission enforcement at API level
- Test custom role creation and usage
