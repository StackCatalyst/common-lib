package main

import (
	"fmt"
	"log"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/config"
)

// AppConfig represents application configuration
type AppConfig struct {
	Server struct {
		Host    string
		Port    int
		Timeout time.Duration
	}
	Database struct {
		Host     string
		Port     int
		Name     string
		User     string
		Password string
	}
	Features struct {
		EnableCache  bool
		MaxRequests  int
		AllowedUsers []string
	}
}

func main() {
	// Example 1: Basic configuration setup
	opts := config.DefaultOptions()
	opts.ConfigName = "config"
	opts.ConfigType = "yaml"
	opts.ConfigPaths = []string{".", "./config"}
	opts.EnvPrefix = "APP"
	opts.AutomaticEnv = true

	// Create config manager
	manager, err := config.New(opts)
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}

	// Example 2: Reading individual values
	host := manager.GetString("server.host")
	port := manager.GetInt("server.port")
	timeout := manager.GetDuration("server.timeout")
	fmt.Printf("Server config: %s:%d (timeout: %v)\n", host, port, timeout)

	// Example 3: Using environment variables
	// Export APP_DATABASE_PASSWORD=secret
	dbPassword := manager.GetString("database.password")
	fmt.Printf("Database password from env: %s\n", dbPassword)

	// Example 4: Unmarshaling into struct
	var appConfig AppConfig
	if err := manager.Unmarshal(&appConfig); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Example 5: Watching for config changes
	manager.WatchConfig(func() {
		fmt.Println("Configuration changed!")
		// Reload configuration
		if err := manager.Unmarshal(&appConfig); err != nil {
			log.Printf("Failed to reload config: %v", err)
		}
	})

	// Example 6: Setting values programmatically
	manager.Set("features.maxRequests", 1000)
	manager.Set("features.allowedUsers", []string{"user1", "user2"})

	// Write configuration to file
	if err := manager.WriteConfig(); err != nil {
		log.Printf("Failed to write config: %v", err)
	}
}
