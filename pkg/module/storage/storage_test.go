package storage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigJSON(t *testing.T) {
	config := &Config{
		Type: "postgres",
		PostgreSQL: &PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "modules",
			User:     "user",
			Password: "pass",
			SSLMode:  "disable",
		},
		S3: &S3Config{
			Bucket:    "modules",
			Region:    "us-west-2",
			Endpoint:  "http://localhost:9000",
			AccessKey: "access",
			SecretKey: "secret",
		},
		Cache: &CacheConfig{
			Enabled: true,
			TTL:     "1h",
			MaxSize: 1024 * 1024 * 1024,
		},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Test unmarshaling
	var decoded Config
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify PostgreSQL config
	require.NotNil(t, decoded.PostgreSQL)
	assert.Equal(t, config.PostgreSQL.Host, decoded.PostgreSQL.Host)
	assert.Equal(t, config.PostgreSQL.Port, decoded.PostgreSQL.Port)
	assert.Equal(t, config.PostgreSQL.Database, decoded.PostgreSQL.Database)
	assert.Equal(t, config.PostgreSQL.User, decoded.PostgreSQL.User)
	assert.Equal(t, config.PostgreSQL.Password, decoded.PostgreSQL.Password)
	assert.Equal(t, config.PostgreSQL.SSLMode, decoded.PostgreSQL.SSLMode)

	// Verify S3 config
	require.NotNil(t, decoded.S3)
	assert.Equal(t, config.S3.Bucket, decoded.S3.Bucket)
	assert.Equal(t, config.S3.Region, decoded.S3.Region)
	assert.Equal(t, config.S3.Endpoint, decoded.S3.Endpoint)
	assert.Equal(t, config.S3.AccessKey, decoded.S3.AccessKey)
	assert.Equal(t, config.S3.SecretKey, decoded.S3.SecretKey)

	// Verify cache config
	require.NotNil(t, decoded.Cache)
	assert.Equal(t, config.Cache.Enabled, decoded.Cache.Enabled)
	assert.Equal(t, config.Cache.TTL, decoded.Cache.TTL)
	assert.Equal(t, config.Cache.MaxSize, decoded.Cache.MaxSize)
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error with underlying error",
			err: &Error{
				Code:    ErrNotFound,
				Message: "module not found",
				Err:     assert.AnError,
			},
			expected: "module not found: assert.AnError general error for testing",
		},
		{
			name: "error without underlying error",
			err: &Error{
				Code:    ErrInvalidInput,
				Message: "invalid input",
			},
			expected: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
			if tt.err.Err != nil {
				assert.Equal(t, tt.err.Err, tt.err.Unwrap())
			} else {
				assert.Nil(t, tt.err.Unwrap())
			}
		})
	}
}
