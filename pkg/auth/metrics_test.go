package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func newAuthTestMetricsReporter() *metrics.Reporter {
	registry := prometheus.NewRegistry()
	return metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "auth",
		Registry:  registry,
	})
}

func TestAuthMetricsReporter(t *testing.T) {
	reporter := newAuthTestMetricsReporter()
	metricsReporter := NewMetricsReporter(reporter)
	require.NotNil(t, metricsReporter)

	t.Run("ObserveTokenValidation", func(t *testing.T) {
		// Test successful validation
		metricsReporter.ObserveTokenValidation(AccessToken, nil, 100*time.Millisecond)

		// Test failed validation
		metricsReporter.ObserveTokenValidation(AccessToken, errors.New("invalid token"), 50*time.Millisecond)
	})

	t.Run("ObserveTokenGeneration", func(t *testing.T) {
		// Test successful generation
		metricsReporter.ObserveTokenGeneration(AccessToken, nil, 75*time.Millisecond)

		// Test failed generation
		metricsReporter.ObserveTokenGeneration(AccessToken, errors.New("generation failed"), 25*time.Millisecond)
	})

	t.Run("ObservePermissionCheck", func(t *testing.T) {
		// Test allowed permission
		metricsReporter.ObservePermissionCheck(Resource("users"), Action("read"), nil)

		// Test denied permission
		metricsReporter.ObservePermissionCheck(Resource("users"), Action("write"), errors.New("permission denied"))
	})

	t.Run("SetActiveTokens", func(t *testing.T) {
		metricsReporter.SetActiveTokens(AccessToken, 100)
		metricsReporter.SetActiveTokens(RefreshToken, 50)
	})
}
