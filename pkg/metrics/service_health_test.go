package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceHealthMetrics(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create metrics directly with prometheus to avoid potential registration issues
	status := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "service_health_status",
			Help:      "Current health status of the service (0: unhealthy, 1: healthy)",
		},
		[]string{"service", "instance"},
	)

	uptime := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "service_uptime_seconds",
			Help:      "Time since service start in seconds",
		},
	)

	lastChecked := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "service_health_last_checked_timestamp",
			Help:      "Unix timestamp of the last health check",
		},
	)

	// Create ServiceHealth with the metrics
	sh := &ServiceHealth{
		status:      status,
		uptime:      uptime,
		lastChecked: lastChecked,
	}

	// Register the collector
	err := registry.Register(sh)
	require.NoError(t, err)

	// Test health status
	sh.SetHealth("test-service", "instance-1", true)
	sh.SetHealth("test-service", "instance-2", false)

	// Test uptime
	startTime := time.Now().Add(-time.Hour)
	sh.UpdateUptime(startTime)

	// Gather metrics
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
