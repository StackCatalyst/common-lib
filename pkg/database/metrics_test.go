package database

import (
	"errors"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func newDBTestMetricsReporter() *metrics.Reporter {
	registry := prometheus.NewRegistry()
	return metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "database",
		Registry:  registry,
	})
}

func TestDatabaseMetricsReporter(t *testing.T) {
	reporter := newDBTestMetricsReporter()
	metricsReporter := NewMetricsReporter(reporter)
	require.NotNil(t, metricsReporter)

	t.Run("ObserveQuery", func(t *testing.T) {
		// Test successful query
		metricsReporter.ObserveQuery("select", nil, 100*time.Millisecond)

		// Test failed query
		metricsReporter.ObserveQuery("insert", errors.New("query failed"), 50*time.Millisecond)
	})

	t.Run("ObserveConnectionError", func(t *testing.T) {
		metricsReporter.ObserveConnectionError("connect")
		metricsReporter.ObserveConnectionError("timeout")
	})

	t.Run("SetPoolStats", func(t *testing.T) {
		metricsReporter.SetPoolStats(10, 5, 5)
	})
}
