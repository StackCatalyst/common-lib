package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// ServiceHealth represents service health metrics
type ServiceHealth struct {
	reporter    *Reporter
	status      *prometheus.GaugeVec
	uptime      prometheus.Gauge
	lastChecked prometheus.Gauge
}

// ResourceMetrics represents system resource metrics
type ResourceMetrics struct {
	reporter     *Reporter
	cpuUsage     *prometheus.GaugeVec
	memoryUsage  *prometheus.GaugeVec
	goroutines   prometheus.Gauge
	allocatedMem prometheus.Gauge
}

// NewServiceHealth creates a new service health metrics collector
func NewServiceHealth(r *Reporter) *ServiceHealth {
	return &ServiceHealth{
		reporter: r,
		status: r.Gauge(
			"service_health_status",
			"Current health status of the service (0: unhealthy, 1: healthy)",
			[]string{"service", "instance"},
		),
		uptime: r.factory.NewGauge(prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Time since service start in seconds",
		}),
		lastChecked: r.factory.NewGauge(prometheus.GaugeOpts{
			Name: "service_health_last_checked_timestamp",
			Help: "Unix timestamp of the last health check",
		}),
	}
}

// SetHealth updates the service health status
func (h *ServiceHealth) SetHealth(service, instance string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	h.status.WithLabelValues(service, instance).Set(value)
	h.lastChecked.Set(float64(time.Now().Unix()))
}

// UpdateUptime updates the service uptime
func (h *ServiceHealth) UpdateUptime(startTime time.Time) {
	h.uptime.Set(time.Since(startTime).Seconds())
}

// NewResourceMetrics creates a new resource metrics collector
func NewResourceMetrics(r *Reporter) *ResourceMetrics {
	return &ResourceMetrics{
		reporter: r,
		cpuUsage: r.Gauge(
			"system_cpu_usage",
			"CPU usage percentage per core",
			[]string{"core"},
		),
		memoryUsage: r.Gauge(
			"system_memory_usage",
			"Memory usage statistics in bytes",
			[]string{"type"},
		),
		goroutines: r.factory.NewGauge(prometheus.GaugeOpts{
			Name: "go_goroutines_current",
			Help: "Current number of goroutines",
		}),
		allocatedMem: r.factory.NewGauge(prometheus.GaugeOpts{
			Name: "go_memory_allocated_bytes",
			Help: "Currently allocated memory in bytes",
		}),
	}
}

// Collect gathers all resource metrics
func (rm *ResourceMetrics) Collect(ctx context.Context) error {
	// Collect CPU metrics
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, true)
	if err == nil {
		for i, usage := range cpuPercent {
			rm.cpuUsage.WithLabelValues(string(i)).Set(usage)
		}
	}

	// Collect memory metrics
	if vmStat, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		rm.memoryUsage.WithLabelValues("total").Set(float64(vmStat.Total))
		rm.memoryUsage.WithLabelValues("used").Set(float64(vmStat.Used))
		rm.memoryUsage.WithLabelValues("free").Set(float64(vmStat.Free))
		rm.memoryUsage.WithLabelValues("cached").Set(float64(vmStat.Cached))
	}

	// Collect Go runtime metrics
	rm.goroutines.Set(float64(runtime.NumGoroutine()))
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rm.allocatedMem.Set(float64(memStats.Alloc))

	return nil
}

// CustomCollector represents a custom metrics collector
type CustomCollector struct {
	metrics []prometheus.Collector
	collect func(context.Context) error
}

// NewCustomCollector creates a new custom metrics collector
func NewCustomCollector(metrics []prometheus.Collector, collectFunc func(context.Context) error) *CustomCollector {
	return &CustomCollector{
		metrics: metrics,
		collect: collectFunc,
	}
}

// Describe implements prometheus.Collector
func (c *CustomCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		m.Describe(ch)
	}
}

// Collect implements prometheus.Collector
func (c *CustomCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.collect(context.Background()); err != nil {
		// Log error but continue collecting other metrics
		return
	}
	for _, m := range c.metrics {
		m.Collect(ch)
	}
}
