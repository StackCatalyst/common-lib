package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/StackCatalyst/common-lib/pkg/module/version"
)

// DependencyValidator defines the interface for dependency validation
type DependencyValidator interface {
	Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error)
}

// DefaultDependencyValidator implements DependencyValidator
type DefaultDependencyValidator struct {
	versionManager version.Manager
}

// NewDependencyValidator creates a new DefaultDependencyValidator instance
func NewDependencyValidator() DependencyValidator {
	return &DefaultDependencyValidator{
		versionManager: version.NewManager(),
	}
}

// Validate performs dependency validation on a module
func (v *DefaultDependencyValidator) Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Validate module version format
	if !v.versionManager.IsValid(mod.Version) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("invalid module version format: %s", mod.Version),
		})
	}

	// Validate dependencies
	seen := make(map[string]bool)
	for i, dep := range mod.Dependencies {
		// Validate required fields
		if dep.Name == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("dependencies[%d].name", i),
				Message: "dependency name is required",
			})
		} else {
			// Check for circular dependencies
			if dep.Name == mod.Name {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("dependencies[%d].name", i),
					Message: "circular dependency detected: module cannot depend on itself",
				})
			}

			// Check for duplicate dependencies
			if seen[dep.Name] {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("dependencies[%d].name", i),
					Message: fmt.Sprintf("duplicate dependency: %s", dep.Name),
				})
			}
			seen[dep.Name] = true
		}

		if dep.Source == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("dependencies[%d].source", i),
				Message: "dependency source is required",
			})
		}

		// Validate version constraint format
		if dep.Version != "" {
			// Check if it's a constraint (starts with ~, ^, >=, >, <=, <, or =)
			isConstraint := strings.ContainsAny(dep.Version[0:1], "~^><")
			if !isConstraint && !strings.HasPrefix(dep.Version, "=") {
				// If it's not a constraint, it should be a valid version
				if !v.versionManager.IsValid(dep.Version) {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("dependencies[%d].version", i),
						Message: fmt.Sprintf("invalid version format: %s", dep.Version),
					})
				}
			} else {
				// Try to parse the constraint
				if _, err := v.versionManager.IsSatisfied("1.0.0", dep.Version); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("dependencies[%d].version", i),
						Message: fmt.Sprintf("invalid version constraint format: %v", err),
					})
				}
			}
		}
	}

	return result, nil
}
