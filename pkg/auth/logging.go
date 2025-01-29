package auth

import (
	"context"

	"github.com/StackCatalyst/common-lib/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogEvent represents different authentication events that can be logged
type LogEvent string

const (
	// Authentication events
	EventTokenValidation LogEvent = "token_validation"
	EventTokenCreation   LogEvent = "token_creation"
	EventTokenRefresh    LogEvent = "token_refresh"
	EventTokenRevocation LogEvent = "token_revocation"

	// Authorization events
	EventPermissionCheck LogEvent = "permission_check"
	EventRoleAssignment  LogEvent = "role_assignment"
	EventRoleRevocation  LogEvent = "role_revocation"
)

// AuthLogger wraps the common logger with auth-specific functionality
type AuthLogger struct {
	logger *logging.Logger
}

// NewAuthLogger creates a new AuthLogger instance
func NewAuthLogger(logger *logging.Logger) *AuthLogger {
	return &AuthLogger{
		logger: logger.With(zap.String("component", "auth")),
	}
}

// logAuthEvent logs an authentication event with structured data
func (l *AuthLogger) logAuthEvent(ctx context.Context, event LogEvent, level logging.Level, msg string, fields ...zapcore.Field) {
	// Add event type to fields
	fields = append(fields, zap.String("event", string(event)))

	// Get logger with context
	ctxLogger := l.logger.WithContext(ctx)

	// Log at appropriate level
	switch level {
	case logging.Debug:
		ctxLogger.Debug(msg, fields...)
	case logging.Info:
		ctxLogger.Info(msg, fields...)
	case logging.Warn:
		ctxLogger.Warn(msg, fields...)
	case logging.Error:
		ctxLogger.Error(msg, fields...)
	}
}

// LogTokenValidation logs token validation events
func (l *AuthLogger) LogTokenValidation(ctx context.Context, success bool, err error) {
	fields := []zapcore.Field{
		zap.Bool("success", success),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	level := logging.Info
	msg := "Token validation successful"
	if !success {
		level = logging.Error
		msg = "Token validation failed"
	}

	l.logAuthEvent(ctx, EventTokenValidation, level, msg, fields...)
}

// LogTokenCreation logs token creation events
func (l *AuthLogger) LogTokenCreation(ctx context.Context, userID string, success bool, err error) {
	fields := []zapcore.Field{
		zap.String("user_id", userID),
		zap.Bool("success", success),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	level := logging.Info
	msg := "Token created successfully"
	if !success {
		level = logging.Error
		msg = "Token creation failed"
	}

	l.logAuthEvent(ctx, EventTokenCreation, level, msg, fields...)
}

// LogPermissionCheck logs authorization check events
func (l *AuthLogger) LogPermissionCheck(ctx context.Context, userID, role, resource, action string, allowed bool) {
	fields := []zapcore.Field{
		zap.String("user_id", userID),
		zap.String("role", role),
		zap.String("resource", resource),
		zap.String("action", action),
		zap.Bool("allowed", allowed),
	}

	level := logging.Info
	msg := "Permission check"
	if !allowed {
		level = logging.Warn
	}

	l.logAuthEvent(ctx, EventPermissionCheck, level, msg, fields...)
}

// LogRoleChange logs role assignment or revocation events
func (l *AuthLogger) LogRoleChange(ctx context.Context, userID, role string, assigned bool, err error) {
	fields := []zapcore.Field{
		zap.String("user_id", userID),
		zap.String("role", role),
		zap.Bool("assigned", assigned),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	event := EventRoleAssignment
	msg := "Role assigned"
	if !assigned {
		event = EventRoleRevocation
		msg = "Role revoked"
	}

	level := logging.Info
	if err != nil {
		level = logging.Error
		if assigned {
			msg = "Role assignment failed"
		} else {
			msg = "Role revocation failed"
		}
	}

	l.logAuthEvent(ctx, event, level, msg, fields...)
}
