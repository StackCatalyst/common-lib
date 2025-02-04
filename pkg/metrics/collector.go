package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// CollectorConfig represents the configuration for the metrics collector
type CollectorConfig struct {
	// ListenAddress is the address where the metrics HTTP server will listen
	ListenAddress string
	// Path is the HTTP path where metrics will be exposed
	Path string
	// CollectionInterval is the interval between metric collections
	CollectionInterval time.Duration
	// Labels are the default labels to be added to all metrics
	Labels map[string]string
}

// DefaultCollectorConfig returns the default collector configuration
func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		ListenAddress:      ":9090",
		Path:               "/metrics",
		CollectionInterval: 15 * time.Second,
		Labels: map[string]string{
			"service": "unknown",
			"env":     "unknown",
		},
	}
}

// MetricsCollector manages metric collection and exposition
type MetricsCollector struct {
	config    CollectorConfig
	registry  *prometheus.Registry
	reporter  *Reporter
	health    *ServiceHealth
	resources *ResourceMetrics
	custom    []*CustomCollector
	server    *http.Server
	mu        sync.RWMutex
}

// NewCollector creates a new metrics collector
func NewCollector(config CollectorConfig) (*MetricsCollector, error) {
	registry := prometheus.NewRegistry()
	reporter := &Reporter{
		registry: registry,
		factory:  promauto.With(registry),
	}

	collector := &MetricsCollector{
		config:    config,
		registry:  registry,
		reporter:  reporter,
		health:    NewServiceHealth(reporter),
		resources: NewResourceMetrics(reporter),
		custom:    make([]*CustomCollector, 0),
	}

	// Register default collectors
	registry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	return collector, nil
}

// RegisterCustomCollector adds a custom collector
func (c *MetricsCollector) RegisterCustomCollector(collector *CustomCollector) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.registry.Register(collector); err != nil {
		return fmt.Errorf("failed to register custom collector: %w", err)
	}
	c.custom = append(c.custom, collector)
	return nil
}

// Start begins collecting and exposing metrics
func (c *MetricsCollector) Start(ctx context.Context) error {
	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle(c.config.Path, promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	c.server = &http.Server{
		Addr:    c.config.ListenAddress,
		Handler: mux,
	}

	// Start collection loop
	go c.collect(ctx)

	// Start HTTP server
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the metrics collector
func (c *MetricsCollector) Stop(ctx context.Context) error {
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

// collect periodically collects metrics
func (c *MetricsCollector) collect(ctx context.Context) {
	ticker := time.NewTicker(c.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.resources.Collect(ctx); err != nil {
				fmt.Printf("Error collecting resource metrics: %v\n", err)
			}

			c.mu.RLock()
			for _, collector := range c.custom {
				if err := collector.collect(ctx); err != nil {
					fmt.Printf("Error collecting custom metrics: %v\n", err)
				}
			}
			c.mu.RUnlock()
		}
	}
}

// GetRegistry returns the Prometheus registry
func (c *MetricsCollector) GetRegistry() *prometheus.Registry {
	return c.registry
}

// GetReporter returns the metrics reporter
func (c *MetricsCollector) GetReporter() *Reporter {
	return c.reporter
}

// GetHealth returns the service health metrics
func (c *MetricsCollector) GetHealth() *ServiceHealth {
	return c.health
}

// GetResources returns the resource metrics
func (c *MetricsCollector) GetResources() *ResourceMetrics {
	return c.resources
}
