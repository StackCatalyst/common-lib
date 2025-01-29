package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	loggingtest "github.com/StackCatalyst/common-lib/pkg/logging/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestLogger(t *testing.T) (*AuthLogger, *bytes.Buffer) {
	var buf bytes.Buffer
	logger, err := loggingtest.NewTestLogger(&buf)
	require.NoError(t, err)
	return NewAuthLogger(logger), &buf
}

func TestLogTokenValidation(t *testing.T) {
	logger, buf := setupTestLogger(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		success bool
		err     error
		check   func(t *testing.T, entry map[string]interface{})
	}{
		{
			name:    "successful validation",
			success: true,
			err:     nil,
			check: func(t *testing.T, entry map[string]interface{}) {
				assert.Equal(t, "info", entry["level"])
				assert.Equal(t, "Token validation successful", entry["msg"])
				assert.Equal(t, "token_validation", entry["event"])
				assert.Equal(t, true, entry["success"])
			},
		},
		{
			name:    "failed validation",
			success: false,
			err:     newInvalidTokenError("bad signature"),
			check: func(t *testing.T, entry map[string]interface{}) {
				assert.Equal(t, "error", entry["level"])
				assert.Equal(t, "Token validation failed", entry["msg"])
				assert.Equal(t, "token_validation", entry["event"])
				assert.Equal(t, false, entry["success"])
				assert.Contains(t, entry["error"], "bad signature")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogTokenValidation(ctx, tt.success, tt.err)

			var entry map[string]interface{}
			err := json.NewDecoder(buf).Decode(&entry)
			require.NoError(t, err)

			tt.check(t, entry)
		})
	}
}

func TestLogPermissionCheck(t *testing.T) {
	logger, buf := setupTestLogger(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		userID   string
		role     string
		resource string
		action   string
		allowed  bool
		check    func(t *testing.T, entry map[string]interface{})
	}{
		{
			name:     "allowed access",
			userID:   "user123",
			role:     "admin",
			resource: "users",
			action:   "read",
			allowed:  true,
			check: func(t *testing.T, entry map[string]interface{}) {
				assert.Equal(t, "info", entry["level"])
				assert.Equal(t, "Permission check", entry["msg"])
				assert.Equal(t, "permission_check", entry["event"])
				assert.Equal(t, "user123", entry["user_id"])
				assert.Equal(t, "admin", entry["role"])
				assert.Equal(t, "users", entry["resource"])
				assert.Equal(t, "read", entry["action"])
				assert.Equal(t, true, entry["allowed"])
			},
		},
		{
			name:     "denied access",
			userID:   "user456",
			role:     "guest",
			resource: "users",
			action:   "write",
			allowed:  false,
			check: func(t *testing.T, entry map[string]interface{}) {
				assert.Equal(t, "warn", entry["level"])
				assert.Equal(t, "Permission check", entry["msg"])
				assert.Equal(t, "permission_check", entry["event"])
				assert.Equal(t, "user456", entry["user_id"])
				assert.Equal(t, "guest", entry["role"])
				assert.Equal(t, "users", entry["resource"])
				assert.Equal(t, "write", entry["action"])
				assert.Equal(t, false, entry["allowed"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogPermissionCheck(ctx, tt.userID, tt.role, tt.resource, tt.action, tt.allowed)

			var entry map[string]interface{}
			err := json.NewDecoder(buf).Decode(&entry)
			require.NoError(t, err)

			tt.check(t, entry)
		})
	}
}

func TestLogRoleChange(t *testing.T) {
	logger, buf := setupTestLogger(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		userID   string
		role     string
		assigned bool
		err      error
		check    func(t *testing.T, entry map[string]interface{})
	}{
		{
			name:     "successful role assignment",
			userID:   "user123",
			role:     "admin",
			assigned: true,
			err:      nil,
			check: func(t *testing.T, entry map[string]interface{}) {
				assert.Equal(t, "info", entry["level"])
				assert.Equal(t, "Role assigned", entry["msg"])
				assert.Equal(t, "role_assignment", entry["event"])
				assert.Equal(t, "user123", entry["user_id"])
				assert.Equal(t, "admin", entry["role"])
				assert.Equal(t, true, entry["assigned"])
			},
		},
		{
			name:     "failed role revocation",
			userID:   "user456",
			role:     "admin",
			assigned: false,
			err:      newInvalidRoleError("admin"),
			check: func(t *testing.T, entry map[string]interface{}) {
				assert.Equal(t, "error", entry["level"])
				assert.Equal(t, "Role revocation failed", entry["msg"])
				assert.Equal(t, "role_revocation", entry["event"])
				assert.Equal(t, "user456", entry["user_id"])
				assert.Equal(t, "admin", entry["role"])
				assert.Equal(t, false, entry["assigned"])
				assert.Contains(t, entry["error"], "invalid role: admin")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogRoleChange(ctx, tt.userID, tt.role, tt.assigned, tt.err)

			var entry map[string]interface{}
			err := json.NewDecoder(buf).Decode(&entry)
			require.NoError(t, err)

			tt.check(t, entry)
		})
	}
}
