package validation

import (
	"context"
	"testing"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceValidator(t *testing.T) {
	validator := NewResourceValidator()
	ctx := context.Background()

	t.Run("valid resources", func(t *testing.T) {
		mod := &module.Module{
			Resources: []*module.Resource{
				{
					Type:        "test_resource",
					Provider:    "test_provider",
					Description: "Test resource",
					Properties: map[string]*module.Property{
						"prop1": {
							Type:        "string",
							Description: "Test property",
							Required:    true,
						},
						"prop2": {
							Type:        "number",
							Description: "Optional property",
							Required:    false,
						},
					},
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mod := &module.Module{
			Resources: []*module.Resource{
				{
					Description: "Test resource",
					Properties: map[string]*module.Property{
						"prop1": {
							Required: true,
						},
					},
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

		assert.Contains(t, fields, "resources[0].type")
		assert.Contains(t, fields, "resources[0].provider")
		assert.Contains(t, fields, "resources[0].properties.prop1.type")
		assert.Contains(t, fields, "resources[0].properties.prop1.description")
	})

	t.Run("invalid resource type format", func(t *testing.T) {
		mod := &module.Module{
			Resources: []*module.Resource{
				{
					Type:        "123invalid",
					Provider:    "test_provider",
					Description: "Test resource",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "resources[0].type")
		assert.Contains(t, result.Errors[0].Message, "must start with a letter")
	})

	t.Run("invalid provider format", func(t *testing.T) {
		mod := &module.Module{
			Resources: []*module.Resource{
				{
					Type:        "test_resource",
					Provider:    "invalid@provider",
					Description: "Test resource",
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "resources[0].provider")
		assert.Contains(t, result.Errors[0].Message, "must start with a letter")
	})

	t.Run("invalid property type", func(t *testing.T) {
		mod := &module.Module{
			Resources: []*module.Resource{
				{
					Type:        "test_resource",
					Provider:    "test_provider",
					Description: "Test resource",
					Properties: map[string]*module.Property{
						"prop1": {
							Type:        "invalid_type",
							Description: "Test property",
							Required:    true,
						},
					},
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "resources[0].properties.prop1.type")
		assert.Contains(t, result.Errors[0].Message, "invalid property type")
	})

	t.Run("missing description for required property", func(t *testing.T) {
		mod := &module.Module{
			Resources: []*module.Resource{
				{
					Type:        "test_resource",
					Provider:    "test_provider",
					Description: "Test resource",
					Properties: map[string]*module.Property{
						"prop1": {
							Type:     "string",
							Required: true,
						},
					},
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Field, "resources[0].properties.prop1.description")
		assert.Contains(t, result.Errors[0].Message, "description is required for required properties")
	})
}
