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

// PostgresConfig holds PostgreSQL specific configuration
type PostgresConfig struct {
	Database string
	User     string
	Password string
	Version  string // e.g., "14-alpine", "15-alpine"
	Port     string // defaults to 5432/tcp
}

// PostgresContainer creates a PostgreSQL test container with advanced configuration
func PostgresContainer(ctx context.Context, config PostgresConfig) (*Container, error) {
	if config.Version == "" {
		config.Version = "14-alpine"
	}
	if config.Port == "" {
		config.Port = "5432/tcp"
	}

	containerConfig := ContainerConfig{
		Image: "postgres",
		Tag:   config.Version,
		Env: map[string]string{
			"POSTGRES_DB":       config.Database,
			"POSTGRES_USER":     config.User,
			"POSTGRES_PASSWORD": config.Password,
		},
		Ports: map[string]string{
			config.Port: "",
		},
		WaitStrategy: wait.ForLog("database system is ready to accept connections"),
	}
	return NewContainer(ctx, containerConfig)
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

// KafkaConfig holds Kafka specific configuration
type KafkaConfig struct {
	Version    string // e.g., "3.5", "3.6"
	BrokerPort string // defaults to 9092/tcp
	ZookerPort string // defaults to 2181/tcp
	Topics     []string
	Partitions int
	Replicas   int
	ExternalIP string // for advertised listeners
}

// KafkaContainer creates a Kafka test container with Zookeeper
func KafkaContainer(ctx context.Context, config KafkaConfig) (*Container, error) {
	if config.Version == "" {
		config.Version = "7.5.1"
	}
	if config.BrokerPort == "" {
		config.BrokerPort = "9092/tcp"
	}
	if config.ZookerPort == "" {
		config.ZookerPort = "2181/tcp"
	}
	if config.Partitions == 0 {
		config.Partitions = 1
	}
	if config.Replicas == 0 {
		config.Replicas = 1
	}
	if config.ExternalIP == "" {
		config.ExternalIP = "localhost"
	}

	containerConfig := ContainerConfig{
		Image: "confluentinc/cp-kafka",
		Tag:   config.Version,
		Env: map[string]string{
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_ZOOKEEPER_CONNECT":                        "localhost:2181",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS":                     fmt.Sprintf("PLAINTEXT://%s:%s,PLAINTEXT_HOST://localhost:%s", config.ExternalIP, config.BrokerPort, config.BrokerPort),
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                "true",
			"KAFKA_NUM_PARTITIONS":                           fmt.Sprintf("%d", config.Partitions),
			"KAFKA_DEFAULT_REPLICATION_FACTOR":               fmt.Sprintf("%d", config.Replicas),
		},
		Ports: map[string]string{
			config.BrokerPort: "",
			config.ZookerPort: "",
		},
		WaitStrategy: wait.ForLog("[KafkaServer id=1] started"),
	}
	return NewContainer(ctx, containerConfig)
}
