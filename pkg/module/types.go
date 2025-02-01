package module

import (
	"time"
)

// Module represents a module in the registry
type Module struct {
	// ID is the unique identifier for the module
	ID string `json:"id"`
	// Name is the module name
	Name string `json:"name"`
	// Provider is the cloud provider (e.g., aws, azure, gcp)
	Provider string `json:"provider"`
	// Version is the semantic version of the module
	Version string `json:"version"`
	// Description is a brief description of the module
	Description string `json:"description"`
	// Author is the module author
	Author string `json:"author"`
	// License is the module license
	License string `json:"license"`
	// Source is the source code location
	Source string `json:"source"`
	// Variables are the input variables for the module
	Variables []*Variable `json:"variables"`
	// Outputs are the output values from the module
	Outputs []*Output `json:"outputs"`
	// Dependencies are the module dependencies
	Dependencies []*Dependency `json:"dependencies"`
	// Resources are the cloud resources managed by the module
	Resources []*Resource `json:"resources"`
	// Tags are searchable tags for the module
	Tags []string `json:"tags"`
	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updated_at"`
	// Metadata is additional module metadata
	Metadata map[string]interface{} `json:"metadata"`
	// Tests are the module test cases
	Tests []*Test `json:"tests"`
}

// Test represents a module test case
type Test struct {
	// Name is the test case name
	Name string `json:"name"`
	// Description describes what the test verifies
	Description string `json:"description"`
	// Variables contains test-specific input variables
	Variables map[string]interface{} `json:"variables"`
	// ExpectedOutputs contains expected output values
	ExpectedOutputs map[string]interface{} `json:"expected_outputs"`
	// Setup contains setup steps to run before the test
	Setup []string `json:"setup"`
	// Teardown contains cleanup steps to run after the test
	Teardown []string `json:"teardown"`
	// Assertions contains test assertions
	Assertions []string `json:"assertions"`
	// Timeout is the maximum test duration
	Timeout time.Duration `json:"timeout"`
	// Skip indicates if the test should be skipped
	Skip bool `json:"skip"`
	// SkipReason explains why the test is skipped
	SkipReason string `json:"skip_reason,omitempty"`
}

// Variable represents a module input variable
type Variable struct {
	// Name is the variable name
	Name string `json:"name"`
	// Type is the variable type (e.g., string, number, list)
	Type string `json:"type"`
	// Description is a brief description of the variable
	Description string `json:"description"`
	// Default is the default value for the variable
	Default interface{} `json:"default,omitempty"`
	// Required indicates if the variable is required
	Required bool `json:"required"`
	// Sensitive indicates if the variable contains sensitive data
	Sensitive bool `json:"sensitive"`
	// Validation is the validation rules for the variable
	Validation *Validation `json:"validation,omitempty"`
}

// Output represents a module output value
type Output struct {
	// Name is the output name
	Name string `json:"name"`
	// Type is the output type
	Type string `json:"type"`
	// Description is a brief description of the output
	Description string `json:"description"`
	// Sensitive indicates if the output contains sensitive data
	Sensitive bool `json:"sensitive"`
}

// Dependency represents a module dependency
type Dependency struct {
	// Name is the dependency name
	Name string `json:"name"`
	// Source is the dependency source
	Source string `json:"source"`
	// Version is the version constraint
	Version string `json:"version"`
	// Required indicates if the dependency is required
	Required bool `json:"required"`
}

// Validation represents variable validation rules
type Validation struct {
	// Pattern is a regex pattern for string validation
	Pattern string `json:"pattern,omitempty"`
	// MinLength is the minimum length for strings
	MinLength *int `json:"min_length,omitempty"`
	// MaxLength is the maximum length for strings
	MaxLength *int `json:"max_length,omitempty"`
	// MinValue is the minimum value for numbers
	MinValue *float64 `json:"min_value,omitempty"`
	// MaxValue is the maximum value for numbers
	MaxValue *float64 `json:"max_value,omitempty"`
	// AllowedValues are the allowed values for the variable
	AllowedValues []interface{} `json:"allowed_values,omitempty"`
}

// Filter represents module search filters
type Filter struct {
	// Provider filters by cloud provider
	Provider string `json:"provider,omitempty"`
	// Tags filters by tags
	Tags []string `json:"tags,omitempty"`
	// Query is a search query string
	Query string `json:"query,omitempty"`
	// Limit is the maximum number of results
	Limit int `json:"limit,omitempty"`
	// Offset is the result offset for pagination
	Offset int `json:"offset,omitempty"`
}

// Resource represents a cloud resource
type Resource struct {
	// Type is the resource type
	Type string `json:"type"`
	// Provider is the cloud provider
	Provider string `json:"provider"`
	// Description describes the resource
	Description string `json:"description"`
	// Properties are resource properties
	Properties map[string]*Property `json:"properties"`
}

// Property represents a resource property
type Property struct {
	// Type is the property type
	Type string `json:"type"`
	// Description describes the property
	Description string `json:"description"`
	// Required indicates if the property is required
	Required bool `json:"required"`
}
