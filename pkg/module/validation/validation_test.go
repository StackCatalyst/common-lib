package validation

import (
	"context"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultValidator(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	t.Run("valid module", func(t *testing.T) {
		mod := &module.Module{
			ID:          "test-module",
			Name:        "Test Module",
			Version:     "1.0.0",
			Description: "Test module description",
			Author:      "Test Author",
			License:     "MIT",
			Dependencies: []*module.Dependency{
				{
					Name:     "dep1",
					Source:   "github.com/test/dep1",
					Version:  ">=1.0.0",
					Required: true,
				},
			},
			Variables: []*module.Variable{
				{
					Name:        "var1",
					Type:        "string",
					Description: "Test variable",
					Required:    true,
					Default:     "default",
				},
			},
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
					},
				},
			},
			Tests: []*module.Test{
				{
					Name:        "test1",
					Description: "Test case 1",
					Variables: map[string]interface{}{
						"var1": "value1",
					},
					ExpectedOutputs: map[string]interface{}{
						"output1": "expected1",
					},
					Assertions: []string{
						"variable var1 equals value1",
					},
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid module", func(t *testing.T) {
		mod := &module.Module{
			// Missing required fields
			Dependencies: []*module.Dependency{
				{
					// Missing required fields
					Version: "invalid-version",
				},
			},
			Variables: []*module.Variable{
				{
					// Missing required fields
					Type: "invalid-type",
				},
			},
			Resources: []*module.Resource{
				{
					// Missing required fields
					Properties: map[string]*module.Property{
						"prop1": {
							// Missing required fields
							Required: true,
						},
					},
				},
			},
			Tests: []*module.Test{
				{
					// Missing required fields
					Skip: true,
				},
			},
		}

		result, err := validator.Validate(ctx, mod)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)

		// Verify specific errors
		var fields []string
		for _, err := range result.Errors {
			fields = append(fields, err.Field)
		}

		// Check for expected error fields
		assert.Contains(t, fields, "id")
		assert.Contains(t, fields, "name")
		assert.Contains(t, fields, "version")
		assert.Contains(t, fields, "dependencies[0].name")
		assert.Contains(t, fields, "dependencies[0].source")
		assert.Contains(t, fields, "dependencies[0].version")
		assert.Contains(t, fields, "variables[0].name")
		assert.Contains(t, fields, "variables[0].type")
		assert.Contains(t, fields, "resources[0].type")
		assert.Contains(t, fields, "resources[0].provider")
		assert.Contains(t, fields, "resources[0].properties.prop1.type")
		assert.Contains(t, fields, "resources[0].properties.prop1.description")
		assert.Contains(t, fields, "tests[0].name")
		assert.Contains(t, fields, "tests[0].skip_reason")
	})
}
