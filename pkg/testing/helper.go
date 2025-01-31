package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Helper provides utility functions for testing
type Helper struct {
	t *testing.T
	// cleanup functions to be called when the test ends
	cleanup []func()
	// tempDirs holds paths to temporary directories created during the test
	tempDirs []string
}

// NewHelper creates a new test helper
func NewHelper(t *testing.T) *Helper {
	h := &Helper{
		t:       t,
		cleanup: make([]func(), 0),
	}

	// Register cleanup function to run when the test ends
	t.Cleanup(h.Cleanup)
	return h
}

// Cleanup runs all registered cleanup functions in reverse order
func (h *Helper) Cleanup() {
	// Run cleanup functions in reverse order
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		h.cleanup[i]()
	}

	// Clean up temporary directories
	for _, dir := range h.tempDirs {
		os.RemoveAll(dir)
	}
}

// AddCleanup adds a cleanup function to be called when the test ends
func (h *Helper) AddCleanup(fn func()) {
	h.cleanup = append(h.cleanup, fn)
}

// TempDir creates a new temporary directory and registers it for cleanup
func (h *Helper) TempDir() string {
	dir, err := os.MkdirTemp("", "test-*")
	require.NoError(h.t, err)
	h.tempDirs = append(h.tempDirs, dir)
	return dir
}

// TempFile creates a new temporary file and registers it for cleanup
func (h *Helper) TempFile(dir, pattern string) *os.File {
	f, err := os.CreateTemp(dir, pattern)
	require.NoError(h.t, err)
	h.AddCleanup(func() {
		f.Close()
		os.Remove(f.Name())
	})
	return f
}

// WriteFile writes data to a file and registers it for cleanup
func (h *Helper) WriteFile(path string, data []byte, perm os.FileMode) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(h.t, err)
	err = os.WriteFile(path, data, perm)
	require.NoError(h.t, err)
	h.AddCleanup(func() {
		os.Remove(path)
	})
}

// RequireNoError asserts that err is nil
func (h *Helper) RequireNoError(err error, msgAndArgs ...interface{}) {
	require.NoError(h.t, err, msgAndArgs...)
}

// RequireError asserts that err is not nil
func (h *Helper) RequireError(err error, msgAndArgs ...interface{}) {
	require.Error(h.t, err, msgAndArgs...)
}

// RequireEqual asserts that expected and actual are equal
func (h *Helper) RequireEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	require.Equal(h.t, expected, actual, msgAndArgs...)
}

// RequireNotEqual asserts that expected and actual are not equal
func (h *Helper) RequireNotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	require.NotEqual(h.t, expected, actual, msgAndArgs...)
}

// RequireTrue asserts that value is true
func (h *Helper) RequireTrue(value bool, msgAndArgs ...interface{}) {
	require.True(h.t, value, msgAndArgs...)
}

// RequireFalse asserts that value is false
func (h *Helper) RequireFalse(value bool, msgAndArgs ...interface{}) {
	require.False(h.t, value, msgAndArgs...)
}

// RequireNil asserts that object is nil
func (h *Helper) RequireNil(object interface{}, msgAndArgs ...interface{}) {
	require.Nil(h.t, object, msgAndArgs...)
}

// RequireNotNil asserts that object is not nil
func (h *Helper) RequireNotNil(object interface{}, msgAndArgs ...interface{}) {
	require.NotNil(h.t, object, msgAndArgs...)
}
