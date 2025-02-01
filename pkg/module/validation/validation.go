package validation

import (
	"context"

	"github.com/StackCatalyst/common-lib/pkg/module"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Validator defines the interface for module validation
type Validator interface {
	// Validate performs validation on a module
	Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error)
}

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
	schemaValidator     SchemaValidator
	dependencyValidator DependencyValidator
	resourceValidator   ResourceValidator
}

// NewValidator creates a new DefaultValidator instance
func NewValidator() *DefaultValidator {
	return &DefaultValidator{
		schemaValidator:     NewSchemaValidator(),
		dependencyValidator: NewDependencyValidator(),
		resourceValidator:   NewResourceValidator(),
	}
}

// Validate performs all validation checks on a module
func (v *DefaultValidator) Validate(ctx context.Context, mod *module.Module) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Perform schema validation
	schemaResult, err := v.schemaValidator.Validate(ctx, mod)
	if err != nil {
		return nil, err
	}
	if !schemaResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, schemaResult.Errors...)
	}

	// Perform dependency validation
	depResult, err := v.dependencyValidator.Validate(ctx, mod)
	if err != nil {
		return nil, err
	}
	if !depResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, depResult.Errors...)
	}

	// Perform resource validation
	resResult, err := v.resourceValidator.Validate(ctx, mod)
	if err != nil {
		return nil, err
	}
	if !resResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, resResult.Errors...)
	}

	return result, nil
}
