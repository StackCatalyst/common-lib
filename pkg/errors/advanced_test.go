package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryableError(t *testing.T) {
	err := New(ErrInternal, "test error")
	retryDuration := 100 * time.Millisecond // Use a shorter duration for tests
	retryable := NewRetryable(err, 3, retryDuration)

	// Initial state
	assert.True(t, retryable.CanRetry())
	assert.Equal(t, 0, retryable.RetryCount)

	// After first attempt
	retryable.Attempt()
	assert.False(t, retryable.CanRetry()) // Should be false immediately after attempt
	assert.Equal(t, 1, retryable.RetryCount)

	// After waiting for retry duration
	time.Sleep(retryDuration + 10*time.Millisecond) // Add a small buffer
	assert.True(t, retryable.CanRetry())

	// After second attempt
	retryable.Attempt()
	time.Sleep(retryDuration + 10*time.Millisecond) // Wait after second attempt
	assert.True(t, retryable.CanRetry())

	// After third attempt (max retries)
	retryable.Attempt()
	assert.False(t, retryable.CanRetry())
	assert.Equal(t, 3, retryable.RetryCount)
}

func TestErrorGroup(t *testing.T) {
	group := NewErrorGroup()
	assert.False(t, group.HasErrors())

	err1 := New(ErrNotFound, "error 1")
	err2 := New(ErrValidation, "error 2")

	group.Add(err1)
	group.Add(err2)
	group.Add(nil) // Should handle nil errors

	assert.True(t, group.HasErrors())
	assert.Contains(t, group.Error(), "error 1")
	assert.Contains(t, group.Error(), "error 2")
}

func TestErrorContext(t *testing.T) {
	ctx := context.Background()
	err := New(ErrInternal, "test error")

	ctxErr := WithContext(ctx, err)
	assert.NotNil(t, ctxErr)

	// Get context back
	gotCtx, ok := GetErrorContext(ctxErr)
	assert.True(t, ok)
	assert.Equal(t, ctx, gotCtx)

	// Try with non-context error
	gotCtx, ok = GetErrorContext(err)
	assert.False(t, ok)
	assert.Nil(t, gotCtx)
}

func TestErrorStack(t *testing.T) {
	err := New(ErrInternal, "test error")
	stackErr := NewErrorStack(err)

	stack := stackErr.Stack()
	assert.NotEmpty(t, stack)

	// Check first frame
	frame := stack[0]
	assert.Contains(t, frame.File, "advanced_test.go")
	assert.Contains(t, frame.Function, "TestErrorStack")
	assert.True(t, frame.Line > 0)

	// String representation
	str := frame.String()
	assert.Contains(t, str, "advanced_test.go")
	assert.Contains(t, str, "TestErrorStack")
}
