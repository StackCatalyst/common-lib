package validation

import (
	"context"
	"fmt"
	"regexp"

	"github.com/StackCatalyst/common-lib/pkg/module"
)

// ResourceValidator defines the interface for resource validation
type ResourceValidator interface {
	Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error)
}

// DefaultResourceValidator implements ResourceValidator
type DefaultResourceValidator struct{}

// NewResourceValidator creates a new DefaultResourceValidator instance
func NewResourceValidator() ResourceValidator {
	return &DefaultResourceValidator{}
}

// Validate performs resource validation on a module
func (v *DefaultResourceValidator) Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Validate resources
	for i, res := range mod.Resources {
		// Validate required fields
		if res.Type == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("resources[%d].type", i),
				Message: "resource type is required",
			})
		}

		if res.Provider == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("resources[%d].provider", i),
				Message: "resource provider is required",
			})
		}

		// Validate properties
		if res.Properties != nil {
			for propName, prop := range res.Properties {
				if prop.Type == "" {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("resources[%d].properties.%s.type", i, propName),
						Message: "property type is required",
					})
				}

				// Validate property type
				validTypes := map[string]bool{
					"string": true,
					"number": true,
					"bool":   true,
					"list":   true,
					"map":    true,
					"object": true,
				}
				if !validTypes[prop.Type] {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("resources[%d].properties.%s.type", i, propName),
						Message: "invalid property type",
					})
				}

				// Validate required properties have descriptions
				if prop.Required && prop.Description == "" {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("resources[%d].properties.%s.description", i, propName),
						Message: "description is required for required properties",
					})
				}
			}
		}

		// Validate resource naming convention (alphanumeric with underscores)
		if matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", res.Type); !matched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("resources[%d].type", i),
				Message: "resource type must start with a letter and contain only alphanumeric characters and underscores",
			})
		}

		// Validate provider naming convention (alphanumeric with underscores)
		if matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", res.Provider); !matched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("resources[%d].provider", i),
				Message: "provider must start with a letter and contain only alphanumeric characters and underscores",
			})
		}
	}

	return result, nil
}
