package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func newTestMetricsReporter() *metrics.Reporter {
	registry := prometheus.NewRegistry()
	return metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "grpc_client",
		Registry:  registry,
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 10*time.Second, cfg.DialTimeout)
	assert.Equal(t, 60*time.Second, cfg.KeepAlive.Time)
	assert.Equal(t, 20*time.Second, cfg.KeepAlive.Timeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, time.Second, cfg.RetryWaitMin)
	assert.Equal(t, 5*time.Second, cfg.RetryWaitMax)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Target:      "localhost:50051",
				Timeout:     30 * time.Second,
				DialTimeout: 10 * time.Second,
				KeepAlive: KeepAliveConfig{
					Time:    60 * time.Second,
					Timeout: 20 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "empty target",
			config: Config{
				Timeout:     30 * time.Second,
				DialTimeout: 10 * time.Second,
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
				client.Close()
			}
		})
	}
}

func TestClientConnection(t *testing.T) {
	// Create a buffer connection for testing
	lis := bufconn.Listen(bufSize)
	defer lis.Close()

	// Create a test server
	server := grpc.NewServer()
	defer server.Stop()

	// Start server
	go func() {
		if err := server.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	// Create a dialer for the buffer connection
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// Create client
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	// Test connection state
	state := conn.GetState()
	assert.NotEqual(t, connectivity.Shutdown, state)
}

func TestClientTimeout(t *testing.T) {
	client := &Client{
		config: Config{
			Timeout: 100 * time.Millisecond,
		},
	}

	ctx, cancel := client.WithTimeout(context.Background())
	defer cancel()

	select {
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Error("Context should have timed out")
	}
}

func TestClientInterceptors(t *testing.T) {
	client := &Client{
		metrics: NewMetricsReporter(newTestMetricsReporter()),
	}

	// Test unary interceptor
	unaryOpt := client.WithUnaryInterceptor()
	assert.NotNil(t, unaryOpt)

	// Test stream interceptor
	streamOpt := client.WithStreamInterceptor()
	assert.NotNil(t, streamOpt)
}
