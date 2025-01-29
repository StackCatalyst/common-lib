package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Reporter handles metrics reporting
type Reporter struct {
	registry prometheus.Registerer
	factory  promauto.Factory
}

// Options configures the metrics reporter
type Options struct {
	// Namespace is the metrics namespace (e.g., "terraorbit")
	Namespace string
	// Subsystem is the metrics subsystem (e.g., "auth")
	Subsystem string
	// Registry is an optional custom Prometheus registry
	Registry prometheus.Registerer
}

// DefaultOptions returns the default metrics options
func DefaultOptions() Options {
	return Options{
		Namespace: "terraorbit",
		Subsystem: "",
		Registry:  prometheus.DefaultRegisterer,
	}
}

// New creates a new metrics reporter
func New(opts Options) *Reporter {
	if opts.Registry == nil {
		opts.Registry = prometheus.DefaultRegisterer
	}

	factory := promauto.With(opts.Registry)
	return &Reporter{
		registry: opts.Registry,
		factory:  factory,
	}
}

// Counter creates a new counter metric
func (r *Reporter) Counter(name, help string, labels []string) *prometheus.CounterVec {
	return r.factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

// Gauge creates a new gauge metric
func (r *Reporter) Gauge(name, help string, labels []string) *prometheus.GaugeVec {
	return r.factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

// Histogram creates a new histogram metric
func (r *Reporter) Histogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	return r.factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
}

// Summary creates a new summary metric
func (r *Reporter) Summary(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	return r.factory.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)
}
