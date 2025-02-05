package testing

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func isDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Docker not available: %v\n", err)
		return false
	}
	return true
}

func TestContainer(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker is not available")
	}

	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(context.Context) (*Container, error)
	}{
		{
			name: "PostgreSQL",
			fn: func(ctx context.Context) (*Container, error) {
				return PostgresContainer(ctx, PostgresConfig{
					Database: "test",
					User:     "test",
					Password: "test",
					Version:  "15-alpine",
					Port:     "5432/tcp",
				})
			},
		},
		{
			name: "PostgreSQL_Default_Config",
			fn: func(ctx context.Context) (*Container, error) {
				return PostgresContainer(ctx, PostgresConfig{})
			},
		},
		{
			name: "Redis",
			fn: func(ctx context.Context) (*Container, error) {
				return RedisContainer(ctx)
			},
		},
		{
			name: "Localstack",
			fn: func(ctx context.Context) (*Container, error) {
				return LocalstackContainer(ctx, []string{"s3"})
			},
		},
		{
			name: "Kafka",
			fn: func(ctx context.Context) (*Container, error) {
				return KafkaContainer(ctx, KafkaConfig{
					Version:    "7.5.1",
					BrokerPort: "9092/tcp",
					Topics:     []string{"test"},
					Partitions: 1,
					Replicas:   1,
					ExternalIP: "localhost",
				})
			},
		},
		{
			name: "Kafka_Default_Config",
			fn: func(ctx context.Context) (*Container, error) {
				return KafkaContainer(ctx, KafkaConfig{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := tt.fn(ctx)
			require.NoError(t, err)
			require.NotNil(t, container)

			// Test port mapping for specific containers
			switch tt.name {
			case "PostgreSQL", "PostgreSQL_Default_Config":
				port, err := container.GetHostPort(ctx, "5432/tcp")
				assert.NotEmpty(t, port)
				assert.NoError(t, err)
			case "Redis":
				port, err := container.GetHostPort(ctx, "6379/tcp")
				assert.NotEmpty(t, port)
				assert.NoError(t, err)
			case "Localstack":
				port, err := container.GetHostPort(ctx, "4566/tcp")
				assert.NotEmpty(t, port)
				assert.NoError(t, err)
			case "Kafka", "Kafka_Default_Config":
				brokerPort, err := container.GetHostPort(ctx, "9092/tcp")
				assert.NotEmpty(t, brokerPort)
				assert.NoError(t, err)
			}

			// Test host retrieval
			host, err := container.GetHost(ctx)
			assert.NotEmpty(t, host)
			assert.NoError(t, err)

			// Test container stop
			err = container.Stop(ctx)
			assert.NoError(t, err)
		})
	}
}

func TestContainerConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container tests in short mode")
	}

	if !isDockerAvailable() {
		t.Skip("Docker is not available")
	}

	ctx := context.Background()

	t.Run("CustomConfig", func(t *testing.T) {
		config := ContainerConfig{
			Image: "nginx",
			Tag:   "alpine",
			Env: map[string]string{
				"NGINX_PORT": "8080",
			},
			Ports: map[string]string{
				"80/tcp": "8080",
			},
			Command:        []string{"nginx", "-g", "daemon off;"},
			StartupTimeout: 5 * time.Second,
			WaitStrategy:   wait.ForLog("Configuration complete; ready for start up"),
		}

		container, err := NewContainer(ctx, config)
		require.NoError(t, err)
		defer container.Stop(ctx)

		host, err := container.GetHost(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, host)

		port, err := container.GetHostPort(ctx, "80/tcp")
		require.NoError(t, err)
		assert.Equal(t, "8080", port)
	})

	t.Run("DefaultTimeout", func(t *testing.T) {
		config := ContainerConfig{
			Image: "nginx",
			Tag:   "alpine",
		}

		container, err := NewContainer(ctx, config)
		require.NoError(t, err)
		defer container.Stop(ctx)

		assert.Equal(t, 60*time.Second, container.config.StartupTimeout)
	})
}
