package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceHealth(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := New(Options{
		Namespace: "test",
		Registry:  registry,
	})

	health := NewServiceHealth(reporter)
	require.NotNil(t, health)

	// Test health status
	health.SetHealth("test-service", "instance-1", true)
	health.SetHealth("test-service", "instance-2", false)

	// Test uptime
	startTime := time.Now().Add(-time.Hour)
	health.UpdateUptime(startTime)

	// Collect metrics
	metrics, err := registry.Gather()
	require.NoError(t, err)

	// Verify metrics
	found := make(map[string]bool)
	for _, m := range metrics {
		switch m.GetName() {
		case "test_service_health_status":
			found["status"] = true
			assert.Equal(t, 2, len(m.GetMetric()))
			for _, metric := range m.GetMetric() {
				labels := make(map[string]string)
				for _, label := range metric.GetLabel() {
					labels[label.GetName()] = label.GetValue()
				}
				if labels["instance"] == "instance-1" {
					assert.Equal(t, 1.0, metric.GetGauge().GetValue())
				} else {
					assert.Equal(t, 0.0, metric.GetGauge().GetValue())
				}
			}
		case "test_service_uptime_seconds":
			found["uptime"] = true
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Greater(t, m.GetMetric()[0].GetGauge().GetValue(), 3500.0) // ~1 hour
		case "test_service_health_last_checked_timestamp":
			found["last_checked"] = true
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Greater(t, m.GetMetric()[0].GetGauge().GetValue(), float64(time.Now().Add(-time.Minute).Unix()))
		}
	}

	assert.True(t, found["status"], "health status metric not found")
	assert.True(t, found["uptime"], "uptime metric not found")
	assert.True(t, found["last_checked"], "last checked metric not found")
}

func TestResourceMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := New(Options{
		Namespace: "test",
		Registry:  registry,
	})

	resources := NewResourceMetrics(reporter)
	require.NotNil(t, resources)

	// Collect metrics
	err := resources.Collect(context.Background())
	require.NoError(t, err)

	// Gather metrics
	metrics, err := registry.Gather()
	require.NoError(t, err)

	// Verify metrics
	found := make(map[string]bool)
	for _, m := range metrics {
		switch m.GetName() {
		case "test_system_cpu_usage":
			found["cpu"] = true
			assert.NotEmpty(t, m.GetMetric())
		case "test_system_memory_usage":
			found["memory"] = true
			assert.NotEmpty(t, m.GetMetric())
			memTypes := make(map[string]bool)
			for _, metric := range m.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "type" {
						memTypes[label.GetValue()] = true
					}
				}
			}
			assert.True(t, memTypes["total"], "total memory metric not found")
			assert.True(t, memTypes["used"], "used memory metric not found")
			assert.True(t, memTypes["free"], "free memory metric not found")
		case "test_go_goroutines_current":
			found["goroutines"] = true
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Greater(t, m.GetMetric()[0].GetGauge().GetValue(), 1.0)
		case "test_go_memory_allocated_bytes":
			found["allocated_mem"] = true
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Greater(t, m.GetMetric()[0].GetGauge().GetValue(), 0.0)
		}
	}

	assert.True(t, found["cpu"], "CPU metrics not found")
	assert.True(t, found["memory"], "memory metrics not found")
	assert.True(t, found["goroutines"], "goroutine metrics not found")
	assert.True(t, found["allocated_mem"], "allocated memory metrics not found")
}

func TestCustomCollector(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := New(Options{
		Namespace: "test",
		Registry:  registry,
	})

	// Create a test metric
	testMetric := reporter.Gauge(
		"custom_metric",
		"A test custom metric",
		[]string{"label"},
	)

	// Create a custom collector
	collected := false
	collector := NewCustomCollector(
		[]prometheus.Collector{testMetric},
		func(ctx context.Context) error {
			collected = true
			testMetric.WithLabelValues("test").Set(42.0)
			return nil
		},
	)

	// Register the collector
	err := registry.Register(collector)
	require.NoError(t, err)

	// Gather metrics
	metrics, err := registry.Gather()
	require.NoError(t, err)

	// Verify metrics
	assert.True(t, collected, "collect function was not called")
	found := false
	for _, m := range metrics {
		if m.GetName() == "test_custom_metric" {
			found = true
			assert.Equal(t, 1, len(m.GetMetric()))
			metric := m.GetMetric()[0]
			assert.Equal(t, 42.0, metric.GetGauge().GetValue())
			assert.Equal(t, "test", metric.GetLabel()[0].GetValue())
		}
	}
	assert.True(t, found, "custom metric not found")
}
