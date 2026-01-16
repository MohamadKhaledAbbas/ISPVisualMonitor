package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// Logger middleware logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Generate request ID
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		
		// Log request
		log.Printf("[%s] %s %s - Started", requestID, r.Method, r.URL.Path)
		
		next.ServeHTTP(w, r.WithContext(ctx))
		
		// Log completion
		duration := time.Since(start)
		log.Printf("[%s] %s %s - Completed in %v", requestID, r.Method, r.URL.Path, duration)
	})
}

// CORS middleware handles CORS headers
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter middleware implements basic rate limiting
func RateLimiter(requestsPerMin int) func(http.Handler) http.Handler {
	// TODO: Implement proper rate limiting with Redis or in-memory store
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for rate limiting logic
			next.ServeHTTP(w, r)
		})
	}
}

// Auth middleware validates JWT tokens
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			
			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			
			token := parts[1]
			
			// TODO: Implement JWT validation
			// For now, just pass through
			_ = token
			_ = jwtSecret
			
			// Mock: Extract tenant_id and user_id from token claims
			// In production, parse JWT and extract claims
			mockTenantID := uuid.New()
			mockUserID := uuid.New()
			
			ctx := context.WithValue(r.Context(), TenantIDKey, mockTenantID)
			ctx = context.WithValue(ctx, UserIDKey, mockUserID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantContext middleware sets the database tenant context
func TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Context().Value(TenantIDKey)
		if tenantID == nil {
			http.Error(w, "Tenant context not found", http.StatusInternalServerError)
			return
		}
		
		// TODO: Set PostgreSQL session variable for RLS
		// db.Exec("SET app.current_tenant = $1", tenantID)
		
		next.ServeHTTP(w, r)
	})
}

// RequirePermission middleware checks if user has required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement permission check
			// 1. Get user ID from context
			// 2. Load user permissions from cache or database
			// 3. Check if permission exists
			// 4. Allow or deny request
			
			// For now, allow all requests
			next.ServeHTTP(w, r)
		})
	}
}
