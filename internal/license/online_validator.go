package license

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Common errors
var (
	ErrLicenseNotFound     = errors.New("license not found")
	ErrLicenseExpired      = errors.New("license has expired")
	ErrLicenseInvalid      = errors.New("license is invalid")
	ErrServerUnavailable   = errors.New("license server unavailable")
	ErrValidationFailed    = errors.New("license validation failed")
	ErrOfflineGraceExpired = errors.New("offline grace period expired")
)

// OnlineValidator validates licenses against the license server
type OnlineValidator struct {
	serverURL     string
	httpClient    *http.Client
	cachedInfo    *LicenseInfo
	cacheMutex    sync.RWMutex
	lastValidated time.Time
	offlineGrace  time.Duration // Grace period when server unreachable
	licenseKey    string
}

// OnlineValidatorConfig configuration for the online validator
type OnlineValidatorConfig struct {
	ServerURL       string
	LicenseKey      string
	OfflineGrace    time.Duration
	RequestTimeout  time.Duration
	ValidationCache time.Duration
}

// NewOnlineValidator creates a new online license validator
func NewOnlineValidator(cfg OnlineValidatorConfig) *OnlineValidator {
	if cfg.OfflineGrace == 0 {
		cfg.OfflineGrace = 72 * time.Hour // Default 72 hour grace period
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 30 * time.Second
	}

	return &OnlineValidator{
		serverURL:    cfg.ServerURL,
		licenseKey:   cfg.LicenseKey,
		offlineGrace: cfg.OfflineGrace,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

// Validate validates a license key against the license server
func (v *OnlineValidator) Validate(ctx context.Context, licenseKey string) (*LicenseInfo, error) {
	// Try to validate with the license server
	info, err := v.validateOnline(ctx, licenseKey)
	if err == nil {
		v.updateCache(info)
		return info, nil
	}

	// If online validation fails, check if we have a valid cached license
	if v.hasCachedLicense() {
		// Check if we're within the offline grace period
		if time.Since(v.lastValidated) < v.offlineGrace {
			v.cacheMutex.RLock()
			defer v.cacheMutex.RUnlock()
			return v.cachedInfo, nil
		}
		return nil, ErrOfflineGraceExpired
	}

	return nil, err
}

// validateOnline performs the actual online validation
func (v *OnlineValidator) validateOnline(ctx context.Context, licenseKey string) (*LicenseInfo, error) {
	url := fmt.Sprintf("%s/validate", v.serverURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", licenseKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServerUnavailable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info LicenseInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Check if expired
		if info.IsExpired() {
			return nil, ErrLicenseExpired
		}

		return &info, nil

	case http.StatusNotFound:
		return nil, ErrLicenseNotFound

	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, ErrLicenseInvalid

	default:
		return nil, fmt.Errorf("%w: status code %d", ErrValidationFailed, resp.StatusCode)
	}
}

// updateCache updates the cached license information
func (v *OnlineValidator) updateCache(info *LicenseInfo) {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()
	v.cachedInfo = info
	v.lastValidated = time.Now()
}

// hasCachedLicense returns true if there is a cached license
func (v *OnlineValidator) hasCachedLicense() bool {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()
	return v.cachedInfo != nil && !v.cachedInfo.IsExpired()
}

// GetFeatures returns the features enabled by the current license
func (v *OnlineValidator) GetFeatures(ctx context.Context) (*Features, error) {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()

	if v.cachedInfo == nil {
		return nil, ErrLicenseNotFound
	}

	return &v.cachedInfo.Features, nil
}

// IsValid returns true if the current license is valid
func (v *OnlineValidator) IsValid() bool {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()

	if v.cachedInfo == nil {
		return false
	}

	// Check if expired
	if v.cachedInfo.IsExpired() {
		return false
	}

	// Check if within offline grace period
	if time.Since(v.lastValidated) > v.offlineGrace {
		return false
	}

	return true
}

// GetInfo returns the current license information
func (v *OnlineValidator) GetInfo() *LicenseInfo {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()
	return v.cachedInfo
}

// StartBackgroundValidation starts periodic background validation
func (v *OnlineValidator) StartBackgroundValidation(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = v.Validate(ctx, v.licenseKey)
			}
		}
	}()
}
