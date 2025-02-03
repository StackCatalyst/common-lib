package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestAdvancedLogger(t *testing.T) {
	// Create a temporary directory for log files
	tmpDir, err := os.MkdirTemp("", "logger_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create advanced config
	cfg := DefaultAdvancedConfig()
	cfg.OutputPath = filepath.Join(tmpDir, "test.log")
	cfg.MaxSize = 1
	cfg.MaxBackups = 2
	cfg.Compress = true

	// Create logger
	logger, err := NewAdvanced(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test log rotation
	for i := 0; i < 1000; i++ {
		logger.Info("Test log rotation", zapcore.Field{
			Key: "data", Type: zapcore.StringType,
			String: "some very long string to help trigger rotation",
		})
	}

	// Check if log files were created and rotated
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.True(t, len(files) > 1, "Expected multiple log files due to rotation")
}

func TestWithTracing(t *testing.T) {
	var buf bytes.Buffer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Add tracing
	tracedLogger := logger.WithTracing()
	tracedLogger.Info("test message")

	// Parse log entry
	var logEntry map[string]interface{}
	err = json.NewDecoder(&buf).Decode(&logEntry)
	require.NoError(t, err)

	// Verify trace ID is present
	assert.NotEmpty(t, logEntry["trace_id"])
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Create context with values
	traceID := TraceID("test-trace-id")
	userID := "test-user-id"
	requestID := "test-request-id"

	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, RequestIDKey, requestID)

	// Log with context
	contextLogger := logger.FromContext(ctx)
	contextLogger.Info("test message")

	// Parse log entry
	var logEntry map[string]interface{}
	err = json.NewDecoder(&buf).Decode(&logEntry)
	require.NoError(t, err)

	// Verify context values are present
	assert.Equal(t, string(traceID), logEntry["trace_id"])
	assert.Equal(t, userID, logEntry["user_id"])
	assert.Equal(t, requestID, logEntry["request_id"])
}

func TestHTTPMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with logging middleware
	loggingHandler := logger.HTTPMiddleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Handle request
	loggingHandler.ServeHTTP(rec, req)

	// Parse log entries (there should be two: start and complete)
	decoder := json.NewDecoder(&buf)

	// Check start log
	var startLog map[string]interface{}
	err = decoder.Decode(&startLog)
	require.NoError(t, err)
	assert.Equal(t, "Request started", startLog["msg"])
	assert.NotEmpty(t, startLog["trace_id"])
	assert.NotEmpty(t, startLog["request_id"])
	assert.Equal(t, "GET", startLog["method"])
	assert.Equal(t, "/test", startLog["path"])

	// Check complete log
	var completeLog map[string]interface{}
	err = decoder.Decode(&completeLog)
	require.NoError(t, err)
	assert.Equal(t, "Request completed", completeLog["msg"])
	assert.NotEmpty(t, completeLog["duration"])
}

func TestAudit(t *testing.T) {
	var buf bytes.Buffer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Create context with user ID
	ctx := context.WithValue(context.Background(), UserIDKey, "test-user")

	// Log audit event
	details := map[string]interface{}{
		"resource_id": "res-123",
		"action_type": "update",
		"changes":     []string{"field1", "field2"},
	}
	logger.Audit(ctx, "resource_updated", details)

	// Parse log entry
	var logEntry map[string]interface{}
	err = json.NewDecoder(&buf).Decode(&logEntry)
	require.NoError(t, err)

	// Verify audit fields
	assert.Equal(t, "Audit event", logEntry["msg"])
	assert.Equal(t, "resource_updated", logEntry["action"])
	assert.Equal(t, "test-user", logEntry["user_id"])
	assert.Equal(t, "res-123", logEntry["resource_id"])
	assert.Equal(t, "update", logEntry["action_type"])
}

func TestLogMetric(t *testing.T) {
	var buf bytes.Buffer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Log metric
	tags := map[string]string{
		"service": "test-service",
		"env":     "test",
	}
	logger.LogMetric("request_duration", 123.45, tags)

	// Parse log entry
	var logEntry map[string]interface{}
	err = json.NewDecoder(&buf).Decode(&logEntry)
	require.NoError(t, err)

	// Verify metric fields
	assert.Equal(t, "Metric", logEntry["msg"])
	assert.Equal(t, "request_duration", logEntry["metric_name"])
	assert.Equal(t, 123.45, logEntry["metric_value"])
	assert.Equal(t, "test-service", logEntry["service"])
	assert.Equal(t, "test", logEntry["env"])
}
