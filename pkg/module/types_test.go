package module

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	minLength := 1
	maxLength := 100
	minValue := 0.0
	maxValue := 100.0

	module := &Module{
		ID:          "test-module",
		Name:        "test",
		Provider:    "aws",
		Version:     "1.0.0",
		Description: "Test module",
		Source:      "github.com/test/module",
		Variables: []*Variable{
			{
				Name:        "test_var",
				Type:        "string",
				Description: "Test variable",
				Default:     "default",
				Required:    true,
				Sensitive:   false,
				Validation: &Validation{
					Pattern:       "^test.*",
					MinLength:     &minLength,
					MaxLength:     &maxLength,
					MinValue:      &minValue,
					MaxValue:      &maxValue,
					AllowedValues: []interface{}{"test1", "test2"},
				},
			},
		},
		Outputs: []*Output{
			{
				Name:        "test_output",
				Type:        "string",
				Description: "Test output",
				Sensitive:   false,
			},
		},
		Dependencies: []*Dependency{
			{
				Name:     "test_dep",
				Source:   "github.com/test/dep",
				Version:  ">=1.0.0",
				Required: true,
			},
		},
		Tags:      []string{"test", "example"},
		CreatedAt: now,
		UpdatedAt: now,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	// Test marshaling
	data, err := json.Marshal(module)
	require.NoError(t, err)

	// Test unmarshaling
	var decoded Module
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, module.ID, decoded.ID)
	assert.Equal(t, module.Name, decoded.Name)
	assert.Equal(t, module.Provider, decoded.Provider)
	assert.Equal(t, module.Version, decoded.Version)
	assert.Equal(t, module.Description, decoded.Description)
	assert.Equal(t, module.Source, decoded.Source)

	// Verify variables
	require.Len(t, decoded.Variables, 1)
	assert.Equal(t, module.Variables[0].Name, decoded.Variables[0].Name)
	assert.Equal(t, module.Variables[0].Type, decoded.Variables[0].Type)
	assert.Equal(t, module.Variables[0].Description, decoded.Variables[0].Description)
	assert.Equal(t, module.Variables[0].Default, decoded.Variables[0].Default)
	assert.Equal(t, module.Variables[0].Required, decoded.Variables[0].Required)
	assert.Equal(t, module.Variables[0].Sensitive, decoded.Variables[0].Sensitive)

	// Verify validation
	require.NotNil(t, decoded.Variables[0].Validation)
	assert.Equal(t, module.Variables[0].Validation.Pattern, decoded.Variables[0].Validation.Pattern)
	assert.Equal(t, *module.Variables[0].Validation.MinLength, *decoded.Variables[0].Validation.MinLength)
	assert.Equal(t, *module.Variables[0].Validation.MaxLength, *decoded.Variables[0].Validation.MaxLength)
	assert.Equal(t, *module.Variables[0].Validation.MinValue, *decoded.Variables[0].Validation.MinValue)
	assert.Equal(t, *module.Variables[0].Validation.MaxValue, *decoded.Variables[0].Validation.MaxValue)
	assert.Equal(t, module.Variables[0].Validation.AllowedValues, decoded.Variables[0].Validation.AllowedValues)

	// Verify outputs
	require.Len(t, decoded.Outputs, 1)
	assert.Equal(t, module.Outputs[0].Name, decoded.Outputs[0].Name)
	assert.Equal(t, module.Outputs[0].Type, decoded.Outputs[0].Type)
	assert.Equal(t, module.Outputs[0].Description, decoded.Outputs[0].Description)
	assert.Equal(t, module.Outputs[0].Sensitive, decoded.Outputs[0].Sensitive)

	// Verify dependencies
	require.Len(t, decoded.Dependencies, 1)
	assert.Equal(t, module.Dependencies[0].Name, decoded.Dependencies[0].Name)
	assert.Equal(t, module.Dependencies[0].Source, decoded.Dependencies[0].Source)
	assert.Equal(t, module.Dependencies[0].Version, decoded.Dependencies[0].Version)
	assert.Equal(t, module.Dependencies[0].Required, decoded.Dependencies[0].Required)

	// Verify other fields
	assert.Equal(t, module.Tags, decoded.Tags)
	assert.Equal(t, module.CreatedAt.UTC(), decoded.CreatedAt.UTC())
	assert.Equal(t, module.UpdatedAt.UTC(), decoded.UpdatedAt.UTC())
	assert.Equal(t, module.Metadata, decoded.Metadata)
}

func TestFilterJSON(t *testing.T) {
	filter := &Filter{
		Provider: "aws",
		Tags:     []string{"test", "example"},
		Query:    "test module",
		Limit:    10,
		Offset:   0,
	}

	// Test marshaling
	data, err := json.Marshal(filter)
	require.NoError(t, err)

	// Test unmarshaling
	var decoded Filter
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, filter.Provider, decoded.Provider)
	assert.Equal(t, filter.Tags, decoded.Tags)
	assert.Equal(t, filter.Query, decoded.Query)
	assert.Equal(t, filter.Limit, decoded.Limit)
	assert.Equal(t, filter.Offset, decoded.Offset)
}
