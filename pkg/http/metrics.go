package http

import (
	"net/http"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsReporter handles HTTP client metrics reporting
type MetricsReporter struct {
	requestDuration *prometheus.HistogramVec
	requestErrors   *prometheus.CounterVec
}

// NewMetricsReporter creates a new HTTP client metrics reporter
func NewMetricsReporter(reporter *metrics.Reporter) *MetricsReporter {
	return &MetricsReporter{
		requestDuration: reporter.Histogram(
			"http_client_request_duration_seconds",
			"HTTP client request duration in seconds",
			[]string{"method", "status"},
			[]float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		),
		requestErrors: reporter.Counter(
			"http_client_errors_total",
			"Total number of HTTP client errors",
			[]string{"type"},
		),
	}
}

// ObserveRequest records a request execution
func (m *MetricsReporter) ObserveRequest(method string, resp *http.Response, err error, duration time.Duration) {
	status := "error"
	if err == nil && resp != nil {
		status = "success"
	}
	m.requestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

// ObserveError records a client error
func (m *MetricsReporter) ObserveError(errorType string) {
	m.requestErrors.WithLabelValues(errorType).Inc()
}
