package storage

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBackend implements the Backend interface for testing
type mockBackend struct {
	modules map[string]*module.Module
	content map[string][]byte
}

func newMockBackend() *mockBackend {
	return &mockBackend{
		modules: make(map[string]*module.Module),
		content: make(map[string][]byte),
	}
}

func (m *mockBackend) Store(ctx context.Context, mod *module.Module) error {
	if _, exists := m.modules[mod.ID]; exists {
		return &Error{Code: ErrAlreadyExists, Message: "module already exists"}
	}
	m.modules[mod.ID] = mod
	return nil
}

func (m *mockBackend) Get(ctx context.Context, id string) (*module.Module, error) {
	mod, exists := m.modules[id]
	if !exists {
		return nil, &Error{Code: ErrNotFound, Message: "module not found"}
	}
	return mod, nil
}

func (m *mockBackend) List(ctx context.Context, filter *module.Filter) ([]*module.Module, error) {
	var result []*module.Module
	for _, mod := range m.modules {
		if filter.Provider != "" && mod.Provider != filter.Provider {
			continue
		}
		if len(filter.Tags) > 0 {
			matches := false
			for _, tag := range filter.Tags {
				for _, modTag := range mod.Tags {
					if tag == modTag {
						matches = true
						break
					}
				}
				if matches {
					break
				}
			}
			if !matches {
				continue
			}
		}
		result = append(result, mod)
	}
	return result, nil
}

func (m *mockBackend) Delete(ctx context.Context, id string) error {
	if _, exists := m.modules[id]; !exists {
		return &Error{Code: ErrNotFound, Message: "module not found"}
	}
	delete(m.modules, id)
	delete(m.content, id)
	return nil
}

func (m *mockBackend) StoreContent(ctx context.Context, id string, content io.Reader) error {
	if _, exists := m.modules[id]; !exists {
		return &Error{Code: ErrNotFound, Message: "module not found"}
	}
	data, err := io.ReadAll(content)
	if err != nil {
		return &Error{Code: ErrInternal, Message: "failed to read content", Err: err}
	}
	m.content[id] = data
	return nil
}

func (m *mockBackend) GetContent(ctx context.Context, id string) (io.ReadCloser, error) {
	content, exists := m.content[id]
	if !exists {
		return nil, &Error{Code: ErrNotFound, Message: "content not found"}
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func TestBackend(t *testing.T) {
	ctx := context.Background()
	backend := newMockBackend()

	// Test storing a module
	mod := &module.Module{
		ID:          "test-module",
		Name:        "Test Module",
		Provider:    "aws",
		Version:     "1.0.0",
		Description: "Test module description",
		Source:      "github.com/test/module",
		Variables: []*module.Variable{
			{
				Name:        "region",
				Type:        "string",
				Description: "AWS region",
				Required:    true,
			},
		},
		Tags:      []string{"test", "aws"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := backend.Store(ctx, mod)
	require.NoError(t, err)

	// Test getting a module
	retrieved, err := backend.Get(ctx, mod.ID)
	require.NoError(t, err)
	assert.Equal(t, mod, retrieved)

	// Test storing content
	content := []byte("test content")
	err = backend.StoreContent(ctx, mod.ID, bytes.NewReader(content))
	require.NoError(t, err)

	// Test getting content
	reader, err := backend.GetContent(ctx, mod.ID)
	require.NoError(t, err)
	retrievedContent, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, retrievedContent)

	// Test listing modules
	filter := &module.Filter{
		Provider: "aws",
		Tags:     []string{"test"},
	}
	modules, err := backend.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, modules, 1)
	assert.Equal(t, mod, modules[0])

	// Test deleting a module
	err = backend.Delete(ctx, mod.ID)
	require.NoError(t, err)

	// Verify module is deleted
	_, err = backend.Get(ctx, mod.ID)
	require.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, ErrNotFound, err.(*Error).Code)
}

func TestBackendErrors(t *testing.T) {
	ctx := context.Background()
	backend := newMockBackend()

	// Test getting non-existent module
	_, err := backend.Get(ctx, "non-existent")
	require.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, ErrNotFound, err.(*Error).Code)

	// Test storing duplicate module
	mod := &module.Module{ID: "test-module"}
	err = backend.Store(ctx, mod)
	require.NoError(t, err)

	err = backend.Store(ctx, mod)
	require.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, ErrAlreadyExists, err.(*Error).Code)

	// Test storing content for non-existent module
	err = backend.StoreContent(ctx, "non-existent", bytes.NewReader([]byte("test")))
	require.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, ErrNotFound, err.(*Error).Code)

	// Test getting content for non-existent module
	_, err = backend.GetContent(ctx, "non-existent")
	require.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, ErrNotFound, err.(*Error).Code)

	// Test deleting non-existent module
	err = backend.Delete(ctx, "non-existent")
	require.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, ErrNotFound, err.(*Error).Code)
}
