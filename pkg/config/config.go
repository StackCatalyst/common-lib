package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Manager handles configuration management
type Manager struct {
	v *viper.Viper
}

// Options configures the behavior of the config manager
type Options struct {
	// ConfigName is the name of the config file without extension
	ConfigName string
	// ConfigType is the type of the config file (yaml, json, etc.)
	ConfigType string
	// ConfigPaths are the paths to search for config files
	ConfigPaths []string
	// EnvPrefix is the prefix for environment variables
	EnvPrefix string
	// AutomaticEnv enables automatic environment variable binding
	AutomaticEnv bool
}

// DefaultOptions returns the default configuration options
func DefaultOptions() Options {
	return Options{
		ConfigName:   "config",
		ConfigType:   "yaml",
		ConfigPaths:  []string{".", "./config", "/etc/terraorbit"},
		EnvPrefix:    "TERRAORBIT",
		AutomaticEnv: true,
	}
}

// New creates a new configuration manager
func New(opts Options) (*Manager, error) {
	v := viper.New()

	// Set config name and type
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)

	// Add config paths
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Configure environment variables
	if opts.AutomaticEnv {
		v.AutomaticEnv()
		v.SetEnvPrefix(opts.EnvPrefix)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if we can't find a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	return &Manager{v: v}, nil
}

// Get retrieves a value from the configuration
func (m *Manager) Get(key string) interface{} {
	return m.v.Get(key)
}

// GetString retrieves a string value from the configuration
func (m *Manager) GetString(key string) string {
	return m.v.GetString(key)
}

// GetInt retrieves an integer value from the configuration
func (m *Manager) GetInt(key string) int {
	return m.v.GetInt(key)
}

// GetBool retrieves a boolean value from the configuration
func (m *Manager) GetBool(key string) bool {
	return m.v.GetBool(key)
}

// GetDuration retrieves a duration value from the configuration
func (m *Manager) GetDuration(key string) time.Duration {
	return m.v.GetDuration(key)
}

// GetStringSlice retrieves a string slice from the configuration
func (m *Manager) GetStringSlice(key string) []string {
	return m.v.GetStringSlice(key)
}

// GetStringMap retrieves a string map from the configuration
func (m *Manager) GetStringMap(key string) map[string]interface{} {
	return m.v.GetStringMap(key)
}

// Set sets a value in the configuration
func (m *Manager) Set(key string, value interface{}) {
	m.v.Set(key, value)
}

// UnmarshalKey takes a key and unmarshals it into a struct
func (m *Manager) UnmarshalKey(key string, rawVal interface{}) error {
	return m.v.UnmarshalKey(key, rawVal)
}

// Unmarshal unmarshals the config into a struct
func (m *Manager) Unmarshal(rawVal interface{}) error {
	return m.v.Unmarshal(rawVal)
}

// WatchConfig watches for configuration changes
func (m *Manager) WatchConfig(onChange func()) {
	m.v.OnConfigChange(func(_ fsnotify.Event) {
		if onChange != nil {
			onChange()
		}
	})
	m.v.WatchConfig()
}

// WriteConfig writes the current configuration to file
func (m *Manager) WriteConfig() error {
	return m.v.WriteConfig()
}

// WriteConfigAs writes the current configuration to a specific file
func (m *Manager) WriteConfigAs(filename string) error {
	return m.v.WriteConfigAs(filename)
}
