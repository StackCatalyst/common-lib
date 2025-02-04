package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporterDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, "terraorbit", opts.Namespace)
	assert.Empty(t, opts.Subsystem)
	assert.Equal(t, prometheus.DefaultRegisterer, opts.Registry)
}

func TestReporterMetricCreation(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := New(Options{
		Namespace: "test",
		Subsystem: "auth",
		Registry:  registry,
	})

	require.NotNil(t, reporter)
	require.NotNil(t, reporter.registry)
	require.NotNil(t, reporter.factory)

	t.Run("counter", func(t *testing.T) {
		counter := reporter.Counter(
			"test_counter",
			"Test counter help",
			[]string{"label1", "label2"},
		)
		require.NotNil(t, counter)

		counter.WithLabelValues("value1", "value2").Inc()
		metrics, err := registry.Gather()
		require.NoError(t, err)
		assert.NotEmpty(t, metrics)
	})

	t.Run("gauge", func(t *testing.T) {
		gauge := reporter.Gauge(
			"test_gauge",
			"Test gauge help",
			[]string{"label1"},
		)
		require.NotNil(t, gauge)

		gauge.WithLabelValues("value1").Set(42)
		metrics, err := registry.Gather()
		require.NoError(t, err)
		assert.NotEmpty(t, metrics)
	})

	t.Run("histogram", func(t *testing.T) {
		histogram := reporter.Histogram(
			"test_histogram",
			"Test histogram help",
			[]string{"label1"},
			[]float64{1, 2, 5, 10},
		)
		require.NotNil(t, histogram)

		histogram.WithLabelValues("value1").Observe(3)
		metrics, err := registry.Gather()
		require.NoError(t, err)
		assert.NotEmpty(t, metrics)
	})

	t.Run("summary", func(t *testing.T) {
		summary := reporter.Summary(
			"test_summary",
			"Test summary help",
			[]string{"label1"},
			map[float64]float64{0.5: 0.05, 0.9: 0.01},
		)
		require.NotNil(t, summary)

		summary.WithLabelValues("value1").Observe(3)
		metrics, err := registry.Gather()
		require.NoError(t, err)
		assert.NotEmpty(t, metrics)
	})
}
