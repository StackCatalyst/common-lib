package testing

import (
	"io"

	"github.com/StackCatalyst/common-lib/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewTestLogger creates a new logger instance for testing that writes to the provided writer
func NewTestLogger(w io.Writer) (*logging.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
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

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(w),
		zapcore.DebugLevel,
	)

	return logging.NewFromZap(zap.New(core))
}
