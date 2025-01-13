package validation

import (
	"context"
	"fmt"
)

// ValidationError represents a validation error
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	IsValid bool
	Errors  []*Error
}

// NewValidationResult creates a new validation result
func NewValidationResult(isValid bool, errors ...*Error) ValidationResult {
	return ValidationResult{
		IsValid: isValid,
		Errors:  errors,
	}
}

// Valid creates a valid validation result
func Valid() ValidationResult {
	return ValidationResult{IsValid: true}
}

// Invalid creates an invalid validation result with errors
func Invalid(errors ...*Error) ValidationResult {
	return ValidationResult{
		IsValid: false,
		Errors:  errors,
	}
}

// ValidationRule defines a single validation rule
type ValidationRule interface {
	// Validate performs the validation and returns any errors
	Validate(ctx context.Context, value interface{}) error
}

// Validator defines the interface for entities that can be validated
type Validator interface {
	// Validate performs all validation rules and returns the result
	Validate(ctx context.Context) ValidationResult
}

// ValidationContext holds the context for validation
type ValidationContext struct {
	Rules []ValidationRule
}

// NewValidationContext creates a new validation context
func NewValidationContext(rules ...ValidationRule) *ValidationContext {
	return &ValidationContext{
		Rules: rules,
	}
}

// Validate executes all validation rules in the context
func (c *ValidationContext) Validate(ctx context.Context, value interface{}) ValidationResult {
	var errors []*Error

	// If value is a map, validate each field with its corresponding rules
	if fields, ok := value.(map[string]interface{}); ok {
		for _, rule := range c.Rules {
			if r, ok := rule.(*RequiredRule); ok {
				if fieldValue, exists := fields[r.Field]; exists {
					if err := rule.Validate(ctx, fieldValue); err != nil {
						errors = append(errors, err.(*Error))
					}
				} else {
					errors = append(errors, NewValidationError(r.Field, "field is required"))
				}
			} else if r, ok := rule.(*MaxLengthRule); ok {
				if fieldValue, exists := fields[r.Field]; exists {
					if err := rule.Validate(ctx, fieldValue); err != nil {
						errors = append(errors, err.(*Error))
					}
				}
			}
		}
	} else {
		// If value is not a map, validate directly
		for _, rule := range c.Rules {
			if err := rule.Validate(ctx, value); err != nil {
				if verr, ok := err.(*Error); ok {
					errors = append(errors, verr)
				} else {
					errors = append(errors, &Error{
						Message: err.Error(),
					})
				}
			}
		}
	}

	if len(errors) > 0 {
		return Invalid(errors...)
	}
	return Valid()
}

// ValidationError creates a new validation error
func NewValidationError(field, message string) *Error {
	return &Error{
		Field:   field,
		Message: message,
	}
}

// Common validation rules

// RequiredRule validates that a value is not empty
type RequiredRule struct {
	Field string
}

func (r *RequiredRule) Validate(ctx context.Context, value interface{}) error {
	if value == nil {
		return NewValidationError(r.Field, "field is required")
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return NewValidationError(r.Field, "field is required")
		}
	}

	return nil
}

// MaxLengthRule validates that a string value does not exceed a maximum length
type MaxLengthRule struct {
	Field     string
	MaxLength int
}

func (r *MaxLengthRule) Validate(ctx context.Context, value interface{}) error {
	if str, ok := value.(string); ok {
		if len(str) > r.MaxLength {
			return NewValidationError(r.Field, fmt.Sprintf("length must not exceed %d characters", r.MaxLength))
		}
	}
	return nil
}
