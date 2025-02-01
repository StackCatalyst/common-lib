package main

import (
	"context"
	"os"

	"github.com/StackCatalyst/common-lib/pkg/logging"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Create default configuration
	cfg := logging.DefaultConfig()
	cfg.Level = logging.Debug
	cfg.OutputPath = "app.log"
	cfg.Encoding = "json"

	// Create a new logger
	logger, err := logging.New(cfg)
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	// Example 1: Basic logging with different levels
	logger.Debug("Debug message", zapcore.Field{Key: "example", Type: zapcore.StringType, String: "value"})
	logger.Info("Info message", zapcore.Field{Key: "user_id", Type: zapcore.StringType, String: "123"})
	logger.Warn("Warning message", zapcore.Field{Key: "attempt", Type: zapcore.Int64Type, Integer: 3})
	logger.Error("Error message", zapcore.Field{Key: "error_code", Type: zapcore.StringType, String: "AUTH001"})

	// Example 2: Logging with context
	ctx := context.Background()
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("Processing request", zapcore.Field{Key: "request_id", Type: zapcore.StringType, String: "req-123"})

	// Example 3: Creating a child logger with additional fields
	userLogger := logger.With(
		zapcore.Field{Key: "user_id", Type: zapcore.StringType, String: "user-123"},
		zapcore.Field{Key: "service", Type: zapcore.StringType, String: "auth"},
	)
	userLogger.Info("User logged in")
	userLogger.Debug("User session created")
}
