package database

import (
	"context"
	"fmt"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the database configuration
type Config struct {
	// Host is the database server hostname
	Host string `json:"host" yaml:"host"`
	// Port is the database server port
	Port int `json:"port" yaml:"port"`
	// Database is the name of the database to connect to
	Database string `json:"database" yaml:"database"`
	// User is the database user
	User string `json:"user" yaml:"user"`
	// Password is the database password
	Password string `json:"password" yaml:"password"`
	// SSLMode is the SSL mode to use for the connection
	SSLMode string `json:"ssl_mode" yaml:"ssl_mode"`
	// MaxConns is the maximum number of connections in the pool
	MaxConns int32 `json:"max_conns" yaml:"max_conns"`
	// MinConns is the minimum number of connections in the pool
	MinConns int32 `json:"min_conns" yaml:"min_conns"`
	// MaxConnLifetime is the maximum lifetime of a connection
	MaxConnLifetime time.Duration `json:"max_conn_lifetime" yaml:"max_conn_lifetime"`
	// MaxConnIdleTime is the maximum idle time of a connection
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time" yaml:"max_conn_idle_time"`
}

// DefaultConfig returns the default database configuration
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            5432,
		SSLMode:         "disable",
		MaxConns:        4,
		MinConns:        0,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}

// Validate validates the database configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host must be provided")
	}
	if c.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	if c.Database == "" {
		return fmt.Errorf("database must be provided")
	}
	if c.User == "" {
		return fmt.Errorf("user must be provided")
	}
	if c.Password == "" {
		return fmt.Errorf("password must be provided")
	}
	if c.MaxConns < c.MinConns {
		return fmt.Errorf("max_conns must be greater than or equal to min_conns")
	}
	return nil
}

// Client is a database client that provides connection management and metrics
type Client struct {
	pool    *pgxpool.Pool
	metrics *MetricsReporter
}

// New creates a new database client
func New(config Config, metricsReporter *metrics.Reporter) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	connString := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Database,
		config.User,
		config.Password,
		config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	return &Client{
		pool:    pool,
		metrics: NewMetricsReporter(metricsReporter),
	}, nil
}

// Close closes the database client and its connection pool
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Ping verifies a connection to the database is still alive
func (c *Client) Ping(ctx context.Context) error {
	start := time.Now()
	err := c.pool.Ping(ctx)
	if err != nil {
		c.metrics.ObserveConnectionError("ping")
	}
	c.metrics.ObserveQuery("ping", err, time.Since(start))
	return err
}

// Begin starts a new transaction
func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	start := time.Now()
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		c.metrics.ObserveConnectionError("begin_transaction")
	}
	c.metrics.ObserveQuery("begin_transaction", err, time.Since(start))
	return tx, err
}

// BeginTx starts a new transaction with the specified options
func (c *Client) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	start := time.Now()
	tx, err := c.pool.BeginTx(ctx, txOptions)
	if err != nil {
		c.metrics.ObserveConnectionError("begin_transaction")
	}
	c.metrics.ObserveQuery("begin_transaction", err, time.Since(start))
	return tx, err
}

// Query executes a query that returns rows
func (c *Client) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	rows, err := c.pool.Query(ctx, sql, args...)
	c.metrics.ObserveQuery("query", err, time.Since(start))
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row
func (c *Client) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()
	row := c.pool.QueryRow(ctx, sql, args...)
	c.metrics.ObserveQuery("query_row", nil, time.Since(start))
	return row
}

// Exec executes a query that doesn't return rows
func (c *Client) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	tag, err := c.pool.Exec(ctx, sql, args...)
	c.metrics.ObserveQuery("exec", err, time.Since(start))
	return tag, err
}

// UpdatePoolStats updates the pool statistics metrics
func (c *Client) UpdatePoolStats() {
	stats := c.pool.Stat()
	c.metrics.SetPoolStats(
		int64(stats.TotalConns()),
		int64(stats.IdleConns()),
		int64(stats.AcquiredConns()),
	)
}
