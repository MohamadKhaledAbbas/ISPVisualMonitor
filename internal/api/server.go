package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/handlers"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/utils"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/auth"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/database"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository/postgres"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
)

// Server represents the API server
type Server struct {
	db           *database.DB
	config       config.APIConfig
	authConfig   config.AuthConfig
	router       *mux.Router
	authProvider auth.AuthProvider
	logger       *zap.Logger

	// Handlers
	authHandler      *handlers.AuthHandler
	routerHandler    *handlers.RouterHandler
	interfaceHandler *handlers.InterfaceHandler
	topologyHandler  *handlers.TopologyHandler
	metricsHandler   *handlers.MetricsHandler
	alertHandler     *handlers.AlertHandler
	userHandler      *handlers.UserHandler
	tenantHandler    *handlers.TenantHandler
}

// NewServer creates a new API server instance
func NewServer(db *database.DB, apiCfg config.APIConfig, authCfg config.AuthConfig) *Server {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Create auth provider
	authProvider, err := auth.NewAuthProvider(&authCfg)
	if err != nil {
		log.Fatalf("Failed to create auth provider: %v", err)
	}

	// Create repositories
	userRepo := postgres.NewUserRepo(db.DB)
	tenantRepo := postgres.NewTenantRepo(db.DB)
	routerRepo := postgres.NewRouterRepo(db.DB)
	interfaceRepo := postgres.NewInterfaceRepo(db.DB)
	linkRepo := postgres.NewLinkRepo(db.DB)
	alertRepo := postgres.NewAlertRepo(db.DB)

	// Create services
	authService := service.NewAuthService(userRepo, tenantRepo, authProvider, logger)
	routerService := service.NewRouterService(routerRepo, logger)
	interfaceService := service.NewInterfaceService(interfaceRepo, routerRepo, logger)
	topologyService := service.NewTopologyService(routerRepo, interfaceRepo, linkRepo, logger)
	metricsService := service.NewMetricsService(interfaceRepo, routerRepo, logger)
	alertService := service.NewAlertService(alertRepo, logger)
	userService := service.NewUserService(userRepo, logger)
	tenantService := service.NewTenantService(tenantRepo, logger)

	// Create validator
	validatorInstance := utils.NewValidator()

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService, validatorInstance.Validator())
	routerHandler := handlers.NewRouterHandler(routerService, validatorInstance.Validator())
	interfaceHandler := handlers.NewInterfaceHandler(interfaceService, validatorInstance.Validator())
	topologyHandler := handlers.NewTopologyHandler(topologyService, validatorInstance.Validator())
	metricsHandler := handlers.NewMetricsHandler(metricsService, validatorInstance.Validator())
	alertHandler := handlers.NewAlertHandler(alertService, validatorInstance.Validator())
	userHandler := handlers.NewUserHandler(userService, validatorInstance.Validator())
	tenantHandler := handlers.NewTenantHandler(tenantService, validatorInstance.Validator())

	s := &Server{
		db:               db,
		config:           apiCfg,
		authConfig:       authCfg,
		router:           mux.NewRouter(),
		authProvider:     authProvider,
		logger:           logger,
		authHandler:      authHandler,
		routerHandler:    routerHandler,
		interfaceHandler: interfaceHandler,
		topologyHandler:  topologyHandler,
		metricsHandler:   metricsHandler,
		alertHandler:     alertHandler,
		userHandler:      userHandler,
		tenantHandler:    tenantHandler,
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

	// Root endpoint (no auth required)
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")

	// Health check endpoint (no auth required)
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
	s.router.HandleFunc("/ready", s.handleReadinessCheck).Methods("GET")
	s.router.HandleFunc("/live", s.handleLivenessCheck).Methods("GET")
	s.router.HandleFunc("/api/v1/health", s.handleHealthCheck).Methods("GET")
	s.router.HandleFunc("/api/v1/ready", s.handleReadinessCheck).Methods("GET")
	s.router.HandleFunc("/api/v1/live", s.handleLivenessCheck).Methods("GET")

	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Auth routes (no auth middleware)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", s.authHandler.HandleLogin).Methods("POST")
	auth.HandleFunc("/register", s.authHandler.HandleRegister).Methods("POST")
	auth.HandleFunc("/refresh", s.authHandler.HandleRefresh).Methods("POST")

	// Protected routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(s.authProvider))
	protected.Use(middleware.TenantContext)

	// Auth logout (requires authentication)
	protected.HandleFunc("/auth/logout", s.authHandler.HandleLogout).Methods("POST")

	// Router endpoints
	protected.HandleFunc("/routers", s.routerHandler.HandleListRouters).Methods("GET")
	protected.HandleFunc("/routers", s.routerHandler.HandleCreateRouter).Methods("POST")
	protected.HandleFunc("/routers/{id}", s.routerHandler.HandleGetRouter).Methods("GET")
	protected.HandleFunc("/routers/{id}", s.routerHandler.HandleUpdateRouter).Methods("PUT")
	protected.HandleFunc("/routers/{id}", s.routerHandler.HandleDeleteRouter).Methods("DELETE")

	// Interface endpoints
	protected.HandleFunc("/interfaces", s.interfaceHandler.HandleListInterfaces).Methods("GET")
	protected.HandleFunc("/routers/{router_id}/interfaces", s.interfaceHandler.HandleListRouterInterfaces).Methods("GET")

	// Topology endpoints
	protected.HandleFunc("/topology", s.topologyHandler.HandleGetTopology).Methods("GET")
	protected.HandleFunc("/topology/geojson", s.topologyHandler.HandleGetTopologyGeoJSON).Methods("GET")

	// Metrics endpoints
	protected.HandleFunc("/metrics/interfaces/{id}", s.metricsHandler.HandleGetInterfaceMetrics).Methods("GET")
	protected.HandleFunc("/metrics/routers/{id}", s.metricsHandler.HandleGetRouterMetrics).Methods("GET")

	// Alert endpoints
	protected.HandleFunc("/alerts", s.alertHandler.HandleListAlerts).Methods("GET")
	protected.HandleFunc("/alerts/{id}/acknowledge", s.alertHandler.HandleAcknowledgeAlert).Methods("POST")

	// User endpoints
	protected.HandleFunc("/users", s.userHandler.HandleListUsers).Methods("GET")
	protected.HandleFunc("/users/me", s.userHandler.HandleGetMe).Methods("GET")
	protected.HandleFunc("/users/{id}", s.userHandler.HandleGetUser).Methods("GET")
	protected.HandleFunc("/users/{id}", s.userHandler.HandleUpdateUser).Methods("PUT")

	// Tenant endpoints (admin only - TODO: add admin middleware)
	protected.HandleFunc("/tenants", s.tenantHandler.HandleListTenants).Methods("GET")
	protected.HandleFunc("/tenants", s.tenantHandler.HandleCreateTenant).Methods("POST")
	protected.HandleFunc("/tenants/{id}", s.tenantHandler.HandleGetTenant).Methods("GET")
	protected.HandleFunc("/tenants/{id}", s.tenantHandler.HandleUpdateTenant).Methods("PUT")
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// handleRoot handles requests to the root path
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "ISP Visual Monitor API",
		"version": "v1",
		"status":  "running",
		"docs":    "https://github.com/MohamadKhaledAbbas/ISPVisualMonitor",
		"endpoints": map[string]interface{}{
			"health": map[string]string{
				"GET /health":        "Health check (no auth required)",
				"GET /ready":         "Readiness probe (checks dependencies)",
				"GET /live":          "Liveness probe (process alive)",
				"GET /api/v1/health": "Health check v1 (no auth required)",
				"GET /api/v1/ready":  "Readiness probe v1 (checks dependencies)",
				"GET /api/v1/live":   "Liveness probe v1 (process alive)",
			},
			"auth": map[string]string{
				"POST /api/v1/auth/login":    "User login (no auth required)",
				"POST /api/v1/auth/register": "User registration (no auth required)",
				"POST /api/v1/auth/refresh":  "Token refresh (no auth required)",
				"POST /api/v1/auth/logout":   "User logout (auth required)",
			},
			"routers": map[string]string{
				"GET /api/v1/routers":         "List all routers (auth required)",
				"POST /api/v1/routers":        "Create new router (auth required)",
				"GET /api/v1/routers/{id}":    "Get router details (auth required)",
				"PUT /api/v1/routers/{id}":    "Update router (auth required)",
				"DELETE /api/v1/routers/{id}": "Delete router (auth required)",
			},
			"topology": map[string]string{
				"GET /api/v1/topology":         "Get network topology (auth required)",
				"GET /api/v1/topology/geojson": "Get topology as GeoJSON (auth required)",
			},
			"metrics": map[string]string{
				"GET /api/v1/metrics/interfaces/{id}": "Get interface metrics (auth required)",
				"GET /api/v1/metrics/routers/{id}":    "Get router metrics (auth required)",
			},
		},
		"note": "Most endpoints require JWT authentication. Use /api/v1/auth/login to get a token.",
	})
}

// handleHealthCheck handles health check requests
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.db.Ping(); err != nil {
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "isp-visual-monitor",
	})
}

// handleReadinessCheck handles readiness probe requests
func (s *Server) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(); err != nil {
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"status":  "ready",
		"service": "isp-visual-monitor",
	})
}

// handleLivenessCheck handles liveness probe requests
func (s *Server) handleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"status":  "alive",
		"service": "isp-visual-monitor",
	})
}
