package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	content := []byte(`
database:
  host: localhost
  port: 5432
  user: admin
server:
  port: 8080
  timeout: 30s
`)

	err := os.WriteFile(configFile, content, 0644)
	require.NoError(t, err)

	// Create config manager
	opts := Options{
		ConfigName:   "config",
		ConfigType:   "yaml",
		ConfigPaths:  []string{dir},
		EnvPrefix:    "TEST",
		AutomaticEnv: true,
	}

	cfg, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test reading values
	assert.Equal(t, "localhost", cfg.GetString("database.host"))
	assert.Equal(t, 5432, cfg.GetInt("database.port"))
	assert.Equal(t, "admin", cfg.GetString("database.user"))
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
	assert.Equal(t, 30*time.Second, cfg.GetDuration("server.timeout"))
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_DATABASE_HOST", "testhost")
	os.Setenv("TEST_SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("TEST_DATABASE_HOST")
		os.Unsetenv("TEST_SERVER_PORT")
	}()

	opts := DefaultOptions()
	opts.EnvPrefix = "TEST"

	cfg, err := New(opts)
	require.NoError(t, err)

	// Test environment variable override
	assert.Equal(t, "testhost", cfg.GetString("database.host"))
	assert.Equal(t, 9090, cfg.GetInt("server.port"))
}

func TestSetAndGet(t *testing.T) {
	cfg, err := New(DefaultOptions())
	require.NoError(t, err)

	// Test setting and getting values
	cfg.Set("test.string", "value")
	cfg.Set("test.int", 42)
	cfg.Set("test.bool", true)
	cfg.Set("test.slice", []string{"a", "b", "c"})
	cfg.Set("test.map", map[string]interface{}{
		"key": "value",
	})

	assert.Equal(t, "value", cfg.GetString("test.string"))
	assert.Equal(t, 42, cfg.GetInt("test.int"))
	assert.Equal(t, true, cfg.GetBool("test.bool"))
	assert.Equal(t, []string{"a", "b", "c"}, cfg.GetStringSlice("test.slice"))
	assert.Equal(t, "value", cfg.GetStringMap("test.map")["key"])
}

type TestConfig struct {
	Database struct {
		Host string
		Port int
	}
	Server struct {
		Port    int
		Timeout time.Duration
	}
}

func TestUnmarshal(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	content := []byte(`
database:
  host: localhost
  port: 5432
server:
  port: 8080
  timeout: 30s
`)

	err := os.WriteFile(configFile, content, 0644)
	require.NoError(t, err)

	opts := Options{
		ConfigName:   "config",
		ConfigType:   "yaml",
		ConfigPaths:  []string{dir},
		EnvPrefix:    "TEST",
		AutomaticEnv: true,
	}

	cfg, err := New(opts)
	require.NoError(t, err)

	var config TestConfig
	err = cfg.Unmarshal(&config)
	require.NoError(t, err)

	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, 30*time.Second, config.Server.Timeout)
}
