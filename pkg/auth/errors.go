package auth

import (
	"errors"

	liberrors "github.com/StackCatalyst/common-lib/pkg/errors"
)

// Auth-specific error codes
const (
	ErrInvalidToken     liberrors.ErrorCode = "INVALID_TOKEN"
	ErrTokenExpired     liberrors.ErrorCode = "TOKEN_EXPIRED"
	ErrMissingToken     liberrors.ErrorCode = "MISSING_TOKEN"
	ErrInvalidRole      liberrors.ErrorCode = "INVALID_ROLE"
	ErrInvalidResource  liberrors.ErrorCode = "INVALID_RESOURCE"
	ErrInvalidAction    liberrors.ErrorCode = "INVALID_ACTION"
	ErrPermissionDenied liberrors.ErrorCode = "PERMISSION_DENIED"
)

// Common error creation functions
func newInvalidTokenError(msg string) error {
	return liberrors.New(ErrInvalidToken, msg)
}

func newTokenExpiredError() error {
	return liberrors.New(ErrTokenExpired, "token has expired")
}

func newMissingTokenError() error {
	return liberrors.New(ErrMissingToken, "authentication token is missing")
}

func newInvalidRoleError(role string) error {
	return liberrors.New(ErrInvalidRole, "invalid role: "+role)
}

func newInvalidResourceError(resource string) error {
	return liberrors.New(ErrInvalidResource, "invalid resource: "+resource)
}

func newInvalidActionError(action string) error {
	return liberrors.New(ErrInvalidAction, "invalid action: "+action)
}

func newPermissionDeniedError(msg string) error {
	return liberrors.New(ErrPermissionDenied, msg)
}

// Error wrapping functions
func wrapTokenError(err error, msg string) error {
	return liberrors.Wrap(err, ErrInvalidToken, msg)
}

func wrapAuthorizationError(err error, msg string) error {
	return liberrors.Wrap(err, ErrPermissionDenied, msg)
}

// Error checking functions
func IsInvalidTokenError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrInvalidToken {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func IsTokenExpiredError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrTokenExpired {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func IsMissingTokenError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrMissingToken {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func IsInvalidRoleError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrInvalidRole {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func IsInvalidResourceError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrInvalidResource {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func IsInvalidActionError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrInvalidAction {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func IsPermissionDeniedError(err error) bool {
	var appErr *liberrors.AppError
	for err != nil {
		if errors.As(err, &appErr) && appErr.Code == ErrPermissionDenied {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}
