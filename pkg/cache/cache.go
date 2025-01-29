package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Config holds the cache configuration
type Config struct {
	// Enabled indicates if caching is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`
	// TTL is the default time-to-live for cache entries
	TTL time.Duration `json:"ttl" yaml:"ttl"`
	// MaxSize is the maximum cache size in bytes
	MaxSize int64 `json:"max_size" yaml:"max_size"`
	// PurgeInterval is how often to check for expired entries
	PurgeInterval time.Duration `json:"purge_interval" yaml:"purge_interval"`
}

// DefaultConfig returns the default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		TTL:           time.Hour,
		MaxSize:       1024 * 1024 * 1024, // 1GB
		PurgeInterval: time.Minute * 5,
	}
}

// entry represents a cache entry
type entry struct {
	value     []byte
	size      int64
	expiresAt time.Time
}

// Cache represents an in-memory cache with TTL and size limits
type Cache struct {
	config     *Config
	mu         sync.RWMutex
	data       map[string]*entry
	totalBytes int64

	// Metrics
	hits        *prometheus.CounterVec
	misses      *prometheus.CounterVec
	sizeMetric  *prometheus.GaugeVec
	itemsMetric *prometheus.GaugeVec
}

// New creates a new cache instance
func New(config *Config, metricsReporter *metrics.Reporter) *Cache {
	if config == nil {
		config = DefaultConfig()
	}

	c := &Cache{
		config: config,
		data:   make(map[string]*entry),
		hits: metricsReporter.Counter("cache_hits_total",
			"Total number of cache hits",
			[]string{"cache"}),
		misses: metricsReporter.Counter("cache_misses_total",
			"Total number of cache misses",
			[]string{"cache"}),
		sizeMetric: metricsReporter.Gauge("cache_size_bytes",
			"Current size of cache in bytes",
			[]string{"cache"}),
		itemsMetric: metricsReporter.Gauge("cache_items_total",
			"Total number of items in cache",
			[]string{"cache"}),
	}

	// Start background cleanup if enabled
	if config.Enabled && config.PurgeInterval > 0 {
		go c.startCleanup(context.Background())
	}

	return c
}

// Set stores a value in the cache
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	if !c.config.Enabled {
		return nil
	}

	// Convert value to bytes
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	size := int64(len(data))
	if size > c.config.MaxSize {
		return fmt.Errorf("value size %d exceeds maximum cache size %d", size, c.config.MaxSize)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to make room
	if c.totalBytes+size > c.config.MaxSize {
		c.evict(size)
	}

	// Store the entry
	c.data[key] = &entry{
		value:     data,
		size:      size,
		expiresAt: time.Now().Add(c.config.TTL),
	}

	c.totalBytes += size
	c.itemsMetric.WithLabelValues("memory").Inc()
	c.sizeMetric.WithLabelValues("memory").Set(float64(c.totalBytes))

	return nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(ctx context.Context, key string, value interface{}) bool {
	if !c.config.Enabled {
		c.misses.WithLabelValues("memory").Inc()
		return false
	}

	c.mu.RLock()
	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expiresAt) {
		c.mu.RUnlock()
		c.misses.WithLabelValues("memory").Inc()
		return false
	}
	data := entry.value
	c.mu.RUnlock()

	if err := json.Unmarshal(data, value); err != nil {
		c.misses.WithLabelValues("memory").Inc()
		return false
	}

	c.hits.WithLabelValues("memory").Inc()
	return true
}

// Delete removes a value from the cache
func (c *Cache) Delete(ctx context.Context, key string) {
	if !c.config.Enabled {
		return
	}

	c.mu.Lock()
	if entry, exists := c.data[key]; exists {
		c.totalBytes -= entry.size
		delete(c.data, key)
		c.itemsMetric.WithLabelValues("memory").Dec()
		c.sizeMetric.WithLabelValues("memory").Set(float64(c.totalBytes))
	}
	c.mu.Unlock()
}

// Clear removes all values from the cache
func (c *Cache) Clear(ctx context.Context) {
	if !c.config.Enabled {
		return
	}

	c.mu.Lock()
	c.data = make(map[string]*entry)
	c.totalBytes = 0
	c.itemsMetric.WithLabelValues("memory").Set(0)
	c.sizeMetric.WithLabelValues("memory").Set(0)
	c.mu.Unlock()
}

// evict removes entries to make room for the requested size
func (c *Cache) evict(needed int64) {
	// Simple LRU-like eviction: remove oldest entries first
	for key, entry := range c.data {
		if c.totalBytes+needed <= c.config.MaxSize {
			break
		}
		c.totalBytes -= entry.size
		delete(c.data, key)
		c.itemsMetric.WithLabelValues("memory").Dec()
	}
}

// startCleanup runs periodic cleanup of expired entries
func (c *Cache) startCleanup(ctx context.Context) {
	ticker := time.NewTicker(c.config.PurgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	now := time.Now()
	c.mu.Lock()
	for key, entry := range c.data {
		if now.After(entry.expiresAt) {
			c.totalBytes -= entry.size
			delete(c.data, key)
			c.itemsMetric.WithLabelValues("memory").Dec()
		}
	}
	c.sizeMetric.WithLabelValues("memory").Set(float64(c.totalBytes))
	c.mu.Unlock()
}
