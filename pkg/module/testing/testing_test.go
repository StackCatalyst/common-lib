package testing

import (
	"context"
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	mod := &module.Module{
		ID:      "test-module",
		Name:    "Test Module",
		Version: "1.0.0",
		Tests: []*module.Test{
			{
				Name:        "test1",
				Description: "Test case 1",
				Variables: map[string]interface{}{
					"var1": "value1",
					"var2": 42,
				},
				ExpectedOutputs: map[string]interface{}{
					"output1": "expected1",
					"output2": true,
				},
				Setup: []string{
					"Create test database",
					"Initialize schema",
				},
				Teardown: []string{
					"Delete test data",
					"Drop database",
				},
				Assertions: []string{
					"variable var1 equals value1",
					"variable var2 equals 42",
					"output output1 equals expected1",
					"output output2 equals true",
				},
				Timeout: 5 * time.Second,
			},
			{
				Name:        "test2",
				Description: "Skipped test case",
				Skip:        true,
				SkipReason:  "Feature not implemented",
			},
		},
	}

	config := &Config{
		Provider: "mock",
		Region:   "us-west-1",
		Credentials: map[string]string{
			"access_key": "test",
			"secret_key": "test",
		},
		Variables: map[string]interface{}{
			"environment": "test",
		},
		Timeout:       30 * time.Second,
		Parallel:      false,
		KeepResources: false,
		Tags: map[string]string{
			"environment": "test",
			"purpose":     "testing",
		},
	}

	result, err := runner.Run(ctx, mod, config)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result
	assert.Equal(t, mod.ID, result.ModuleID)
	assert.Equal(t, mod.Version, result.Version)
	assert.NotZero(t, result.StartTime)
	assert.NotZero(t, result.EndTime)
	assert.NotZero(t, result.Duration)
	assert.Len(t, result.Tests, 2)

	// Verify first test case
	test1 := result.Tests[0]
	assert.Equal(t, "test1", test1.Name)
	assert.Equal(t, "Test case 1", test1.Description)
	assert.Equal(t, StatusPassed, test1.Status)
	assert.NotZero(t, test1.StartTime)
	assert.NotZero(t, test1.EndTime)
	assert.NotZero(t, test1.Duration)
	assert.Nil(t, test1.Error)
	assert.NotEmpty(t, test1.Logs)

	// Verify second test case (skipped)
	test2 := result.Tests[1]
	assert.Equal(t, "test2", test2.Name)
	assert.Equal(t, "Skipped test case", test2.Description)
	assert.Equal(t, StatusSkipped, test2.Status)
	assert.NotZero(t, test2.StartTime)
	assert.NotZero(t, test2.EndTime)
	assert.NotZero(t, test2.Duration)
	assert.Nil(t, test2.Error)
	assert.NotEmpty(t, test2.Logs)
	assert.Contains(t, test2.Logs[0], "Feature not implemented")

	// Test cleanup
	err = runner.Cleanup(ctx, result)
	assert.NoError(t, err)
}

func TestMockProvider(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	// Test creating a resource
	resource := &Resource{
		ID:       "test-resource",
		Type:     "test",
		Provider: "mock",
		Region:   "us-west-1",
		Tags: map[string]string{
			"environment": "test",
		},
		Properties: map[string]interface{}{
			"name": "test",
		},
	}

	err := provider.CreateResource(ctx, resource)
	require.NoError(t, err)
	assert.NotZero(t, resource.CreatedAt)

	// Test getting a resource
	got, err := provider.GetResource(ctx, resource.ID)
	require.NoError(t, err)
	assert.Equal(t, resource, got)

	// Test listing resources
	resources, err := provider.ListResources(ctx, map[string]string{
		"environment": "test",
	})
	require.NoError(t, err)
	assert.Len(t, resources, 1)
	assert.Equal(t, resource, resources[0])

	// Test deleting a resource
	err = provider.DeleteResource(ctx, resource)
	require.NoError(t, err)

	// Verify resource is deleted
	_, err = provider.GetResource(ctx, resource.ID)
	assert.Error(t, err)
}

func TestResourceValidation(t *testing.T) {
	provider := NewMockProvider()

	tests := []struct {
		name     string
		resource *Resource
		wantErr  bool
	}{
		{
			name: "valid resource",
			resource: &Resource{
				ID:       "test",
				Type:     "test",
				Provider: "mock",
				Region:   "us-west-1",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			resource: &Resource{
				ID:       "test",
				Provider: "mock",
				Region:   "us-west-1",
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			resource: &Resource{
				ID:     "test",
				Type:   "test",
				Region: "us-west-1",
			},
			wantErr: true,
		},
		{
			name: "missing region",
			resource: &Resource{
				ID:       "test",
				Type:     "test",
				Provider: "mock",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateResource(tt.resource)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
