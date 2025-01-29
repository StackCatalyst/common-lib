package http

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestHTTPMetricsReporter(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "http_client",
		Registry:  registry,
	})

	metricsReporter := NewMetricsReporter(reporter)
	require.NotNil(t, metricsReporter)

	t.Run("ObserveRequest", func(t *testing.T) {
		// Test successful request
		resp := &http.Response{
			StatusCode: http.StatusOK,
		}
		metricsReporter.ObserveRequest("GET", resp, nil, 100*time.Millisecond)

		// Test failed request
		metricsReporter.ObserveRequest("POST", nil, errors.New("request failed"), 50*time.Millisecond)
	})

	t.Run("ObserveError", func(t *testing.T) {
		metricsReporter.ObserveError("connection_failed")
		metricsReporter.ObserveError("timeout")
	})
}
