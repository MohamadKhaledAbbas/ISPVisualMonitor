package license

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineCache provides offline caching capabilities for license validation
type OfflineCache struct {
	cachePath     string
	cachedInfo    *LicenseInfo
	cacheMutex    sync.RWMutex
	lastValidated time.Time
	offlineGrace  time.Duration
}

// CacheEntry represents a cached license entry
type CacheEntry struct {
	LicenseInfo   *LicenseInfo `json:"license_info"`
	LastValidated time.Time    `json:"last_validated"`
	Checksum      string       `json:"checksum"`
}

// NewOfflineCache creates a new offline license cache
func NewOfflineCache(cachePath string, offlineGrace time.Duration) *OfflineCache {
	if offlineGrace == 0 {
		offlineGrace = 72 * time.Hour
	}

	cache := &OfflineCache{
		cachePath:    cachePath,
		offlineGrace: offlineGrace,
	}

	// Load existing cache
	_ = cache.load()

	return cache
}

// Save saves license information to the cache
func (c *OfflineCache) Save(info *LicenseInfo) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cachedInfo = info
	c.lastValidated = time.Now()

	entry := CacheEntry{
		LicenseInfo:   info,
		LastValidated: c.lastValidated,
		Checksum:      c.computeChecksum(info),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(c.cachePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(c.cachePath, data, 0600)
}

// load loads the cached license from disk
func (c *OfflineCache) load() error {
	data, err := os.ReadFile(c.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}

	// Validate checksum
	if entry.Checksum != c.computeChecksum(entry.LicenseInfo) {
		return errors.New("cache checksum mismatch")
	}

	c.cachedInfo = entry.LicenseInfo
	c.lastValidated = entry.LastValidated

	return nil
}

// Get returns the cached license if valid
func (c *OfflineCache) Get() (*LicenseInfo, error) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	if c.cachedInfo == nil {
		return nil, ErrLicenseNotFound
	}

	// Check if license is expired
	if c.cachedInfo.IsExpired() {
		return nil, ErrLicenseExpired
	}

	// Check if within offline grace period
	if time.Since(c.lastValidated) > c.offlineGrace {
		return nil, ErrOfflineGraceExpired
	}

	return c.cachedInfo, nil
}

// IsValid returns true if there is a valid cached license
func (c *OfflineCache) IsValid() bool {
	_, err := c.Get()
	return err == nil
}

// Clear clears the offline cache
func (c *OfflineCache) Clear() error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cachedInfo = nil
	c.lastValidated = time.Time{}

	return os.Remove(c.cachePath)
}

// computeChecksum computes a SHA-256 checksum for the license info
func (c *OfflineCache) computeChecksum(info *LicenseInfo) string {
	if info == nil {
		return ""
	}
	data, err := json.Marshal(info)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CachedValidator wraps an online validator with offline caching
type CachedValidator struct {
	online *OnlineValidator
	cache  *OfflineCache
}

// NewCachedValidator creates a validator with offline caching support
func NewCachedValidator(online *OnlineValidator, cachePath string) *CachedValidator {
	return &CachedValidator{
		online: online,
		cache:  NewOfflineCache(cachePath, online.offlineGrace),
	}
}

// Validate validates the license, using cache as fallback
func (v *CachedValidator) Validate(ctx context.Context, licenseKey string) (*LicenseInfo, error) {
	// Try online validation first
	info, err := v.online.Validate(ctx, licenseKey)
	if err == nil {
		// Save to cache on successful validation
		_ = v.cache.Save(info)
		return info, nil
	}

	// Fall back to cache if online validation fails
	if cachedInfo, cacheErr := v.cache.Get(); cacheErr == nil {
		return cachedInfo, nil
	}

	return nil, err
}

// GetFeatures returns the features from the current license
func (v *CachedValidator) GetFeatures(ctx context.Context) (*Features, error) {
	return v.online.GetFeatures(ctx)
}

// IsValid returns true if the license is valid
func (v *CachedValidator) IsValid() bool {
	if v.online.IsValid() {
		return true
	}
	return v.cache.IsValid()
}

// GetInfo returns the current license information
func (v *CachedValidator) GetInfo() *LicenseInfo {
	if info := v.online.GetInfo(); info != nil {
		return info
	}
	info, _ := v.cache.Get()
	return info
}
