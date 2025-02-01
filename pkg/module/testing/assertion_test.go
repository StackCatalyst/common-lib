package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertionEvaluation(t *testing.T) {
	ctx := &AssertionContext{
		Variables: map[string]interface{}{
			"name":    "test",
			"count":   42,
			"enabled": true,
		},
		Outputs: map[string]interface{}{
			"result": "success",
			"items":  []interface{}{"a", "b", "c"},
		},
		Resources: []*Resource{
			{
				ID:       "resource1",
				Type:     "test",
				Provider: "mock",
				Region:   "us-west-1",
				Properties: map[string]interface{}{
					"status": "active",
					"config": map[string]interface{}{
						"port": 8080,
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		assertion string
		want      bool
	}{
		{
			name:      "variable equals string",
			assertion: "variable name equals test",
			want:      true,
		},
		{
			name:      "variable equals number",
			assertion: "variable count equals 42",
			want:      true,
		},
		{
			name:      "variable equals boolean",
			assertion: "variable enabled equals true",
			want:      true,
		},
		{
			name:      "output equals",
			assertion: "output result equals success",
			want:      true,
		},
		{
			name:      "output contains",
			assertion: "output items contains b",
			want:      true,
		},
		{
			name:      "resource property equals",
			assertion: "resource resource1 status equals active",
			want:      true,
		},
		{
			name:      "resource nested property equals",
			assertion: "resource resource1 config.port equals 8080",
			want:      true,
		},
		{
			name:      "variable matches pattern",
			assertion: "variable name matches ^test$",
			want:      true,
		},
		{
			name:      "variable exists",
			assertion: "variable name exists",
			want:      true,
		},
		{
			name:      "variable type check",
			assertion: "variable count type int",
			want:      true,
		},
		{
			name:      "invalid assertion format",
			assertion: "invalid",
			want:      false,
		},
		{
			name:      "unknown reference type",
			assertion: "unknown name equals test",
			want:      false,
		},
		{
			name:      "unknown condition",
			assertion: "variable name unknown test",
			want:      false,
		},
		{
			name:      "non-existent variable",
			assertion: "variable unknown equals test",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateAssertion(tt.assertion, ctx)
			assert.Equal(t, tt.want, result.Success)
			assert.NotEmpty(t, result.Message)
		})
	}
}

func TestValueComparison(t *testing.T) {
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
		want     bool
	}{
		{
			name:     "equal strings",
			actual:   "test",
			expected: "test",
			want:     true,
		},
		{
			name:     "equal numbers",
			actual:   42,
			expected: "42",
			want:     true,
		},
		{
			name:     "equal booleans",
			actual:   true,
			expected: "true",
			want:     true,
		},
		{
			name:     "equal slices",
			actual:   []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
			want:     true,
		},
		{
			name: "equal maps",
			actual: map[string]interface{}{
				"key": "value",
			},
			expected: map[string]interface{}{
				"key": "value",
			},
			want: true,
		},
		{
			name:     "different types",
			actual:   42,
			expected: "string",
			want:     false,
		},
		{
			name:     "nil values",
			actual:   nil,
			expected: nil,
			want:     true,
		},
		{
			name:     "one nil value",
			actual:   nil,
			expected: "test",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equal, message := CompareValues(tt.actual, tt.expected)
			assert.Equal(t, tt.want, equal)
			if !equal {
				assert.NotEmpty(t, message)
			}
		})
	}
}

func TestResourcePropertyAccess(t *testing.T) {
	resources := []*Resource{
		{
			ID:       "test",
			Type:     "test",
			Provider: "mock",
			Properties: map[string]interface{}{
				"simple": "value",
				"nested": map[string]interface{}{
					"key": "value",
					"deep": map[string]interface{}{
						"number": 42,
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		id       string
		property string
		want     interface{}
	}{
		{
			name:     "simple property",
			id:       "test",
			property: "simple",
			want:     "value",
		},
		{
			name:     "nested property",
			id:       "test",
			property: "nested.key",
			want:     "value",
		},
		{
			name:     "deep nested property",
			id:       "test",
			property: "nested.deep.number",
			want:     42,
		},
		{
			name:     "non-existent resource",
			id:       "unknown",
			property: "simple",
			want:     nil,
		},
		{
			name:     "non-existent property",
			id:       "test",
			property: "unknown",
			want:     nil,
		},
		{
			name:     "non-existent nested property",
			id:       "test",
			property: "nested.unknown",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findResourceProperty(resources, tt.id, tt.property)
			assert.Equal(t, tt.want, got)
		})
	}
}
