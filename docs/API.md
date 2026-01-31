# ISP Visual Monitor API Documentation

## Overview

The ISP Visual Monitor API provides RESTful endpoints for managing network infrastructure, monitoring, and alerts. All endpoints except authentication require a valid JWT token.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

### Register a New User

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "tenant_id": "optional-tenant-uuid"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "status": "active",
  "email_verified": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Login

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "tenant_id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active"
  }
}
```

### Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `204 No Content`

## Routers

### List Routers

**Endpoint:** `GET /api/v1/routers`

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 20)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "name": "Router-01",
      "hostname": "router01.example.com",
      "management_ip": "192.168.1.1",
      "vendor": "Cisco",
      "model": "ISR4321",
      "status": "active",
      "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 1,
    "total_pages": 1
  }
}
```

### Create Router

**Endpoint:** `POST /api/v1/routers`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "name": "Router-01",
  "hostname": "router01.example.com",
  "management_ip": "192.168.1.1",
  "vendor": "Cisco",
  "model": "ISR4321",
  "location": {
    "latitude": 40.7128,
    "longitude": -74.0060
  }
}
```

**Response:** `201 Created`

### Get Router

**Endpoint:** `GET /api/v1/routers/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Update Router

**Endpoint:** `PUT /api/v1/routers/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "name": "Router-01-Updated",
  "status": "maintenance"
}
```

**Response:** `200 OK`

### Delete Router

**Endpoint:** `DELETE /api/v1/routers/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `204 No Content`

## Interfaces

### List All Interfaces

**Endpoint:** `GET /api/v1/interfaces`

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 20)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### List Router Interfaces

**Endpoint:** `GET /api/v1/routers/{router_id}/interfaces`

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 20)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

## Topology

### Get Network Topology

**Endpoint:** `GET /api/v1/topology`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "routers": [
    {
      "id": "uuid",
      "name": "Router-01",
      "management_ip": "192.168.1.1",
      "status": "active",
      "vendor": "Cisco",
      "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
      }
    }
  ],
  "links": [
    {
      "id": "uuid",
      "source_interface_id": "uuid",
      "target_interface_id": "uuid",
      "source_router_id": "uuid",
      "target_router_id": "uuid",
      "link_type": "ethernet",
      "status": "active"
    }
  ]
}
```

### Get Topology as GeoJSON

**Endpoint:** `GET /api/v1/topology/geojson`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {
        "type": "Point",
        "coordinates": [-74.0060, 40.7128]
      },
      "properties": {
        "type": "router",
        "id": "uuid",
        "name": "Router-01",
        "status": "active"
      }
    }
  ]
}
```

## Metrics

### Get Interface Metrics

**Endpoint:** `GET /api/v1/metrics/interfaces/{id}`

**Query Parameters:**
- `from` (ISO 8601 timestamp, default: 1 hour ago)
- `to` (ISO 8601 timestamp, default: now)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "interface_id": "uuid",
  "interface_name": "GigabitEthernet0/0",
  "in_bps": [],
  "out_bps": [],
  "utilization": []
}
```

### Get Router Metrics

**Endpoint:** `GET /api/v1/metrics/routers/{id}`

**Query Parameters:**
- `from` (ISO 8601 timestamp, default: 1 hour ago)
- `to` (ISO 8601 timestamp, default: now)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

## Alerts

### List Alerts

**Endpoint:** `GET /api/v1/alerts`

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 20)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Acknowledge Alert

**Endpoint:** `POST /api/v1/alerts/{id}/acknowledge`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "note": "Investigating the issue"
}
```

**Response:** `204 No Content`

## Users

### List Users

**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 20)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Get Current User

**Endpoint:** `GET /api/v1/users/me`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Get User

**Endpoint:** `GET /api/v1/users/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Update User

**Endpoint:** `PUT /api/v1/users/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith"
}
```

**Response:** `200 OK`

## Tenants (Admin Only)

### List Tenants

**Endpoint:** `GET /api/v1/tenants`

**Query Parameters:**
- `page` (default: 1)
- `page_size` (default: 20)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Create Tenant

**Endpoint:** `POST /api/v1/tenants`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "name": "New ISP",
  "slug": "new-isp",
  "contact_email": "admin@newisp.com",
  "subscription_tier": "professional",
  "max_devices": 100,
  "max_users": 20
}
```

**Response:** `201 Created`

### Get Tenant

**Endpoint:** `GET /api/v1/tenants/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

### Update Tenant

**Endpoint:** `PUT /api/v1/tenants/{id}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "max_devices": 200,
  "status": "active"
}
```

**Response:** `200 OK`

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "code": "BAD_REQUEST",
  "message": "Invalid request",
  "details": "Additional error information"
}
```

### 401 Unauthorized
```json
{
  "code": "UNAUTHORIZED",
  "message": "Authentication required"
}
```

### 403 Forbidden
```json
{
  "code": "FORBIDDEN",
  "message": "Permission denied"
}
```

### 404 Not Found
```json
{
  "code": "NOT_FOUND",
  "message": "Resource not found"
}
```

### 409 Conflict
```json
{
  "code": "CONFLICT",
  "message": "Resource already exists"
}
```

### 422 Validation Error
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": [
    {
      "field": "email",
      "message": "email must be a valid email address"
    }
  ]
}
```

### 500 Internal Server Error
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Internal server error"
}
```

## Rate Limiting

API requests are rate-limited to prevent abuse. The default limit is 100 requests per minute per IP address.

## Multi-Tenant Isolation

All API endpoints enforce tenant isolation. Users can only access resources belonging to their tenant. This is enforced at both the middleware and repository levels.

## Security

- All endpoints (except `/health`, `/auth/login`, and `/auth/register`) require JWT authentication
- Passwords are hashed using bcrypt with cost factor 12
- JWT tokens expire after 15 minutes (configurable)
- Refresh tokens expire after 7 days (configurable)
- Tokens can be revoked using the logout endpoint
- CORS is enabled and configurable via environment variables
