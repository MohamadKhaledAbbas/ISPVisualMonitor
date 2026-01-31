package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/auth"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/database"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
)

// Server represents the API server
type Server struct {
	db           *database.DB
	config       config.APIConfig
	authConfig   config.AuthConfig
	router       *mux.Router
	authProvider auth.AuthProvider
}

// NewServer creates a new API server instance
func NewServer(db *database.DB, apiCfg config.APIConfig, authCfg config.AuthConfig) *Server {
	// Create auth provider
	authProvider, err := auth.NewAuthProvider(&authCfg)
	if err != nil {
		log.Fatalf("Failed to create auth provider: %v", err)
	}

	s := &Server{
		db:           db,
		config:       apiCfg,
		authConfig:   authCfg,
		router:       mux.NewRouter(),
		authProvider: authProvider,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Apply global middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.CORS(s.config.AllowedOrigins))
	s.router.Use(middleware.RateLimiter(s.config.RateLimitPerMin))

	// Health check endpoint (no auth required)
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
	s.router.HandleFunc("/api/v1/health", s.handleHealthCheck).Methods("GET")

	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Auth routes (no auth middleware)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", s.handleLogin).Methods("POST")
	auth.HandleFunc("/register", s.handleRegister).Methods("POST")

	// Protected routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(s.authProvider))
	protected.Use(middleware.TenantContext)

	// Router endpoints
	protected.HandleFunc("/routers", s.handleListRouters).Methods("GET")
	protected.HandleFunc("/routers", s.handleCreateRouter).Methods("POST")
	protected.HandleFunc("/routers/{id}", s.handleGetRouter).Methods("GET")
	protected.HandleFunc("/routers/{id}", s.handleUpdateRouter).Methods("PUT")
	protected.HandleFunc("/routers/{id}", s.handleDeleteRouter).Methods("DELETE")

	// Interface endpoints
	protected.HandleFunc("/interfaces", s.handleListInterfaces).Methods("GET")
	protected.HandleFunc("/routers/{router_id}/interfaces", s.handleListRouterInterfaces).Methods("GET")

	// Topology endpoints
	protected.HandleFunc("/topology", s.handleGetTopology).Methods("GET")
	protected.HandleFunc("/topology/geojson", s.handleGetTopologyGeoJSON).Methods("GET")

	// Metrics endpoints
	protected.HandleFunc("/metrics/interfaces/{id}", s.handleGetInterfaceMetrics).Methods("GET")
	protected.HandleFunc("/metrics/routers/{id}", s.handleGetRouterMetrics).Methods("GET")

	// Alert endpoints
	protected.HandleFunc("/alerts", s.handleListAlerts).Methods("GET")
	protected.HandleFunc("/alerts/{id}/acknowledge", s.handleAcknowledgeAlert).Methods("POST")
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// handleHealthCheck handles health check requests
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.db.Ping(); err != nil {
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "isp-visual-monitor",
	})
}

// Placeholder handlers (to be implemented)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleListRouters(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleCreateRouter(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleGetRouter(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleUpdateRouter(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleDeleteRouter(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleListInterfaces(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleListRouterInterfaces(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleGetTopology(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleGetTopologyGeoJSON(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleGetInterfaceMetrics(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleGetRouterMetrics(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
