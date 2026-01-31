package auth

import (
	"context"
	"sync"
	"time"
)

// TokenBlacklist defines the interface for token revocation
// Implementations can be in-memory (dev) or Redis (production)
type TokenBlacklist interface {
	// Add adds a token to the blacklist with TTL
	Add(ctx context.Context, jti string, expiration time.Time) error

	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, jti string) (bool, error)

	// Cleanup removes expired tokens from the blacklist
	Cleanup(ctx context.Context) error
}

// InMemoryBlacklist implements TokenBlacklist using in-memory storage
// Suitable for development and single-instance deployments
type InMemoryBlacklist struct {
	mu      sync.RWMutex
	tokens  map[string]time.Time
}

// NewInMemoryBlacklist creates a new in-memory blacklist
func NewInMemoryBlacklist() *InMemoryBlacklist {
	bl := &InMemoryBlacklist{
		tokens: make(map[string]time.Time),
	}
	
	// Start cleanup goroutine
	go bl.cleanupLoop()
	
	return bl
}

// Add adds a token to the blacklist
func (b *InMemoryBlacklist) Add(ctx context.Context, jti string, expiration time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.tokens[jti] = expiration
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (b *InMemoryBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	expiration, exists := b.tokens[jti]
	if !exists {
		return false, nil
	}
	
	// Check if token has expired (should have been cleaned up)
	if time.Now().After(expiration) {
		return false, nil
	}
	
	return true, nil
}

// Cleanup removes expired tokens from the blacklist
func (b *InMemoryBlacklist) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	now := time.Now()
	for jti, expiration := range b.tokens {
		if now.After(expiration) {
			delete(b.tokens, jti)
		}
	}
	
	return nil
}

// cleanupLoop runs periodic cleanup of expired tokens
func (b *InMemoryBlacklist) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		_ = b.Cleanup(context.Background())
	}
}
