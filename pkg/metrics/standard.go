package metrics

import "time"

// Standard label names
const (
	LabelService     = "service"
	LabelInstance    = "instance"
	LabelEnvironment = "environment"
	LabelVersion     = "version"
	LabelEndpoint    = "endpoint"
	LabelMethod      = "method"
	LabelStatusCode  = "status_code"
	LabelErrorType   = "error_type"
	LabelComponent   = "component"
)

// Standard metric names
const (
	// HTTP metrics
	MetricHTTPRequestsTotal          = "http_requests_total"
	MetricHTTPRequestDurationSeconds = "http_request_duration_seconds"
	MetricHTTPRequestSizeBytes       = "http_request_size_bytes"
	MetricHTTPResponseSizeBytes      = "http_response_size_bytes"
	MetricHTTPRequestsInFlight       = "http_requests_in_flight"

	// Database metrics
	MetricDBConnectionsTotal     = "db_connections_total"
	MetricDBConnectionsInUse     = "db_connections_in_use"
	MetricDBQueryDurationSeconds = "db_query_duration_seconds"
	MetricDBErrorsTotal          = "db_errors_total"
	MetricDBTransactionsTotal    = "db_transactions_total"

	// Cache metrics
	MetricCacheHitsTotal                = "cache_hits_total"
	MetricCacheMissesTotal              = "cache_misses_total"
	MetricCacheItemsTotal               = "cache_items_total"
	MetricCacheSizeBytes                = "cache_size_bytes"
	MetricCacheOperationDurationSeconds = "cache_operation_duration_seconds"

	// Service metrics
	MetricServiceUptime       = "service_uptime_seconds"
	MetricServiceHealth       = "service_health_status"
	MetricServiceLastCheck    = "service_last_check_timestamp"
	MetricServiceDependencyUp = "service_dependency_up"

	// Resource metrics
	MetricCPUUsagePercent   = "cpu_usage_percent"
	MetricMemoryUsageBytes  = "memory_usage_bytes"
	MetricGoroutinesTotal   = "goroutines_total"
	MetricGCDurationSeconds = "gc_duration_seconds"
	MetricHeapSizeBytes     = "heap_size_bytes"
)

// Standard buckets for histograms
var (
	// DurationBuckets are suitable for measuring HTTP request, database query,
	// and other operation durations in seconds
	DurationBuckets = []float64{
		.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10,
	}

	// SizeBuckets are suitable for measuring sizes (request, response, payload)
	// in bytes
	SizeBuckets = []float64{
		256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536,
	}
)

// StandardLabels returns a set of standard labels for a service
type StandardLabels struct {
	Service     string
	Instance    string
	Environment string
	Version     string
}

// ToMap converts StandardLabels to a map
func (l StandardLabels) ToMap() map[string]string {
	return map[string]string{
		LabelService:     l.Service,
		LabelInstance:    l.Instance,
		LabelEnvironment: l.Environment,
		LabelVersion:     l.Version,
	}
}

// MetricDefinition represents a standard metric definition
type MetricDefinition struct {
	Name       string
	Help       string
	Type       string // "counter", "gauge", "histogram", "summary"
	Labels     []string
	Buckets    []float64           // for histograms
	Objectives map[float64]float64 // for summaries
	MaxAge     time.Duration       // for summaries
}

// StandardMetrics returns a map of standard metric definitions
func StandardMetrics() map[string]MetricDefinition {
	return map[string]MetricDefinition{
		MetricHTTPRequestsTotal: {
			Name:   MetricHTTPRequestsTotal,
			Help:   "Total number of HTTP requests",
			Type:   "counter",
			Labels: []string{LabelService, LabelEndpoint, LabelMethod, LabelStatusCode},
		},
		MetricHTTPRequestDurationSeconds: {
			Name:    MetricHTTPRequestDurationSeconds,
			Help:    "HTTP request duration in seconds",
			Type:    "histogram",
			Labels:  []string{LabelService, LabelEndpoint, LabelMethod},
			Buckets: DurationBuckets,
		},
		MetricDBQueryDurationSeconds: {
			Name:    MetricDBQueryDurationSeconds,
			Help:    "Database query duration in seconds",
			Type:    "histogram",
			Labels:  []string{LabelService, LabelComponent},
			Buckets: DurationBuckets,
		},
		MetricCacheHitsTotal: {
			Name:   MetricCacheHitsTotal,
			Help:   "Total number of cache hits",
			Type:   "counter",
			Labels: []string{LabelService, LabelComponent},
		},
		MetricServiceHealth: {
			Name:   MetricServiceHealth,
			Help:   "Current health status (0 = unhealthy, 1 = healthy)",
			Type:   "gauge",
			Labels: []string{LabelService, LabelInstance},
		},
	}
}

// NewStandardReporter creates a Reporter with standard metrics pre-registered
func NewStandardReporter(opts Options, labels StandardLabels) *Reporter {
	reporter := New(opts)
	metrics := StandardMetrics()

	// Pre-register standard metrics
	for _, metric := range metrics {
		allLabels := append([]string{}, metric.Labels...)
		switch metric.Type {
		case "counter":
			reporter.Counter(metric.Name, metric.Help, allLabels)
		case "gauge":
			reporter.Gauge(metric.Name, metric.Help, allLabels)
		case "histogram":
			reporter.Histogram(metric.Name, metric.Help, allLabels, metric.Buckets)
		case "summary":
			reporter.Summary(metric.Name, metric.Help, allLabels, metric.Objectives)
		}
	}

	return reporter
}
