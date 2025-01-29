package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Config holds the gRPC client configuration
type Config struct {
	// Target is the server address to connect to
	Target string `json:"target" yaml:"target"`
	// Timeout is the maximum time to wait for a request to complete
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// DialTimeout is the maximum time to wait for connection establishment
	DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	// KeepAlive is the keepalive configuration
	KeepAlive KeepAliveConfig `json:"keep_alive" yaml:"keep_alive"`
	// MaxRetries is the maximum number of retries for a request
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	// RetryWaitMin is the minimum time to wait between retries
	RetryWaitMin time.Duration `json:"retry_wait_min" yaml:"retry_wait_min"`
	// RetryWaitMax is the maximum time to wait between retries
	RetryWaitMax time.Duration `json:"retry_wait_max" yaml:"retry_wait_max"`
}

// KeepAliveConfig holds the keepalive configuration
type KeepAliveConfig struct {
	// Time is the duration after which if there are no client requests,
	// a keepalive ping will be sent
	Time time.Duration `json:"time" yaml:"time"`
	// Timeout is the duration the client waits for a response to a keepalive ping
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// DefaultConfig returns the default gRPC client configuration
func DefaultConfig() Config {
	return Config{
		Timeout:     30 * time.Second,
		DialTimeout: 10 * time.Second,
		KeepAlive: KeepAliveConfig{
			Time:    60 * time.Second,
			Timeout: 20 * time.Second,
		},
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 5 * time.Second,
	}
}

// Client is a gRPC client with connection management and metrics capabilities
type Client struct {
	conn    *grpc.ClientConn
	config  Config
	metrics *MetricsReporter
}

// New creates a new gRPC client
func New(config Config, metricsReporter *metrics.Reporter) (*Client, error) {
	if config.Target == "" {
		return nil, fmt.Errorf("target address must be provided")
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepAlive.Time,
			Timeout:             config.KeepAlive.Timeout,
			PermitWithoutStream: true,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.Target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	return &Client{
		conn:    conn,
		config:  config,
		metrics: NewMetricsReporter(metricsReporter),
	}, nil
}

// Close closes the gRPC client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Connection returns the underlying gRPC client connection
func (c *Client) Connection() *grpc.ClientConn {
	return c.conn
}

// WithTimeout returns a context with the configured timeout
func (c *Client) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.config.Timeout)
}

// WithUnaryInterceptor returns a gRPC dial option that adds the unary interceptor
func (c *Client) WithUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(c.unaryInterceptor())
}

// WithStreamInterceptor returns a gRPC dial option that adds the stream interceptor
func (c *Client) WithStreamInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(c.streamInterceptor())
}

// unaryInterceptor returns a gRPC unary interceptor that adds metrics and error handling
func (c *Client) unaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		c.metrics.ObserveRequest(method, err, time.Since(start))
		return err
	}
}

// streamInterceptor returns a gRPC stream interceptor that adds metrics and error handling
func (c *Client) streamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()
		stream, err := streamer(ctx, desc, cc, method, opts...)
		c.metrics.ObserveRequest(method, err, time.Since(start))
		return stream, err
	}
}
