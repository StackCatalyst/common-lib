package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelper(t *testing.T) {
	h := NewHelper(t)

	t.Run("TempDir", func(t *testing.T) {
		dir := h.TempDir()
		assert.DirExists(t, dir)
		assert.Contains(t, h.tempDirs, dir)
	})

	t.Run("TempFile", func(t *testing.T) {
		dir := h.TempDir()
		f := h.TempFile(dir, "test-*.txt")
		assert.FileExists(t, f.Name())
	})

	t.Run("WriteFile", func(t *testing.T) {
		dir := h.TempDir()
		path := filepath.Join(dir, "test.txt")
		data := []byte("test data")
		h.WriteFile(path, data, 0644)

		// Verify file contents
		content, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, data, content)
	})

	t.Run("Cleanup", func(t *testing.T) {
		cleanupCalled := false
		h.AddCleanup(func() {
			cleanupCalled = true
		})

		h.Cleanup()
		assert.True(t, cleanupCalled)

		// Verify temp directories are cleaned up
		for _, dir := range h.tempDirs {
			assert.NoDirExists(t, dir)
		}
	})

	t.Run("Assertions", func(t *testing.T) {
		// Test RequireNoError
		h.RequireNoError(nil)

		// Test RequireError
		h.RequireError(assert.AnError)

		// Test RequireEqual
		h.RequireEqual("test", "test")

		// Test RequireNotEqual
		h.RequireNotEqual("test", "other")

		// Test RequireTrue
		h.RequireTrue(true)

		// Test RequireFalse
		h.RequireFalse(false)

		// Test RequireNil
		h.RequireNil(nil)

		// Test RequireNotNil
		h.RequireNotNil("not nil")
	})
}

func TestHelperCleanupOrder(t *testing.T) {
	h := NewHelper(t)
	order := make([]int, 0)

	h.AddCleanup(func() {
		order = append(order, 1)
	})
	h.AddCleanup(func() {
		order = append(order, 2)
	})
	h.AddCleanup(func() {
		order = append(order, 3)
	})

	h.Cleanup()

	// Verify cleanup functions are called in reverse order
	assert.Equal(t, []int{3, 2, 1}, order)
}
