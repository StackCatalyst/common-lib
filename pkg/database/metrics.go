package database

import (
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsReporter handles database metrics reporting
type MetricsReporter struct {
	queryExecutions  *prometheus.CounterVec
	queryLatency     *prometheus.HistogramVec
	connectionErrors *prometheus.CounterVec
	poolStats        *prometheus.GaugeVec
}

// NewMetricsReporter creates a new database metrics reporter
func NewMetricsReporter(reporter *metrics.Reporter) *MetricsReporter {
	return &MetricsReporter{
		queryExecutions: reporter.Counter(
			"database_query_executions_total",
			"Total number of database query executions",
			[]string{"type", "status"},
		),
		queryLatency: reporter.Histogram(
			"database_query_duration_seconds",
			"Database query duration in seconds",
			[]string{"type"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		),
		connectionErrors: reporter.Counter(
			"database_connection_errors_total",
			"Total number of database connection errors",
			[]string{"type"},
		),
		poolStats: reporter.Gauge(
			"database_pool_stats",
			"Database connection pool statistics",
			[]string{"type"},
		),
	}
}

// ObserveQuery records a query execution
func (m *MetricsReporter) ObserveQuery(queryType string, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "failure"
	}
	m.queryExecutions.WithLabelValues(queryType, status).Inc()
	m.queryLatency.WithLabelValues(queryType).Observe(duration.Seconds())
}

// ObserveConnectionError records a connection error
func (m *MetricsReporter) ObserveConnectionError(errorType string) {
	m.connectionErrors.WithLabelValues(errorType).Inc()
}

// SetPoolStats sets the current pool statistics
func (m *MetricsReporter) SetPoolStats(totalConns, idleConns, inUseConns int64) {
	m.poolStats.WithLabelValues("total").Set(float64(totalConns))
	m.poolStats.WithLabelValues("idle").Set(float64(idleConns))
	m.poolStats.WithLabelValues("in_use").Set(float64(inUseConns))
}
