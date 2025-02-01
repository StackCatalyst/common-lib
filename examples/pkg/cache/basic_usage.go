package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/cache"
	"github.com/StackCatalyst/common-lib/pkg/metrics"
)

type User struct {
	ID       string
	Name     string
	Email    string
	LastSeen time.Time
}

func main() {
	// Example 1: Basic cache setup
	metricsReporter := metrics.New(metrics.DefaultOptions())
	cacheConfig := &cache.Config{
		Enabled:       true,
		TTL:           5 * time.Minute,
		MaxSize:       1024 * 1024 * 10, // 10MB
		PurgeInterval: time.Minute,
	}

	cache := cache.New(cacheConfig, metricsReporter)
	ctx := context.Background()

	// Example 2: Storing and retrieving simple values
	err := cache.Set(ctx, "greeting", "Hello, World!")
	if err != nil {
		log.Printf("Failed to set cache: %v", err)
	}

	var greeting string
	if found := cache.Get(ctx, "greeting", &greeting); found {
		fmt.Printf("Retrieved greeting: %s\n", greeting)
	}

	// Example 3: Storing and retrieving complex objects
	user := User{
		ID:       "user123",
		Name:     "John Doe",
		Email:    "john@example.com",
		LastSeen: time.Now(),
	}

	err = cache.Set(ctx, fmt.Sprintf("user:%s", user.ID), user)
	if err != nil {
		log.Printf("Failed to cache user: %v", err)
	}

	var cachedUser User
	if found := cache.Get(ctx, fmt.Sprintf("user:%s", user.ID), &cachedUser); found {
		fmt.Printf("Retrieved user: %+v\n", cachedUser)
	}

	// Example 4: Cache deletion
	cache.Delete(ctx, "greeting")
	if found := cache.Get(ctx, "greeting", &greeting); !found {
		fmt.Println("Greeting was successfully deleted from cache")
	}

	// Example 5: Cache clearing
	cache.Clear(ctx)
	if found := cache.Get(ctx, fmt.Sprintf("user:%s", user.ID), &cachedUser); !found {
		fmt.Println("Cache was successfully cleared")
	}

	// Example 6: Working with expiration
	err = cache.Set(ctx, "short-lived", "I will expire soon")
	if err != nil {
		log.Printf("Failed to set cache: %v", err)
	}

	// Wait for value to expire
	time.Sleep(6 * time.Minute)

	var expiredValue string
	if found := cache.Get(ctx, "short-lived", &expiredValue); !found {
		fmt.Println("Value has expired as expected")
	}
}
