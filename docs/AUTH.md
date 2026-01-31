# Authentication Documentation

## Overview

ISP Visual Monitor implements production-strength JWT authentication with a pluggable provider pattern. This design allows for flexible authentication strategies ranging from simple local token management (ideal for development and small deployments) to enterprise-grade external identity providers like Keycloak, Auth0, or any OIDC-compliant system.

## Architecture

### Provider Pattern

The authentication system is built around the `AuthProvider` interface, which defines the contract for all authentication implementations:

```go
type AuthProvider interface {
    IssueToken(ctx context.Context, user *models.User, tenant *models.Tenant) (*TokenPair, error)
    ValidateToken(ctx context.Context, tokenString string) (*Claims, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    RevokeToken(ctx context.Context, tokenString string) error
    GetProviderName() string
}
```

This abstraction enables:
- **Easy testing** with mock providers
- **Gradual migration** from local to external providers
- **Multi-deployment scenarios** (dev, staging, prod with different providers)
- **Zero-downtime provider switching** via configuration

## JWT Claims Structure

### Standard Claims

The JWT tokens include standard claims as defined by RFC 7519:

- `jti` (JWT ID): Unique identifier for the token
- `sub` (Subject): User ID
- `iss` (Issuer): Identifies the token issuer (e.g., "ispvisualmonitor")
- `iat` (Issued At): Token creation timestamp
- `exp` (Expires At): Token expiration timestamp
- `nbf` (Not Before): Token validity start time

### Custom Claims

For multi-tenant RBAC support, we include custom claims:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "email": "user@example.com",
  "roles": ["admin", "user"],
  "permissions": ["routers:read", "routers:write", "alerts:read"],
  "token_type": "access",
  "jti": "...",
  "sub": "...",
  "iss": "ispvisualmonitor",
  "iat": 1640000000,
  "exp": 1640000900,
  "nbf": 1640000000
}
```

### Token Types

- **Access Token**: Short-lived token (default 15 minutes) for API authentication
- **Refresh Token**: Long-lived token (default 7 days) for obtaining new access tokens

## Authentication Providers

### Local Provider

The local provider implements JWT authentication using in-memory token management. It's ideal for:
- Development and testing
- Small single-instance deployments
- On-premise installations without external dependencies

**Features:**
- Supports both HS256 (symmetric) and RS256 (asymmetric) signing
- In-memory token blacklist for revocation
- Configurable token lifetimes
- Full multi-tenant support

**Limitations:**
- Token blacklist doesn't persist across restarts
- Not suitable for horizontally scaled deployments (use Redis blacklist in production)
- No federated identity management

### OIDC Provider (Future)

Support for OpenID Connect providers is planned for future releases. This will enable integration with:
- Keycloak
- Auth0
- Okta
- Azure AD
- Google Identity Platform
- Any OIDC-compliant provider

## Configuration

### Environment Variables

#### Core Settings

```bash
# Authentication provider (local, oidc, keycloak, auth0)
AUTH_PROVIDER=local

# JWT Secret (REQUIRED for HS256)
# Generate a strong secret: openssl rand -base64 64
JWT_SECRET=your-secret-key-here

# Signing method (HS256 or RS256)
JWT_SIGNING_METHOD=HS256

# JWT Issuer identifier
JWT_ISSUER=ispvisualmonitor

# Token Time-To-Live
ACCESS_TOKEN_TTL=15m      # 15 minutes
REFRESH_TOKEN_TTL=168h    # 7 days (168 hours)

# Password hashing cost (10-14 recommended)
BCRYPT_COST=12
```

#### RS256 Configuration (Optional)

For asymmetric signing with RS256:

```bash
JWT_SIGNING_METHOD=RS256
JWT_PRIVATE_KEY=/path/to/private-key.pem
JWT_PUBLIC_KEY=/path/to/public-key.pem
```

Generate RS256 keys:

```bash
# Generate private key
openssl genrsa -out private-key.pem 4096

# Extract public key
openssl rsa -in private-key.pem -pubout -out public-key.pem
```

#### OIDC Configuration (Future)

For external OIDC providers:

```bash
AUTH_PROVIDER=oidc
OIDC_ISSUER_URL=https://keycloak.example.com/realms/myrealm
OIDC_CLIENT_ID=ispvisualmonitor
OIDC_CLIENT_SECRET=your-client-secret
```

### Development vs Production

#### Development Setup (.env)

```bash
AUTH_PROVIDER=local
JWT_SECRET=dev-secret-not-for-production
JWT_SIGNING_METHOD=HS256
ACCESS_TOKEN_TTL=60m
REFRESH_TOKEN_TTL=720h
BCRYPT_COST=10
```

#### Production Setup (.env)

```bash
AUTH_PROVIDER=local
JWT_SECRET=<strong-random-secret-from-secrets-manager>
JWT_SIGNING_METHOD=RS256
JWT_PRIVATE_KEY=/run/secrets/jwt-private-key
JWT_PUBLIC_KEY=/run/secrets/jwt-public-key
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
BCRYPT_COST=12
```

Or with external OIDC:

```bash
AUTH_PROVIDER=keycloak
OIDC_ISSUER_URL=https://auth.company.com/realms/production
OIDC_CLIENT_ID=ispvisualmonitor-prod
OIDC_CLIENT_SECRET=<from-secrets-manager>
```

## Token Lifecycle

### 1. Authentication & Token Issuance

```
User -> Login Request -> API Server
                        -> Verify Credentials
                        -> AuthProvider.IssueToken()
                        -> Return Token Pair
```

### 2. API Request with Token

```
Client -> API Request + Bearer Token
       -> Middleware extracts token
       -> AuthProvider.ValidateToken()
       -> Inject claims into context
       -> Route to handler
```

### 3. Token Refresh

```
Client -> Refresh Request + Refresh Token
       -> AuthProvider.RefreshToken()
       -> Validate refresh token
       -> Issue new token pair
```

### 4. Logout (Token Revocation)

```
Client -> Logout Request + Token
       -> AuthProvider.RevokeToken()
       -> Add to blacklist
       -> Token becomes invalid
```

## Security Best Practices

### Token Management

1. **Always use HTTPS** in production
2. **Store tokens securely** on the client (httpOnly cookies preferred over localStorage)
3. **Implement token rotation** - issue new refresh token with each refresh
4. **Short access token lifetime** - 15 minutes or less
5. **Reasonable refresh token lifetime** - 7 days, force re-authentication after

### Secret Management

1. **Never commit secrets** to version control
2. **Use environment variables** or secrets management systems (Vault, AWS Secrets Manager)
3. **Rotate secrets regularly** - especially after team member changes
4. **Use strong random secrets** - minimum 256 bits of entropy
5. **Use RS256 in production** when possible for better security

### Password Security

1. **Enforce minimum password length** (8 characters minimum)
2. **Use bcrypt cost 12+** for production
3. **Consider additional requirements** (uppercase, lowercase, numbers, symbols)
4. **Implement rate limiting** on login endpoints
5. **Use timing-safe comparison** for password verification

### Multi-Tenant Isolation

1. **Always validate tenant ID** from token claims
2. **Use PostgreSQL Row-Level Security (RLS)** to enforce isolation at database level
3. **Log cross-tenant access attempts** for security auditing
4. **Implement tenant-scoped permissions**

## Usage Examples

### Issuing Tokens

```go
// Create auth provider
cfg := config.Load()
provider, err := auth.NewAuthProvider(&cfg.Auth)
if err != nil {
    log.Fatal(err)
}

// Issue token for user
ctx := context.Background()
tokenPair, err := provider.IssueToken(ctx, user, tenant)
if err != nil {
    return err
}

// Return to client
response := map[string]interface{}{
    "access_token":  tokenPair.AccessToken,
    "refresh_token": tokenPair.RefreshToken,
    "token_type":    tokenPair.TokenType,
    "expires_in":    tokenPair.ExpiresIn,
}
```

### Validating Tokens

```go
// In middleware
func Auth(provider auth.AuthProvider) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token
            token := extractBearerToken(r)
            
            // Validate
            claims, err := provider.ValidateToken(r.Context(), token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            // Inject into context
            ctx := context.WithValue(r.Context(), middleware.ClaimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Password Hashing

```go
// Hash password on registration
hashedPassword, err := auth.HashPassword(plainPassword, cfg.Auth.BcryptCost)
if err != nil {
    return err
}

// Store hashedPassword in database
user.PasswordHash = hashedPassword

// Verify password on login
err = auth.VerifyPassword(loginPassword, user.PasswordHash)
if err != nil {
    return auth.ErrInvalidCredentials
}
```

### Token Refresh

```go
// Client sends refresh token
refreshToken := r.Header.Get("X-Refresh-Token")

// Get new token pair
newTokenPair, err := provider.RefreshToken(ctx, refreshToken)
if err != nil {
    http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
    return
}

// Return new tokens
json.NewEncoder(w).Encode(newTokenPair)
```

### Token Revocation

```go
// On logout
token := extractBearerToken(r)

err := provider.RevokeToken(ctx, token)
if err != nil {
    log.Printf("Failed to revoke token: %v", err)
}

// Return success
w.WriteHeader(http.StatusNoContent)
```

## Testing

### Unit Tests

Run authentication tests:

```bash
go test -v ./internal/auth/...
```

Run with coverage:

```bash
go test -cover -coverprofile=coverage.out ./internal/auth/...
go tool cover -html=coverage.out
```

### Integration Tests

Test with real API endpoints:

```bash
# Start server
make run

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Use token
curl http://localhost:8080/api/v1/routers \
  -H "Authorization: Bearer <access-token>"

# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "X-Refresh-Token: <refresh-token>"

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access-token>"
```

## Troubleshooting

### Common Issues

#### "Invalid token" errors

- Check token expiration (`exp` claim)
- Verify `JWT_SECRET` matches between token issuance and validation
- Ensure `JWT_SIGNING_METHOD` is correct
- Check for token revocation (blacklist)

#### "Token has expired"

- Access tokens expire quickly by design (15 minutes)
- Use refresh token to obtain new access token
- Don't cache access tokens for long periods

#### "Token has been revoked"

- User logged out
- Token was explicitly revoked
- Check blacklist implementation

#### Performance issues with bcrypt

- Reduce `BCRYPT_COST` for development (minimum 10)
- Keep at 12+ for production
- Consider async password verification for high-traffic scenarios

### Debugging

Enable debug logging:

```bash
LOG_LEVEL=debug
```

Check token claims:

```bash
# Decode JWT (without verification)
echo "<token>" | cut -d. -f2 | base64 -d | jq
```

## Migration Guide

### Upgrading from Legacy Auth

The new auth system maintains backward compatibility with legacy tokens:

1. Both old and new middleware work during transition
2. Gradually migrate endpoints to use new provider-based middleware
3. Old tokens continue to work with compatibility layer
4. Issue new tokens using new provider after user re-authentication

### Switching Providers

To switch from local to OIDC:

1. Configure OIDC provider settings
2. Update `AUTH_PROVIDER` environment variable
3. Restart application
4. All new authentications use new provider
5. Existing tokens remain valid until expiration
6. Optional: Force re-authentication for immediate switch

## Future Enhancements

- [ ] OIDC provider implementation
- [ ] Redis-backed token blacklist for horizontal scaling
- [ ] Token rotation on refresh
- [ ] Refresh token family tracking
- [ ] OAuth2 social login integration
- [ ] MFA/2FA support
- [ ] Session management UI
- [ ] Token usage analytics
- [ ] Anomaly detection for token usage
- [ ] Automatic token cleanup for expired blacklist entries in Redis

## References

- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
