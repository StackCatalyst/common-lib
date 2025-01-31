package testing

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerConfig holds configuration for a test container
type ContainerConfig struct {
	// Image is the Docker image to use
	Image string
	// Tag is the image tag to use
	Tag string
	// Env is a map of environment variables
	Env map[string]string
	// Ports is a map of container ports to expose
	Ports map[string]string
	// Command is the command to run in the container
	Command []string
	// Entrypoint is the entrypoint to use
	Entrypoint []string
	// WaitStrategy is the strategy to wait for container readiness
	WaitStrategy wait.Strategy
	// StartupTimeout is the maximum time to wait for container startup
	StartupTimeout time.Duration
}

// Container represents a test container
type Container struct {
	container testcontainers.Container
	config    ContainerConfig
}

// NewContainer creates a new test container
func NewContainer(ctx context.Context, config ContainerConfig) (*Container, error) {
	if config.StartupTimeout == 0 {
		config.StartupTimeout = 60 * time.Second
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      fmt.Sprintf("%s:%s", config.Image, config.Tag),
			Env:        config.Env,
			Cmd:        config.Command,
			Entrypoint: config.Entrypoint,
			WaitingFor: config.WaitStrategy,
		},
		Started: true,
	}

	// Convert ports map to exposed ports
	exposedPorts := make([]string, 0)
	for containerPort, hostPort := range config.Ports {
		if hostPort == "" {
			exposedPorts = append(exposedPorts, containerPort)
		} else {
			exposedPorts = append(exposedPorts, fmt.Sprintf("%s:%s", hostPort, containerPort))
		}
	}
	req.ContainerRequest.ExposedPorts = exposedPorts

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return &Container{
		container: container,
		config:    config,
	}, nil
}

// GetHostPort returns the host port for a given container port
func (c *Container) GetHostPort(ctx context.Context, containerPort string) (string, error) {
	mappedPort, err := c.container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}
	return mappedPort.Port(), nil
}

// GetHost returns the host where the container is running
func (c *Container) GetHost(ctx context.Context) (string, error) {
	return c.container.Host(ctx)
}

// Stop stops the container
func (c *Container) Stop(ctx context.Context) error {
	return c.container.Terminate(ctx)
}

// PostgresContainer creates a PostgreSQL test container
func PostgresContainer(ctx context.Context, database, user, password string) (*Container, error) {
	config := ContainerConfig{
		Image: "postgres",
		Tag:   "14-alpine",
		Env: map[string]string{
			"POSTGRES_DB":       database,
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
		},
		Ports: map[string]string{
			"5432/tcp": "",
		},
		WaitStrategy: wait.ForLog("database system is ready to accept connections"),
	}
	return NewContainer(ctx, config)
}

// RedisContainer creates a Redis test container
func RedisContainer(ctx context.Context) (*Container, error) {
	config := ContainerConfig{
		Image: "redis",
		Tag:   "6-alpine",
		Ports: map[string]string{
			"6379/tcp": "",
		},
		WaitStrategy: wait.ForLog("Ready to accept connections"),
	}
	return NewContainer(ctx, config)
}

// LocalstackContainer creates a Localstack test container
func LocalstackContainer(ctx context.Context, services []string) (*Container, error) {
	config := ContainerConfig{
		Image: "localstack/localstack",
		Tag:   "latest",
		Env: map[string]string{
			"SERVICES":       "s3,dynamodb",
			"DEFAULT_REGION": "us-east-1",
		},
		Ports: map[string]string{
			"4566/tcp": "",
		},
		WaitStrategy: wait.ForLog("Ready."),
	}
	return NewContainer(ctx, config)
}
