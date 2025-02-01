package validation

import (
	"context"
	"testing"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyValidator(t *testing.T) {
	validator := NewDependencyValidator()
	ctx := context.Background()

	t.Run("valid dependencies", func(t *testing.T) {
		mod := &module.Module{
			Name:    "test-module",
			Version: "1.0.0",
			Dependencies: []*module.Dependency{
				{
					Name:     "dep1",
					Source:   "github.com/test/dep1",
					Version:  ">=1.0.0",
					Required: true,
				},
				{
					Name:     "dep2",
					Source:   "github.com/test/dep2",
					Version:  "~2.0.0",
					Required: false,
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid version format", func(t *testing.T) {
		mod := &module.Module{
			Name:    "test-module",
			Version: "invalid-version",
			Dependencies: []*module.Dependency{
				{
					Name:    "dep1",
					Source:  "github.com/test/dep1",
					Version: "1.x",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)

		var fields []string
		for _, err := range result.Errors {
			fields = append(fields, err.Field)
		}

		assert.Contains(t, fields, "version")
		assert.Contains(t, fields, "dependencies[0].version")
	})

	t.Run("missing required fields", func(t *testing.T) {
		mod := &module.Module{
			Name:    "test-module",
			Version: "1.0.0",
			Dependencies: []*module.Dependency{
				{
					Version: ">=1.0.0",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)

		var fields []string
		for _, err := range result.Errors {
			fields = append(fields, err.Field)
		}

		assert.Contains(t, fields, "dependencies[0].name")
		assert.Contains(t, fields, "dependencies[0].source")
	})

	t.Run("circular dependency", func(t *testing.T) {
		mod := &module.Module{
			Name:    "test-module",
			Version: "1.0.0",
			Dependencies: []*module.Dependency{
				{
					Name:    "test-module",
					Source:  "github.com/test/test-module",
					Version: ">=1.0.0",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "dependencies[0].name")
		assert.Contains(t, result.Errors[0].Message, "circular dependency")
	})

	t.Run("duplicate dependencies", func(t *testing.T) {
		mod := &module.Module{
			Name:    "test-module",
			Version: "1.0.0",
			Dependencies: []*module.Dependency{
				{
					Name:    "dep1",
					Source:  "github.com/test/dep1",
					Version: ">=1.0.0",
				},
				{
					Name:    "dep1",
					Source:  "github.com/test/dep1",
					Version: ">=2.0.0",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "dependencies[1].name")
		assert.Contains(t, result.Errors[0].Message, "duplicate dependency")
	})
}
