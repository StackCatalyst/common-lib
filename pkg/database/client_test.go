package database

import (
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetricsReporter() *metrics.Reporter {
	registry := prometheus.NewRegistry()
	return metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "database",
		Registry:  registry,
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "disable", cfg.SSLMode)
	assert.Equal(t, int32(4), cfg.MaxConns)
	assert.Equal(t, int32(0), cfg.MinConns)
	assert.Equal(t, time.Hour, cfg.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, cfg.MaxConnIdleTime)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "test",
				User:     "user",
				Password: "pass",
				MaxConns: 4,
				MinConns: 0,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: Config{
				Port:     5432,
				Database: "test",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Host:     "localhost",
				Port:     0,
				Database: "test",
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "empty database",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "empty user",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "test",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "test",
				User:     "user",
			},
			wantErr: true,
		},
		{
			name: "invalid connection limits",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "test",
				User:     "user",
				Password: "pass",
				MaxConns: 2,
				MinConns: 4,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid_config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Database: "test",
				MaxConns: 4,
			},
			wantErr: false,
		},
		{
			name: "invalid_config",
			config: Config{
				Host: "localhost",
				Port: 5432,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config, newTestMetricsReporter())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if client != nil {
					client.Close()
				}
			}
		})
	}
}

func TestMetricsReporter(t *testing.T) {
	reporter := NewMetricsReporter(newTestMetricsReporter())
	require.NotNil(t, reporter)

	t.Run("ObserveQuery", func(t *testing.T) {
		// Test successful query
		reporter.ObserveQuery("select", nil, 100*time.Millisecond)

		// Test failed query
		reporter.ObserveQuery("insert", assert.AnError, 50*time.Millisecond)
	})

	t.Run("ObserveConnectionError", func(t *testing.T) {
		reporter.ObserveConnectionError("connect")
	})

	t.Run("SetPoolStats", func(t *testing.T) {
		reporter.SetPoolStats(10, 5, 5)
	})
}
