package auth

import (
	"errors"
	"testing"

	liberrors "github.com/StackCatalyst/common-lib/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorCreation(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		checkFn func(error) bool
		message string
		errCode liberrors.ErrorCode
	}{
		{
			name:    "invalid token error",
			err:     newInvalidTokenError("invalid signature"),
			checkFn: IsInvalidTokenError,
			message: "invalid signature",
			errCode: ErrInvalidToken,
		},
		{
			name:    "token expired error",
			err:     newTokenExpiredError(),
			checkFn: IsTokenExpiredError,
			message: "token has expired",
			errCode: ErrTokenExpired,
		},
		{
			name:    "missing token error",
			err:     newMissingTokenError(),
			checkFn: IsMissingTokenError,
			message: "authentication token is missing",
			errCode: ErrMissingToken,
		},
		{
			name:    "invalid role error",
			err:     newInvalidRoleError("unknown"),
			checkFn: IsInvalidRoleError,
			message: "invalid role: unknown",
			errCode: ErrInvalidRole,
		},
		{
			name:    "invalid resource error",
			err:     newInvalidResourceError("unknown"),
			checkFn: IsInvalidResourceError,
			message: "invalid resource: unknown",
			errCode: ErrInvalidResource,
		},
		{
			name:    "invalid action error",
			err:     newInvalidActionError("unknown"),
			checkFn: IsInvalidActionError,
			message: "invalid action: unknown",
			errCode: ErrInvalidAction,
		},
		{
			name:    "permission denied error",
			err:     newPermissionDeniedError("access denied"),
			checkFn: IsPermissionDeniedError,
			message: "access denied",
			errCode: ErrPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check error type
			assert.True(t, tt.checkFn(tt.err))

			// Check error message
			assert.Contains(t, tt.err.Error(), tt.message)

			// Check error code
			var appErr *liberrors.AppError
			assert.True(t, errors.As(tt.err, &appErr))
			assert.Equal(t, tt.errCode, appErr.Code)
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test token error wrapping
	t.Run("wrap token error", func(t *testing.T) {
		originalErr := newInvalidTokenError("bad format")
		wrappedErr := wrapTokenError(originalErr, "failed to validate")

		assert.True(t, IsInvalidTokenError(wrappedErr))
		assert.Contains(t, wrappedErr.Error(), "failed to validate")
		assert.Contains(t, wrappedErr.Error(), "bad format")
	})

	// Test authorization error wrapping
	t.Run("wrap authorization error", func(t *testing.T) {
		originalErr := newInvalidRoleError("guest")
		wrappedErr := wrapAuthorizationError(originalErr, "access check failed")

		assert.True(t, IsPermissionDeniedError(wrappedErr))
		assert.Contains(t, wrappedErr.Error(), "access check failed")
		assert.Contains(t, wrappedErr.Error(), "invalid role: guest")
	})
}

func TestErrorChecking(t *testing.T) {
	// Test that error checking functions work with wrapped errors
	t.Run("check wrapped errors", func(t *testing.T) {
		originalErr := newInvalidTokenError("bad signature")
		wrappedErr := liberrors.Wrap(originalErr, liberrors.ErrInternal, "processing failed")

		assert.True(t, IsInvalidTokenError(wrappedErr))
		assert.False(t, IsTokenExpiredError(wrappedErr))
	})

	// Test that error checking functions work with nil errors
	t.Run("check nil error", func(t *testing.T) {
		assert.False(t, IsInvalidTokenError(nil))
		assert.False(t, IsTokenExpiredError(nil))
		assert.False(t, IsMissingTokenError(nil))
		assert.False(t, IsInvalidRoleError(nil))
		assert.False(t, IsInvalidResourceError(nil))
		assert.False(t, IsInvalidActionError(nil))
		assert.False(t, IsPermissionDeniedError(nil))
	})
}
