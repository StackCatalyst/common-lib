package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	logger, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create logger with custom writer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Test different log levels
	tests := []struct {
		level   Level
		logFunc func(string, ...zapcore.Field)
	}{
		{Debug, logger.Debug},
		{Info, logger.Info},
		{Warn, logger.Warn},
		{Error, logger.Error},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			buf.Reset()
			msg := "test message"
			tt.logFunc(msg, zap.String("key", "value"))

			// Parse the log entry
			var logEntry map[string]interface{}
			err := json.NewDecoder(&buf).Decode(&logEntry)
			require.NoError(t, err)

			// Verify log entry
			assert.Equal(t, msg, logEntry["msg"])
			assert.Equal(t, string(tt.level), logEntry["level"])
			assert.Equal(t, "value", logEntry["key"])
		})
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	logger, err := createTestLogger(&buf)
	require.NoError(t, err)

	// Create child logger with additional fields
	childLogger := logger.With(zap.String("service", "test"))
	childLogger.Info("test message")

	// Parse the log entry
	var logEntry map[string]interface{}
	err = json.NewDecoder(&buf).Decode(&logEntry)
	require.NoError(t, err)

	// Verify fields are present
	assert.Equal(t, "test", logEntry["service"])
	assert.Equal(t, "test message", logEntry["msg"])
}

// createTestLogger creates a logger that writes to the provided io.Writer
func createTestLogger(w io.Writer) (*Logger, error) {
	encoderConfig := defaultEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(w),
		zapcore.DebugLevel,
	)

	return &Logger{zap: zap.New(core)}, nil
}

func TestNewFromZap(t *testing.T) {
	tests := []struct {
		name      string
		zapLogger *zap.Logger
		wantErr   bool
	}{
		{
			name:      "valid zap logger",
			zapLogger: zap.NewExample(),
			wantErr:   false,
		},
		{
			name:      "nil zap logger",
			zapLogger: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewFromZap(tt.zapLogger)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove("test.log")

	os.Exit(code)
}
