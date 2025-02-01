package validation

import (
	"context"
	"testing"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator(t *testing.T) {
	validator := NewSchemaValidator()
	ctx := context.Background()

	t.Run("valid schema", func(t *testing.T) {
		mod := &module.Module{
			ID:      "test-module",
			Name:    "Test Module",
			Version: "1.0.0",
			Variables: []*module.Variable{
				{
					Name:        "var1",
					Type:        "string",
					Description: "Test variable",
					Required:    true,
					Validation: &module.Validation{
						Pattern: "^test.*",
					},
				},
			},
			Resources: []*module.Resource{
				{
					Type:        "test_resource",
					Provider:    "test_provider",
					Description: "Test resource",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid ID format", func(t *testing.T) {
		mod := &module.Module{
			ID:      "test@module",
			Name:    "Test Module",
			Version: "1.0.0",
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
		assert.Equal(t, "id", result.Errors[0].Field)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mod := &module.Module{
			Variables: []*module.Variable{
				{
					Type: "string",
				},
			},
			Resources: []*module.Resource{
				{
					Description: "Test resource",
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

		assert.Contains(t, fields, "id")
		assert.Contains(t, fields, "name")
		assert.Contains(t, fields, "version")
		assert.Contains(t, fields, "variables[0].name")
		assert.Contains(t, fields, "resources[0].type")
		assert.Contains(t, fields, "resources[0].provider")
	})

	t.Run("invalid variable type", func(t *testing.T) {
		mod := &module.Module{
			ID:      "test-module",
			Name:    "Test Module",
			Version: "1.0.0",
			Variables: []*module.Variable{
				{
					Name:        "var1",
					Type:        "invalid-type",
					Description: "Test variable",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "variables[0].type")
	})

	t.Run("invalid validation pattern", func(t *testing.T) {
		mod := &module.Module{
			ID:      "test-module",
			Name:    "Test Module",
			Version: "1.0.0",
			Variables: []*module.Variable{
				{
					Name:        "var1",
					Type:        "string",
					Description: "Test variable",
					Validation: &module.Validation{
						Pattern: "[invalid regex",
					},
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "variables[0].validation.pattern")
	})
}
