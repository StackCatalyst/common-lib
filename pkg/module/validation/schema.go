package validation

import (
	"context"
	"fmt"
	"regexp"

	"github.com/StackCatalyst/common-lib/pkg/module"
)

// SchemaValidator defines the interface for schema validation
type SchemaValidator interface {
	Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error)
}

// DefaultSchemaValidator implements SchemaValidator
type DefaultSchemaValidator struct{}

// NewSchemaValidator creates a new DefaultSchemaValidator instance
func NewSchemaValidator() SchemaValidator {
	return &DefaultSchemaValidator{}
}

// Validate performs schema validation on a module
func (v *DefaultSchemaValidator) Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Validate required fields
	if mod.ID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "id",
			Message: "module ID is required",
		})
	}

	if mod.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "module name is required",
		})
	}

	if mod.Version == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "version",
			Message: "module version is required",
		})
	}

	// Validate ID format (alphanumeric with hyphens)
	if matched, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", mod.ID); !matched && mod.ID != "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "id",
			Message: "module ID must contain only alphanumeric characters and hyphens",
		})
	}

	// Validate variables
	for i, v := range mod.Variables {
		if v.Name == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("variables[%d].name", i),
				Message: "variable name is required",
			})
		}

		if v.Type == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("variables[%d].type", i),
				Message: "variable type is required",
			})
		}

		// Validate variable type
		validTypes := map[string]bool{
			"string": true,
			"number": true,
			"bool":   true,
			"list":   true,
			"map":    true,
			"object": true,
		}
		if !validTypes[v.Type] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("variables[%d].type", i),
				Message: "invalid variable type",
			})
		}

		// Validate validation rules if present
		if v.Validation != nil {
			if v.Validation.Pattern != "" {
				if _, err := regexp.Compile(v.Validation.Pattern); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("variables[%d].validation.pattern", i),
						Message: "invalid regex pattern",
					})
				}
			}
		}
	}

	// Validate resources
	for i, r := range mod.Resources {
		if r.Type == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("resources[%d].type", i),
				Message: "resource type is required",
			})
		}

		if r.Provider == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("resources[%d].provider", i),
				Message: "resource provider is required",
			})
		}
	}

	// Validate tests
	for i, t := range mod.Tests {
		if t.Name == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("tests[%d].name", i),
				Message: "test name is required",
			})
		}

		if t.Skip && t.SkipReason == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("tests[%d].skip_reason", i),
				Message: "skip reason is required when test is skipped",
			})
		}
	}

	return result, nil
}
