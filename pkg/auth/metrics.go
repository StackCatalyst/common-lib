package auth

import (
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsReporter handles authentication metrics
type MetricsReporter struct {
	tokenValidations  *prometheus.CounterVec
	tokenGenerations  *prometheus.CounterVec
	permissionChecks  *prometheus.CounterVec
	validationLatency *prometheus.HistogramVec
	generationLatency *prometheus.HistogramVec
	activeTokens      *prometheus.GaugeVec
}

// NewMetricsReporter creates a new authentication metrics reporter
func NewMetricsReporter(reporter *metrics.Reporter) *MetricsReporter {
	return &MetricsReporter{
		tokenValidations: reporter.Counter(
			"auth_token_validations_total",
			"Total number of token validations",
			[]string{"type", "status"},
		),
		tokenGenerations: reporter.Counter(
			"auth_token_generations_total",
			"Total number of token generations",
			[]string{"type", "status"},
		),
		permissionChecks: reporter.Counter(
			"auth_permission_checks_total",
			"Total number of permission checks",
			[]string{"resource", "action", "status"},
		),
		validationLatency: reporter.Histogram(
			"auth_token_validation_duration_seconds",
			"Token validation duration in seconds",
			[]string{"type"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		),
		generationLatency: reporter.Histogram(
			"auth_token_generation_duration_seconds",
			"Token generation duration in seconds",
			[]string{"type"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		),
		activeTokens: reporter.Gauge(
			"auth_active_tokens",
			"Number of active tokens",
			[]string{"type"},
		),
	}
}

// ObserveTokenValidation records a token validation attempt
func (m *MetricsReporter) ObserveTokenValidation(tokenType TokenType, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "failure"
	}
	m.tokenValidations.WithLabelValues(string(tokenType), status).Inc()
	m.validationLatency.WithLabelValues(string(tokenType)).Observe(duration.Seconds())
}

// ObserveTokenGeneration records a token generation attempt
func (m *MetricsReporter) ObserveTokenGeneration(tokenType TokenType, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "failure"
	}
	m.tokenGenerations.WithLabelValues(string(tokenType), status).Inc()
	m.generationLatency.WithLabelValues(string(tokenType)).Observe(duration.Seconds())
}

// ObservePermissionCheck records a permission check attempt
func (m *MetricsReporter) ObservePermissionCheck(resource Resource, action Action, err error) {
	status := "allowed"
	if err != nil {
		status = "denied"
	}
	m.permissionChecks.WithLabelValues(string(resource), string(action), status).Inc()
}

// SetActiveTokens sets the number of active tokens
func (m *MetricsReporter) SetActiveTokens(tokenType TokenType, count int) {
	m.activeTokens.WithLabelValues(string(tokenType)).Set(float64(count))
}
