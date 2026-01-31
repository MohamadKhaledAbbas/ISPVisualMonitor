package poller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/database"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/poller/adapter"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// EnhancedService handles router polling using the adapter pattern
type EnhancedService struct {
	db       *database.DB
	config   config.PollerConfig
	registry *adapter.Registry

	// Channels for work distribution
	jobs    chan *models.EnhancedRouter
	results chan *adapter.PollResult

	// Worker management
	wg sync.WaitGroup
}

// NewEnhancedService creates a new enhanced poller service
func NewEnhancedService(db *database.DB, cfg config.PollerConfig) *EnhancedService {
	// Create adapter registry with configuration
	adapterConfig := adapter.AdapterConfig{
		TimeoutSeconds: cfg.TimeoutSeconds,
		RetryAttempts:  cfg.RetryAttempts,
		RetryDelay:     2 * time.Second,
	}

	return &EnhancedService{
		db:       db,
		config:   cfg,
		registry: adapter.NewRegistry(adapterConfig),
		jobs:     make(chan *models.EnhancedRouter, cfg.ConcurrentPolls),
		results:  make(chan *adapter.PollResult, cfg.ConcurrentPolls),
	}
}

// Start begins the enhanced polling service
func (s *EnhancedService) Start(ctx context.Context) error {
	log.Printf("Starting enhanced poller service with %d workers", s.config.WorkerCount)
	log.Printf("Registered adapters: %v", s.registry.ListAdapters())

	// Start worker goroutines
	for i := 0; i < s.config.WorkerCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i)
	}

	// Start result processor
	s.wg.Add(1)
	go s.resultProcessor(ctx)

	// Start job scheduler
	s.wg.Add(1)
	go s.scheduler(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Enhanced poller service shutting down...")

	// Close channels
	close(s.jobs)

	// Wait for workers to finish
	s.wg.Wait()
	close(s.results)

	log.Println("Enhanced poller service stopped")
	return nil
}

// scheduler periodically fetches routers that need polling
func (s *EnhancedService) scheduler(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run immediately on start
	s.fetchAndScheduleRouters()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fetchAndScheduleRouters()
		}
	}
}

// fetchAndScheduleRouters queries database for routers needing polling
func (s *EnhancedService) fetchAndScheduleRouters() {
	// Query routers with their capabilities
	query := `
		SELECT 
			r.id, r.tenant_id, r.name, r.management_ip, r.vendor, r.status,
			r.polling_enabled, r.polling_interval_seconds, r.last_polled_at,
			rc.preferred_method, rc.fallback_order
		FROM routers r
		LEFT JOIN router_capabilities rc ON r.id = rc.router_id
		WHERE r.polling_enabled = true
		  AND r.status = 'active'
		  AND (r.last_polled_at IS NULL 
		       OR r.last_polled_at < NOW() - (r.polling_interval_seconds || ' seconds')::INTERVAL)
		ORDER BY r.last_polled_at NULLS FIRST
		LIMIT $1
	`

	rows, err := s.db.Query(query, s.config.ConcurrentPolls)
	if err != nil {
		log.Printf("Error querying routers: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		router := &models.EnhancedRouter{}
		router.Capabilities = &models.RouterCapabilities{}

		var lastPolled *time.Time
		var preferredMethod, fallbackOrder *string

		err := rows.Scan(
			&router.ID,
			&router.TenantID,
			&router.Name,
			&router.ManagementIP,
			&router.Vendor,
			&router.Status,
			&router.PollingEnabled,
			&router.PollingIntervalSeconds,
			&lastPolled,
			&preferredMethod,
			&fallbackOrder,
		)

		if err != nil {
			log.Printf("Error scanning router: %v", err)
			continue
		}

		router.LastPolledAt = lastPolled

		if preferredMethod != nil {
			router.Capabilities.PreferredMethod = *preferredMethod
		}

		// Load full capabilities for this router
		s.loadRouterCapabilities(router)

		// Load router roles
		s.loadRouterRoles(router)

		// Send to job channel (non-blocking)
		select {
		case s.jobs <- router:
			count++
		default:
			// Channel full, skip this router
			log.Printf("Job queue full, skipping router %s", router.Name)
		}
	}

	if count > 0 {
		log.Printf("Scheduled %d routers for polling", count)
	}
}

// loadRouterCapabilities loads full capabilities for a router
func (s *EnhancedService) loadRouterCapabilities(router *models.EnhancedRouter) {
	query := `
		SELECT 
			snmp_enabled, snmp_version, snmp_community, snmp_port, snmp_timeout_seconds, snmp_retries,
			api_enabled, api_type, api_endpoint, api_port, api_username, api_password, 
			api_use_tls, api_verify_cert, api_timeout_seconds,
			ssh_enabled, ssh_host, ssh_port, ssh_username, ssh_password, ssh_timeout_seconds,
			preferred_method, fallback_order
		FROM router_capabilities
		WHERE router_id = $1
	`

	capabilities := &models.RouterCapabilities{
		SNMP: &models.SNMPCapability{},
		API:  &models.APICapability{},
		SSH:  &models.SSHCapability{},
	}

	var fallbackOrder *string

	err := s.db.QueryRow(query, router.ID).Scan(
		&capabilities.SNMP.Enabled,
		&capabilities.SNMP.Version,
		&capabilities.SNMP.Community,
		&capabilities.SNMP.Port,
		&capabilities.SNMP.TimeoutSeconds,
		&capabilities.SNMP.Retries,
		&capabilities.API.Enabled,
		&capabilities.API.Type,
		&capabilities.API.Endpoint,
		&capabilities.API.Port,
		&capabilities.API.Username,
		&capabilities.API.Password,
		&capabilities.API.UseTLS,
		&capabilities.API.VerifyCert,
		&capabilities.API.TimeoutSeconds,
		&capabilities.SSH.Enabled,
		&capabilities.SSH.Host,
		&capabilities.SSH.Port,
		&capabilities.SSH.Username,
		&capabilities.SSH.Password,
		&capabilities.SSH.TimeoutSeconds,
		&capabilities.PreferredMethod,
		&fallbackOrder,
	)

	if err != nil {
		// If no capabilities found, router might be using legacy config
		log.Printf("No capabilities found for router %s, using legacy config", router.Name)
		return
	}

	router.Capabilities = capabilities
}

// loadRouterRoles loads roles for a router
func (s *EnhancedService) loadRouterRoles(router *models.EnhancedRouter) {
	query := `
		SELECT rra.id, rra.role_id, rra.priority, rra.is_primary,
		       rr.code, rr.name, rr.category
		FROM router_role_assignments rra
		JOIN router_roles rr ON rra.role_id = rr.id
		WHERE rra.router_id = $1
		ORDER BY rra.priority ASC
	`

	rows, err := s.db.Query(query, router.ID)
	if err != nil {
		log.Printf("Error loading roles for router %s: %v", router.Name, err)
		return
	}
	defer rows.Close()

	roles := []models.RouterRoleAssignment{}

	for rows.Next() {
		rra := models.RouterRoleAssignment{
			Role: &models.RouterRole{},
		}

		err := rows.Scan(
			&rra.ID,
			&rra.RoleID,
			&rra.Priority,
			&rra.IsPrimary,
			&rra.Role.Code,
			&rra.Role.Name,
			&rra.Role.Category,
		)

		if err != nil {
			log.Printf("Error scanning role: %v", err)
			continue
		}

		roles = append(roles, rra)
	}

	router.Roles = roles
}

// worker processes polling jobs
func (s *EnhancedService) worker(ctx context.Context, id int) {
	defer s.wg.Done()

	log.Printf("Enhanced poller worker %d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Enhanced poller worker %d stopping", id)
			return
		case router, ok := <-s.jobs:
			if !ok {
				log.Printf("Enhanced poller worker %d stopped (channel closed)", id)
				return
			}

			result := s.pollRouter(ctx, router)

			// Send result (non-blocking)
			select {
			case s.results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// pollRouter performs the actual polling using the adapter registry
func (s *EnhancedService) pollRouter(ctx context.Context, router *models.EnhancedRouter) *adapter.PollResult {
	log.Printf("Polling router: %s (%s) with roles: %v",
		router.Name, router.ManagementIP, router.GetRoleCodes())

	// Record polling start
	pollStartTime := time.Now()

	// Use registry to poll with fallback
	result, err := s.registry.PollWithFallback(ctx, router)

	if err != nil {
		log.Printf("Failed to poll router %s: %v", router.Name, err)

		// Create error result
		result = adapter.NewPollResult(router.ID, router.TenantID, "none")
		result.Success = false
		result.ErrorMessage = err.Error()
		result.ResponseTimeMs = int(time.Since(pollStartTime).Milliseconds())
	}

	// Record polling in history
	s.recordPollingHistory(result, pollStartTime, time.Now())

	return result
}

// recordPollingHistory records polling attempt in the database
func (s *EnhancedService) recordPollingHistory(result *adapter.PollResult, startTime, endTime time.Time) {
	query := `
		INSERT INTO polling_history (
			tenant_id, router_id, poll_started_at, poll_completed_at,
			adapter_used, success, error_message, metrics_collected, response_time_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.Exec(
		query,
		result.TenantID,
		result.RouterID,
		startTime,
		endTime,
		result.AdapterUsed,
		result.Success,
		result.ErrorMessage,
		result.GetMetricsCount(),
		result.ResponseTimeMs,
	)

	if err != nil {
		log.Printf("Error recording polling history: %v", err)
	}
}

// resultProcessor handles polling results
func (s *EnhancedService) resultProcessor(ctx context.Context) {
	defer s.wg.Done()

	log.Println("Enhanced result processor started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Enhanced result processor stopping")
			return
		case result, ok := <-s.results:
			if !ok {
				log.Println("Enhanced result processor stopped (channel closed)")
				return
			}

			if result.Success {
				s.handleSuccessfulPoll(result)
			} else {
				s.handleFailedPoll(result)
			}
		}
	}
}

// handleSuccessfulPoll processes a successful polling result
func (s *EnhancedService) handleSuccessfulPoll(result *adapter.PollResult) {
	// Update last_polled_at timestamp
	_, err := s.db.Exec(
		"UPDATE routers SET last_polled_at = $1 WHERE id = $2",
		result.Timestamp,
		result.RouterID,
	)

	if err != nil {
		log.Printf("Error updating router poll timestamp: %v", err)
	}

	// Store router metrics
	s.storeRouterMetrics(result)

	// Store interface metrics
	s.storeInterfaceMetrics(result)

	// Store PPPoE sessions
	if len(result.PPPoESessions) > 0 {
		s.storePPPoESessions(result)
	}

	// Store NAT sessions
	if len(result.NATSessions) > 0 {
		s.storeNATSessions(result)
	}

	// Store DHCP leases
	if len(result.DHCPLeases) > 0 {
		s.storeDHCPLeases(result)
	}

	log.Printf("Successfully polled router %s with %d metrics",
		result.RouterID, result.GetMetricsCount())
}

// storeRouterMetrics stores router-level metrics
func (s *EnhancedService) storeRouterMetrics(result *adapter.PollResult) {
	query := `
		INSERT INTO router_metrics (
			tenant_id, router_id, timestamp, 
			cpu_percent, memory_percent, uptime_seconds, temperature_celsius
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	cpuPercent, _ := result.Metrics["cpu_percent"].(float64)
	memoryPercent, _ := result.Metrics["memory_percent"].(float64)
	uptimeSeconds, _ := result.Metrics["uptime_seconds"].(int64)
	temperatureCelsius, _ := result.Metrics["temperature_celsius"].(float64)

	_, err := s.db.Exec(
		query,
		result.TenantID,
		result.RouterID,
		result.Timestamp,
		cpuPercent,
		memoryPercent,
		uptimeSeconds,
		temperatureCelsius,
	)

	if err != nil {
		log.Printf("Error storing router metrics: %v", err)
	}
}

// storeInterfaceMetrics stores interface-level metrics
func (s *EnhancedService) storeInterfaceMetrics(result *adapter.PollResult) {
	// TODO: Implement interface metrics storage
	// This requires matching interfaces by name and storing stats
	log.Printf("TODO: Store %d interface metrics", len(result.Interfaces))
}

// storePPPoESessions stores PPPoE session data
func (s *EnhancedService) storePPPoESessions(result *adapter.PollResult) {
	// TODO: Implement PPPoE session storage
	// This should update existing sessions or insert new ones
	log.Printf("TODO: Store %d PPPoE sessions", len(result.PPPoESessions))
}

// storeNATSessions stores NAT session data
func (s *EnhancedService) storeNATSessions(result *adapter.PollResult) {
	// TODO: Implement NAT session storage
	log.Printf("TODO: Store %d NAT sessions", len(result.NATSessions))
}

// storeDHCPLeases stores DHCP lease data
func (s *EnhancedService) storeDHCPLeases(result *adapter.PollResult) {
	// TODO: Implement DHCP lease storage
	log.Printf("TODO: Store %d DHCP leases", len(result.DHCPLeases))
}

// handleFailedPoll processes a failed polling result
func (s *EnhancedService) handleFailedPoll(result *adapter.PollResult) {
	log.Printf("Failed to poll router %s: %s", result.RouterID, result.ErrorMessage)

	// TODO: Create alert for failed polling
	// TODO: Update router status if consecutive failures exceed threshold
}

// PollNow triggers immediate polling of a specific router
func (s *EnhancedService) PollNow(routerID uuid.UUID) error {
	// Query router details with capabilities
	router := &models.EnhancedRouter{}
	err := s.db.QueryRow(
		"SELECT id, tenant_id, name, management_ip FROM routers WHERE id = $1",
		routerID,
	).Scan(&router.ID, &router.TenantID, &router.Name, &router.ManagementIP)

	if err != nil {
		return fmt.Errorf("router not found: %w", err)
	}

	// Load capabilities and roles
	s.loadRouterCapabilities(router)
	s.loadRouterRoles(router)

	// Send to job channel
	s.jobs <- router

	return nil
}

// GetAdapterRegistry returns the adapter registry
func (s *EnhancedService) GetAdapterRegistry() *adapter.Registry {
	return s.registry
}
