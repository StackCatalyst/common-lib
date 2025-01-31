package testing

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

func TestContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container tests in short mode")
	}

	if !isDockerAvailable() {
		t.Skip("Docker is not available")
	}

	ctx := context.Background()

	t.Run("PostgreSQL", func(t *testing.T) {
		config := PostgresConfig{
			Database: "test",
			User:     "user",
			Password: "password",
			Version:  "14-alpine",
			Port:     "5432/tcp",
		}
		container, err := PostgresContainer(ctx, config)
		require.NoError(t, err)
		defer container.Stop(ctx)

		host, err := container.GetHost(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, host)

		port, err := container.GetHostPort(ctx, config.Port)
		require.NoError(t, err)
		assert.NotEmpty(t, port)
	})

	t.Run("PostgreSQL Default Config", func(t *testing.T) {
		config := PostgresConfig{
			Database: "test",
			User:     "user",
			Password: "password",
		}
		container, err := PostgresContainer(ctx, config)
		require.NoError(t, err)
		defer container.Stop(ctx)

		port, err := container.GetHostPort(ctx, "5432/tcp")
		require.NoError(t, err)
		assert.NotEmpty(t, port)
	})

	t.Run("Redis", func(t *testing.T) {
		container, err := RedisContainer(ctx)
		require.NoError(t, err)
		defer container.Stop(ctx)

		host, err := container.GetHost(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, host)

		port, err := container.GetHostPort(ctx, "6379/tcp")
		require.NoError(t, err)
		assert.NotEmpty(t, port)
	})

	t.Run("Localstack", func(t *testing.T) {
		container, err := LocalstackContainer(ctx, []string{"s3", "dynamodb"})
		require.NoError(t, err)
		defer container.Stop(ctx)

		host, err := container.GetHost(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, host)

		port, err := container.GetHostPort(ctx, "4566/tcp")
		require.NoError(t, err)
		assert.NotEmpty(t, port)
	})

	t.Run("Kafka", func(t *testing.T) {
		config := KafkaConfig{
			Version:    "3.5",
			BrokerPort: "9092/tcp",
			ZookerPort: "2181/tcp",
			Topics:     []string{"test-topic"},
			Partitions: 3,
			Replicas:   1,
			ExternalIP: "localhost",
		}
		container, err := KafkaContainer(ctx, config)
		require.NoError(t, err)
		defer container.Stop(ctx)

		host, err := container.GetHost(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, host)

		brokerPort, err := container.GetHostPort(ctx, config.BrokerPort)
		require.NoError(t, err)
		assert.NotEmpty(t, brokerPort)

		zookeeperPort, err := container.GetHostPort(ctx, config.ZookerPort)
		require.NoError(t, err)
		assert.NotEmpty(t, zookeeperPort)
	})

	t.Run("Kafka Default Config", func(t *testing.T) {
		config := KafkaConfig{
			Topics: []string{"test-topic"},
		}
		container, err := KafkaContainer(ctx, config)
		require.NoError(t, err)
		defer container.Stop(ctx)

		brokerPort, err := container.GetHostPort(ctx, "9092/tcp")
		require.NoError(t, err)
		assert.NotEmpty(t, brokerPort)
	})
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
