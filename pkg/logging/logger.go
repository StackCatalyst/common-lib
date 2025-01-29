package logging

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level represents the logging level
type Level string

const (
	// Debug level for development
	Debug Level = "debug"
	// Info level for general operational entries
	Info Level = "info"
	// Warn level for non-critical issues
	Warn Level = "warn"
	// Error level for errors that need attention
	Error Level = "error"
)

// Logger wraps zap logger with additional functionality
type Logger struct {
	zap *zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level      Level  `json:"level"`
	OutputPath string `json:"output_path"`
	Encoding   string `json:"encoding"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      Info,
		OutputPath: "stdout",
		Encoding:   "json",
	}
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         cfg.Encoding,
		EncoderConfig:    defaultEncoderConfig(),
		OutputPaths:      []string{cfg.OutputPath},
		ErrorOutputPaths: []string{cfg.OutputPath},
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{zap: logger}, nil
}

// NewFromZap creates a new logger instance from an existing zap logger
func NewFromZap(zapLogger *zap.Logger) (*Logger, error) {
	if zapLogger == nil {
		return nil, errors.New("zap logger cannot be nil")
	}
	return &Logger{zap: zapLogger}, nil
}

// With creates a child logger with additional fields
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs a message at info level
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, fields ...zapcore.Field) {
	l.zap.Error(msg, fields...)
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// TODO: Extract relevant fields from context
	return l
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

func parseLevel(level Level) (zapcore.Level, error) {
	switch level {
	case Debug:
		return zapcore.DebugLevel, nil
	case Info:
		return zapcore.InfoLevel, nil
	case Warn:
		return zapcore.WarnLevel, nil
	case Error:
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

func defaultEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
