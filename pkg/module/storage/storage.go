package storage

import (
	"context"
	"io"

	"github.com/StackCatalyst/common-lib/pkg/module"
)

// Backend represents a storage backend for modules
type Backend interface {
	// Store stores a module in the backend
	Store(ctx context.Context, mod *module.Module, content io.Reader) error

	// Get retrieves a module from the backend
	Get(ctx context.Context, id string) (*module.Module, io.ReadCloser, error)

	// List lists modules matching the filter
	List(ctx context.Context, filter module.Filter) ([]*module.Module, error)

	// Delete deletes a module from the backend
	Delete(ctx context.Context, id string) error

	// GetContent retrieves the module content
	GetContent(ctx context.Context, id string) (io.ReadCloser, error)

	// StoreContent stores the module content
	StoreContent(ctx context.Context, id string, content io.Reader) error

	// DeleteContent deletes the module content
	DeleteContent(ctx context.Context, id string) error

	// Close closes the storage backend
	Close() error
}

// Config holds the storage backend configuration
type Config struct {
	// Type is the storage backend type (e.g., postgres, s3)
	Type string `json:"type" yaml:"type"`

	// PostgreSQL configuration
	PostgreSQL *PostgreSQLConfig `json:"postgresql,omitempty" yaml:"postgresql,omitempty"`

	// S3 configuration
	S3 *S3Config `json:"s3,omitempty" yaml:"s3,omitempty"`

	// Cache configuration
	Cache *CacheConfig `json:"cache,omitempty" yaml:"cache,omitempty"`
}

// PostgreSQLConfig holds PostgreSQL-specific configuration
type PostgreSQLConfig struct {
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
}

// S3Config holds S3-specific configuration
type S3Config struct {
	// Bucket is the S3 bucket name
	Bucket string `json:"bucket" yaml:"bucket"`
	// Region is the AWS region
	Region string `json:"region" yaml:"region"`
	// Endpoint is the S3-compatible endpoint (optional)
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	// AccessKey is the AWS access key
	AccessKey string `json:"access_key" yaml:"access_key"`
	// SecretKey is the AWS secret key
	SecretKey string `json:"secret_key" yaml:"secret_key"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	// Enabled indicates if caching is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`
	// TTL is the cache time-to-live
	TTL string `json:"ttl" yaml:"ttl"`
	// MaxSize is the maximum cache size in bytes
	MaxSize int64 `json:"max_size" yaml:"max_size"`
}

// Error represents a storage error
type Error struct {
	// Code is the error code
	Code string
	// Message is the error message
	Message string
	// Err is the underlying error
	Err error
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// Error codes
const (
	ErrNotFound      = "NOT_FOUND"
	ErrAlreadyExists = "ALREADY_EXISTS"
	ErrInvalidInput  = "INVALID_INPUT"
	ErrInternal      = "INTERNAL"
)
