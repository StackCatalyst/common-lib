package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, "terraorbit", opts.Namespace)
	assert.Empty(t, opts.Subsystem)
	assert.Equal(t, prometheus.DefaultRegisterer, opts.Registry)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{
			name: "default registry",
			opts: DefaultOptions(),
		},
		{
			name: "custom registry",
			opts: Options{
				Namespace: "test",
				Subsystem: "auth",
				Registry:  prometheus.NewRegistry(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := New(tt.opts)
			require.NotNil(t, reporter)
			assert.NotNil(t, reporter.registry)
			assert.NotNil(t, reporter.factory)
		})
	}
}

func TestMetricCreation(t *testing.T) {
	registry := prometheus.NewRegistry()
	reporter := New(Options{
		Namespace: "test",
		Subsystem: "auth",
		Registry:  registry,
	})

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
