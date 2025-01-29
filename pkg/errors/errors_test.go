package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(ErrNotFound, "resource not found")
	assert.Equal(t, ErrNotFound, err.Code)
	assert.Equal(t, "resource not found", err.Message)
	assert.Nil(t, err.Err)
	assert.Equal(t, "NOT_FOUND: resource not found", err.Error())
}

func TestWrap(t *testing.T) {
	originalErr := New(ErrValidation, "invalid input")
	wrappedErr := Wrap(originalErr, ErrInternal, "processing failed")

	assert.Equal(t, ErrInternal, wrappedErr.Code)
	assert.Equal(t, "processing failed", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Err)
	assert.Equal(t, "INTERNAL: processing failed: VALIDATION: invalid input", wrappedErr.Error())
}

func TestIs(t *testing.T) {
	err := New(ErrNotFound, "resource not found")
	assert.True(t, Is(err, ErrNotFound))
	assert.False(t, Is(err, ErrInternal))

	wrapped := Wrap(err, ErrInternal, "processing failed")
	assert.True(t, Is(wrapped, ErrInternal))
}
