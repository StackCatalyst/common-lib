package logging

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// TraceID represents a unique identifier for request tracing
type TraceID string

// ContextKey is used for storing values in context
type ContextKey string

const (
	// TraceIDKey is the key used to store trace IDs in context
	TraceIDKey ContextKey = "trace_id"
	// UserIDKey is the key used to store user IDs in context
	UserIDKey ContextKey = "user_id"
	// RequestIDKey is the key used to store request IDs in context
	RequestIDKey ContextKey = "request_id"
)

// AdvancedConfig extends the basic Config with additional options
type AdvancedConfig struct {
	Config
	// Rotation settings
	MaxSize    int  `json:"max_size"`    // Maximum size in megabytes before rotation
	MaxBackups int  `json:"max_backups"` // Maximum number of old log files to retain
	MaxAge     int  `json:"max_age"`     // Maximum number of days to retain old log files
	Compress   bool `json:"compress"`    // Compress rotated files
	// Sampling settings
	SampleInitial    int `json:"sample_initial"`    // Sample up to n entries per second
	SampleThereafter int `json:"sample_thereafter"` // Sample every nth entry after initial
}

// DefaultAdvancedConfig returns default advanced logger configuration
func DefaultAdvancedConfig() AdvancedConfig {
	return AdvancedConfig{
		Config:           DefaultConfig(),
		MaxSize:          100,
		MaxBackups:       3,
		MaxAge:           28,
		Compress:         true,
		SampleInitial:    100,
		SampleThereafter: 100,
	}
}

// NewAdvanced creates a new logger with advanced features
func NewAdvanced(cfg AdvancedConfig) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Create the log directory if it doesn't exist
	if cfg.OutputPath != "stdout" && cfg.OutputPath != "stderr" {
		dir := filepath.Dir(cfg.OutputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Configure log rotation if not writing to stdout/stderr
	var output zapcore.WriteSyncer
	if cfg.OutputPath == "stdout" {
		output = zapcore.AddSync(os.Stdout)
	} else if cfg.OutputPath == "stderr" {
		output = zapcore.AddSync(os.Stderr)
	} else {
		output = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
	}

	encoderConfig := defaultEncoderConfig()
	var encoder zapcore.Encoder
	if cfg.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewSamplerWithOptions(
		zapcore.NewCore(encoder, output, level),
		time.Second,
		cfg.SampleInitial,
		cfg.SampleThereafter,
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Logger{zap: logger}, nil
}

// WithTracing adds request tracing to the logger
func (l *Logger) WithTracing() *Logger {
	traceID := TraceID(uuid.New().String())
	return l.With(zap.String("trace_id", string(traceID)))
}

// FromContext extracts logging fields from context
func (l *Logger) FromContext(ctx context.Context) *Logger {
	logger := l

	// Extract trace ID
	if traceID, ok := ctx.Value(TraceIDKey).(TraceID); ok {
		logger = logger.With(zap.String("trace_id", string(traceID)))
	}

	// Extract user ID
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.With(zap.String("user_id", userID))
	}

	// Extract request ID
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With(zap.String("request_id", requestID))
	}

	return logger
}

// HTTPMiddleware creates a middleware that adds request information to the logger
func (l *Logger) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate trace ID and request ID
		traceID := TraceID(uuid.New().String())
		requestID := uuid.New().String()

		// Add IDs to context
		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		ctx = context.WithValue(ctx, RequestIDKey, requestID)

		// Create request-scoped logger
		reqLogger := l.With(
			zap.String("trace_id", string(traceID)),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)

		// Log request
		reqLogger.Info("Request started")

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))

		// Log request completion
		reqLogger.Info("Request completed",
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// Audit logs an audit event with user and action information
func (l *Logger) Audit(ctx context.Context, action string, details map[string]interface{}) {
	logger := l.FromContext(ctx)

	fields := make([]zapcore.Field, 0, len(details)+1)
	fields = append(fields, zap.String("action", action))

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	logger.Info("Audit event", fields...)
}

// LogMetric logs a metric with the given name and value
func (l *Logger) LogMetric(name string, value interface{}, tags map[string]string) {
	fields := make([]zapcore.Field, 0, len(tags)+2)
	fields = append(fields, zap.String("metric_name", name))
	fields = append(fields, zap.Any("metric_value", value))

	for k, v := range tags {
		fields = append(fields, zap.String(k, v))
	}

	l.Info("Metric", fields...)
}
