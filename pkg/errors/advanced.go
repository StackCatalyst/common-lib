package errors

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// RetryableError represents an error that can be retried
type RetryableError struct {
	*AppError
	MaxRetries  int
	RetryAfter  time.Duration
	RetryCount  int
	LastAttempt time.Time
}

// NewRetryable creates a new RetryableError
func NewRetryable(err *AppError, maxRetries int, retryAfter time.Duration) *RetryableError {
	return &RetryableError{
		AppError:    err,
		MaxRetries:  maxRetries,
		RetryAfter:  retryAfter,
		RetryCount:  0,
		LastAttempt: time.Now().Add(-2 * retryAfter),
	}
}

// CanRetry checks if the error can be retried
func (e *RetryableError) CanRetry() bool {
	return e.RetryCount < e.MaxRetries && time.Since(e.LastAttempt) >= e.RetryAfter
}

// Attempt increments the retry count and updates the last attempt time
func (e *RetryableError) Attempt() {
	e.RetryCount++
	e.LastAttempt = time.Now()
}

// ErrorGroup represents a collection of errors
type ErrorGroup struct {
	mu     sync.Mutex
	errors []error
}

// NewErrorGroup creates a new ErrorGroup
func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{
		errors: make([]error, 0),
	}
}

// Add adds an error to the group
func (g *ErrorGroup) Add(err error) {
	if err == nil {
		return
	}
	g.mu.Lock()
	g.errors = append(g.errors, err)
	g.mu.Unlock()
}

// HasErrors checks if the group contains any errors
func (g *ErrorGroup) HasErrors() bool {
	return len(g.errors) > 0
}

// Error implements the error interface
func (g *ErrorGroup) Error() string {
	if !g.HasErrors() {
		return ""
	}

	if len(g.errors) == 1 {
		return g.errors[0].Error()
	}

	msg := fmt.Sprintf("%d errors occurred:\n", len(g.errors))
	for i, err := range g.errors {
		msg += fmt.Sprintf("  %d) %s\n", i+1, err.Error())
	}
	return msg
}

// ErrorContext adds context to errors
type ErrorContext struct {
	context.Context
	err error
}

// NewErrorContext creates a new ErrorContext
func NewErrorContext(ctx context.Context, err error) *ErrorContext {
	return &ErrorContext{
		Context: ctx,
		err:     err,
	}
}

// Error implements the error interface
func (c *ErrorContext) Error() string {
	return c.err.Error()
}

// GetContext returns the context
func (c *ErrorContext) GetContext() context.Context {
	return c.Context
}

// WithContext wraps an error with context
func WithContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	return NewErrorContext(ctx, err)
}

// GetErrorContext extracts context from an error
func GetErrorContext(err error) (context.Context, bool) {
	var errCtx *ErrorContext
	if errors.As(err, &errCtx) {
		return errCtx.GetContext(), true
	}
	return nil, false
}

// StackFrame represents a single stack frame
type StackFrame struct {
	File     string
	Line     int
	Function string
}

// String returns a string representation of the stack frame
func (f StackFrame) String() string {
	return fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Function)
}

// ErrorStack represents an error with a stack trace
type ErrorStack struct {
	*AppError
	stack []StackFrame
}

// NewErrorStack creates a new ErrorStack
func NewErrorStack(err *AppError) *ErrorStack {
	return &ErrorStack{
		AppError: err,
		stack:    callers(),
	}
}

// Stack returns the error's stack trace
func (e *ErrorStack) Stack() []StackFrame {
	return e.stack
}

// callers returns the current stack trace
func callers() []StackFrame {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	stack := make([]StackFrame, 0, n)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		if strings.Contains(frame.Function, "runtime.") {
			continue
		}
		stack = append(stack, StackFrame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		})
	}
	return stack
}
