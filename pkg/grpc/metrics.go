package grpc

import (
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsReporter handles gRPC client metrics reporting
type MetricsReporter struct {
	requestDuration *prometheus.HistogramVec
	requestErrors   *prometheus.CounterVec
}

// NewMetricsReporter creates a new gRPC client metrics reporter
func NewMetricsReporter(reporter *metrics.Reporter) *MetricsReporter {
	return &MetricsReporter{
		requestDuration: reporter.Histogram(
			"grpc_client_request_duration_seconds",
			"gRPC client request duration in seconds",
			[]string{"method", "status"},
			[]float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		),
		requestErrors: reporter.Counter(
			"grpc_client_errors_total",
			"Total number of gRPC client errors",
			[]string{"method"},
		),
	}
}

// ObserveRequest records a request execution
func (m *MetricsReporter) ObserveRequest(method string, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "error"
		m.requestErrors.WithLabelValues(method).Inc()
	}
	m.requestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}
