package grpc

import (
	"errors"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestGRPCMetricsReporter(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "grpc_client",
		Registry:  registry,
	})

	metricsReporter := NewMetricsReporter(reporter)
	require.NotNil(t, metricsReporter)

	t.Run("ObserveRequest", func(t *testing.T) {
		// Test successful request
		metricsReporter.ObserveRequest("/test.service/Method1", nil, 100*time.Millisecond)

		// Test failed request
		metricsReporter.ObserveRequest("/test.service/Method2", errors.New("request failed"), 50*time.Millisecond)
	})
}
