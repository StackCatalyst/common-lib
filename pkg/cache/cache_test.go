package cache

import (
	"context"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetricsReporter() *metrics.Reporter {
	registry := prometheus.NewRegistry()
	return metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "cache",
		Registry:  registry,
	})
}

func TestCache(t *testing.T) {
	ctx := context.Background()
	metricsReporter := newTestMetricsReporter()
	config := &Config{
		Enabled:       true,
		TTL:           time.Second,
		MaxSize:       1024,
		PurgeInterval: time.Millisecond * 100,
	}

	cache := New(config, metricsReporter)
	require.NotNil(t, cache)

	// Test setting and getting values
	value := map[string]interface{}{
		"key": "value",
		"num": float64(123), // Use float64 for JSON compatibility
	}

	err := cache.Set(ctx, "test", value)
	require.NoError(t, err)

	var retrieved map[string]interface{}
	ok := cache.Get(ctx, "test", &retrieved)
	require.True(t, ok)
	assert.Equal(t, value, retrieved)

	// Test cache miss
	ok = cache.Get(ctx, "nonexistent", &retrieved)
	assert.False(t, ok)

	// Test expiration
	time.Sleep(time.Second * 2)
	ok = cache.Get(ctx, "test", &retrieved)
	assert.False(t, ok)

	// Test deletion
	err = cache.Set(ctx, "test", value)
	require.NoError(t, err)
	cache.Delete(ctx, "test")
	ok = cache.Get(ctx, "test", &retrieved)
	assert.False(t, ok)

	// Test clearing
	err = cache.Set(ctx, "test1", value)
	require.NoError(t, err)
	err = cache.Set(ctx, "test2", value)
	require.NoError(t, err)
	cache.Clear(ctx)
	ok = cache.Get(ctx, "test1", &retrieved)
	assert.False(t, ok)
	ok = cache.Get(ctx, "test2", &retrieved)
	assert.False(t, ok)
}

func TestCacheEviction(t *testing.T) {
	ctx := context.Background()
	metricsReporter := newTestMetricsReporter()
	config := &Config{
		Enabled:       true,
		TTL:           time.Hour,
		MaxSize:       100, // Small size to force eviction
		PurgeInterval: time.Hour,
	}

	cache := New(config, metricsReporter)
	require.NotNil(t, cache)

	// Add items until eviction occurs
	value := map[string]string{"data": "this is a long string that will exceed the cache size limit"}

	// First item should fit
	err := cache.Set(ctx, "item1", value)
	require.NoError(t, err)

	// Second item should trigger eviction of first item
	err = cache.Set(ctx, "item2", value)
	require.NoError(t, err)

	// First item should be evicted
	var retrieved map[string]string
	ok := cache.Get(ctx, "item1", &retrieved)
	assert.False(t, ok)

	// Second item should still be present
	ok = cache.Get(ctx, "item2", &retrieved)
	assert.True(t, ok)
	assert.Equal(t, value, retrieved)
}

func TestCacheDisabled(t *testing.T) {
	ctx := context.Background()
	metricsReporter := newTestMetricsReporter()
	config := &Config{
		Enabled: false,
	}

	cache := New(config, metricsReporter)
	require.NotNil(t, cache)

	value := "test"
	err := cache.Set(ctx, "test", value)
	require.NoError(t, err)

	var retrieved string
	ok := cache.Get(ctx, "test", &retrieved)
	assert.False(t, ok)
}
