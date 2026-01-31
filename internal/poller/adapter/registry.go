package adapter

import (
	"context"
	"fmt"
	"log"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
)

// Registry manages all available polling adapters
type Registry struct {
	adapters []PollerAdapter
	config   AdapterConfig
}

// NewRegistry creates a new adapter registry
func NewRegistry(config AdapterConfig) *Registry {
	r := &Registry{
		adapters: []PollerAdapter{},
		config:   config,
	}

	// Register default adapters in priority order
	// MikroTik API adapter (highest priority for MikroTik devices)
	r.Register(NewMikroTikAdapter(config))

	// SNMP adapter (generic fallback)
	r.Register(NewSNMPAdapter(config))

	return r
}

// Register adds an adapter to the registry
func (r *Registry) Register(adapter PollerAdapter) {
	r.adapters = append(r.adapters, adapter)
	log.Printf("Registered adapter: %s", adapter.GetAdapterName())
}

// GetAdapter returns the best adapter for a given router
func (r *Registry) GetAdapter(router *models.EnhancedRouter) (PollerAdapter, error) {
	// First, try to use the preferred method from router capabilities
	if router.Capabilities != nil && router.Capabilities.PreferredMethod != "" {
		for _, adapter := range r.adapters {
			if adapter.GetAdapterName() == router.Capabilities.PreferredMethod && adapter.CanHandle(router) {
				return adapter, nil
			}
		}
	}

	// If preferred method doesn't work, try adapters in order
	for _, adapter := range r.adapters {
		if adapter.CanHandle(router) {
			return adapter, nil
		}
	}

	return nil, fmt.Errorf("no suitable adapter found for router %s", router.Name)
}

// GetAdapterWithFallback attempts to get an adapter with fallback logic
func (r *Registry) GetAdapterWithFallback(router *models.EnhancedRouter) []PollerAdapter {
	suitableAdapters := []PollerAdapter{}

	// Get fallback order from router capabilities
	fallbackOrder := []string{}
	if router.Capabilities != nil {
		fallbackOrder = router.Capabilities.GetPreferredMethodOrder()
	}

	// Build list of adapters in fallback order
	for _, methodName := range fallbackOrder {
		for _, adapter := range r.adapters {
			if adapter.GetAdapterName() == methodName && adapter.CanHandle(router) {
				suitableAdapters = append(suitableAdapters, adapter)
				break
			}
		}
	}

	// If no fallback order or nothing matched, add all compatible adapters
	if len(suitableAdapters) == 0 {
		for _, adapter := range r.adapters {
			if adapter.CanHandle(router) {
				suitableAdapters = append(suitableAdapters, adapter)
			}
		}
	}

	return suitableAdapters
}

// PollWithFallback attempts to poll a router using multiple adapters if needed
func (r *Registry) PollWithFallback(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
	adapters := r.GetAdapterWithFallback(router)

	if len(adapters) == 0 {
		return nil, fmt.Errorf("no suitable adapter found for router %s", router.Name)
	}

	var lastErr error

	// Try each adapter in order
	for _, adapter := range adapters {
		log.Printf("Attempting to poll router %s with adapter %s", router.Name, adapter.GetAdapterName())

		result, err := adapter.Poll(ctx, router)
		if err == nil && result.Success {
			log.Printf("Successfully polled router %s with adapter %s", router.Name, adapter.GetAdapterName())
			return result, nil
		}

		// Log failure and try next adapter
		log.Printf("Failed to poll router %s with adapter %s: %v", router.Name, adapter.GetAdapterName(), err)
		lastErr = err
	}

	// All adapters failed
	return nil, fmt.Errorf("all adapters failed for router %s, last error: %v", router.Name, lastErr)
}

// HealthCheckWithFallback checks router health using the first working adapter
func (r *Registry) HealthCheckWithFallback(ctx context.Context, router *models.EnhancedRouter) error {
	adapters := r.GetAdapterWithFallback(router)

	if len(adapters) == 0 {
		return fmt.Errorf("no suitable adapter found for router %s", router.Name)
	}

	// Try each adapter
	for _, adapter := range adapters {
		err := adapter.HealthCheck(ctx, router)
		if err == nil {
			return nil
		}
		log.Printf("Health check failed for router %s with adapter %s: %v", router.Name, adapter.GetAdapterName(), err)
	}

	return fmt.Errorf("all health checks failed for router %s", router.Name)
}

// ListAdapters returns all registered adapters
func (r *Registry) ListAdapters() []string {
	names := make([]string, len(r.adapters))
	for i, adapter := range r.adapters {
		names[i] = adapter.GetAdapterName()
	}
	return names
}

// GetAdapterByName returns an adapter by name
func (r *Registry) GetAdapterByName(name string) (PollerAdapter, error) {
	for _, adapter := range r.adapters {
		if adapter.GetAdapterName() == name {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("adapter not found: %s", name)
}

// GetSupportedMetrics returns all metrics supported by any adapter for a router
func (r *Registry) GetSupportedMetrics(router *models.EnhancedRouter) []string {
	metricsMap := make(map[string]bool)

	for _, adapter := range r.adapters {
		if adapter.CanHandle(router) {
			for _, metric := range adapter.GetSupportedMetrics() {
				metricsMap[metric] = true
			}
		}
	}

	metrics := make([]string, 0, len(metricsMap))
	for metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	return metrics
}
