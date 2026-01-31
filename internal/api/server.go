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

	// Health check endpoint (no auth required)
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
	s.router.HandleFunc("/api/v1/health", s.handleHealthCheck).Methods("GET")

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
