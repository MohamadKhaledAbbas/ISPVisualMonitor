// Package metrics provides Prometheus metrics for ISP Visual Monitor.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP metrics
var (
	// HTTPRequestsTotal counts total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ispmonitor",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ispmonitor",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// HTTPRequestsInFlight tracks the number of in-flight requests
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "http_requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)
)

// Poller metrics
var (
	// ActivePollers tracks the number of active router pollers
	ActivePollers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "active_pollers",
			Help:      "Number of active router pollers",
		},
	)

	// RoutersPolled counts total routers polled
	RoutersPolled = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ispmonitor",
			Name:      "routers_polled_total",
			Help:      "Total number of routers polled",
		},
	)

	// PollDuration tracks the duration of polling operations
	PollDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ispmonitor",
			Name:      "poll_duration_seconds",
			Help:      "Duration of router polling operations in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"router_type", "status"},
	)

	// PollErrors counts polling errors
	PollErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ispmonitor",
			Name:      "poll_errors_total",
			Help:      "Total number of polling errors",
		},
		[]string{"router_type", "error_type"},
	)

	// QueuedPolls tracks the number of polls waiting in queue
	QueuedPolls = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "queued_polls",
			Help:      "Number of polls waiting in the queue",
		},
	)
)

// Router metrics
var (
	// RouterCount tracks the total number of routers
	RouterCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "routers_total",
			Help:      "Total number of routers by status",
		},
		[]string{"status"},
	)

	// InterfaceCount tracks the total number of interfaces
	InterfaceCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "interfaces_total",
			Help:      "Total number of interfaces by status",
		},
		[]string{"status"},
	)

	// RouterUptime tracks router uptime
	RouterUptime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "router_uptime_seconds",
			Help:      "Router uptime in seconds",
		},
		[]string{"router_id", "router_name"},
	)
)

// Database metrics
var (
	// DBConnectionsTotal tracks total database connections
	DBConnectionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "db_connections_total",
			Help:      "Total number of database connections",
		},
	)

	// DBConnectionsIdle tracks idle database connections
	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "db_connections_idle",
			Help:      "Number of idle database connections",
		},
	)

	// DBQueryDuration tracks database query duration
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ispmonitor",
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// License metrics
var (
	// LicenseValid indicates if the license is valid
	LicenseValid = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "license_valid",
			Help:      "Whether the license is valid (1) or not (0)",
		},
	)

	// LicenseDaysUntilExpiry tracks days until license expires
	LicenseDaysUntilExpiry = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "license_days_until_expiry",
			Help:      "Number of days until the license expires",
		},
	)

	// LicenseRouterLimit tracks the router limit in the license
	LicenseRouterLimit = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "license_router_limit",
			Help:      "Maximum number of routers allowed by the license (-1 for unlimited)",
		},
	)
)

// Alert metrics
var (
	// AlertsTotal counts total alerts by severity
	AlertsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ispmonitor",
			Name:      "alerts_total",
			Help:      "Total number of alerts by severity",
		},
		[]string{"severity"},
	)

	// AlertsActive tracks active alerts
	AlertsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "alerts_active",
			Help:      "Number of active alerts by severity",
		},
		[]string{"severity"},
	)
)

// Application info metric
var (
	// BuildInfo exposes build information
	BuildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ispmonitor",
			Name:      "build_info",
			Help:      "Build information about the ISP Monitor",
		},
		[]string{"version", "commit", "build_date", "go_version"},
	)
)

// RecordBuildInfo records the build information metric
func RecordBuildInfo(version, commit, buildDate, goVersion string) {
	BuildInfo.WithLabelValues(version, commit, buildDate, goVersion).Set(1)
}
