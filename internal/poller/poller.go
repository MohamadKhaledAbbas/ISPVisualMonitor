package poller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/database"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// Service handles router polling operations
type Service struct {
	db     *database.DB
	config config.PollerConfig

	// Channels for work distribution
	jobs    chan *models.Router
	results chan *PollResult

	// Worker management
	wg sync.WaitGroup
}

// PollResult represents the result of a polling operation
type PollResult struct {
	RouterID  uuid.UUID
	Success   bool
	Error     error
	Timestamp time.Time
	Metrics   map[string]interface{}
}

// NewService creates a new poller service
func NewService(db *database.DB, cfg config.PollerConfig) *Service {
	return &Service{
		db:      db,
		config:  cfg,
		jobs:    make(chan *models.Router, cfg.ConcurrentPolls),
		results: make(chan *PollResult, cfg.ConcurrentPolls),
	}
}

// Start begins the polling service
func (s *Service) Start(ctx context.Context) error {
	log.Printf("Starting poller service with %d workers", s.config.WorkerCount)

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
	log.Println("Poller service shutting down...")

	// Close channels
	close(s.jobs)

	// Wait for workers to finish
	s.wg.Wait()
	close(s.results)

	log.Println("Poller service stopped")
	return nil
}

// scheduler periodically fetches routers that need polling
func (s *Service) scheduler(ctx context.Context) {
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
func (s *Service) fetchAndScheduleRouters() {
	// Query routers that need polling
	query := `
		SELECT id, tenant_id, name, management_ip, snmp_version, snmp_community, 
		       snmp_port, polling_interval_seconds, last_polled_at
		FROM routers
		WHERE polling_enabled = true
		  AND status = 'active'
		  AND (last_polled_at IS NULL 
		       OR last_polled_at < NOW() - (polling_interval_seconds || ' seconds')::INTERVAL)
		ORDER BY last_polled_at NULLS FIRST
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
		router := &models.Router{}
		var lastPolled *time.Time

		err := rows.Scan(
			&router.ID,
			&router.TenantID,
			&router.Name,
			&router.ManagementIP,
			&router.SNMPVersion,
			&router.SNMPCommunity,
			&router.SNMPPort,
			&router.PollingIntervalSeconds,
			&lastPolled,
		)

		if err != nil {
			log.Printf("Error scanning router: %v", err)
			continue
		}

		router.LastPolledAt = lastPolled

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

// worker processes polling jobs
func (s *Service) worker(ctx context.Context, id int) {
	defer s.wg.Done()

	log.Printf("Poller worker %d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Poller worker %d stopping", id)
			return
		case router, ok := <-s.jobs:
			if !ok {
				log.Printf("Poller worker %d stopped (channel closed)", id)
				return
			}

			result := s.pollRouter(router)

			// Send result (non-blocking)
			select {
			case s.results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// pollRouter performs the actual SNMP polling of a router
func (s *Service) pollRouter(router *models.Router) *PollResult {
	log.Printf("Polling router: %s (%s)", router.Name, router.ManagementIP)

	// TODO: Implement actual SNMP polling
	// For now, simulate a successful poll
	time.Sleep(100 * time.Millisecond)

	return &PollResult{
		RouterID:  router.ID,
		Success:   true,
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"cpu_percent":    45.5,
			"memory_percent": 62.3,
			"uptime_seconds": 86400,
		},
	}
}

// resultProcessor handles polling results
func (s *Service) resultProcessor(ctx context.Context) {
	defer s.wg.Done()

	log.Println("Result processor started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Result processor stopping")
			return
		case result, ok := <-s.results:
			if !ok {
				log.Println("Result processor stopped (channel closed)")
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
func (s *Service) handleSuccessfulPoll(result *PollResult) {
	// Update last_polled_at timestamp
	_, err := s.db.Exec(
		"UPDATE routers SET last_polled_at = $1 WHERE id = $2",
		result.Timestamp,
		result.RouterID,
	)

	if err != nil {
		log.Printf("Error updating router poll timestamp: %v", err)
	}

	// TODO: Store metrics in database
	// INSERT INTO router_metrics (router_id, timestamp, cpu_percent, memory_percent, uptime_seconds)

	log.Printf("Successfully polled router %s", result.RouterID)
}

// handleFailedPoll processes a failed polling result
func (s *Service) handleFailedPoll(result *PollResult) {
	log.Printf("Failed to poll router %s: %v", result.RouterID, result.Error)

	// TODO: Create alert for failed polling
	// TODO: Update router status if consecutive failures exceed threshold
}

// PollNow triggers immediate polling of a specific router
func (s *Service) PollNow(routerID uuid.UUID) error {
	// Query router details
	router := &models.Router{}
	err := s.db.QueryRow(
		"SELECT id, tenant_id, name, management_ip FROM routers WHERE id = $1",
		routerID,
	).Scan(&router.ID, &router.TenantID, &router.Name, &router.ManagementIP)

	if err != nil {
		return fmt.Errorf("router not found: %w", err)
	}

	// Send to job channel
	s.jobs <- router

	return nil
}
