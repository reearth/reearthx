package domain

// ValidationResult represents the result of a domain validation
type ValidationResult struct {
	IsValid bool
	Errors  []error
}

// NewValidationResult creates a new validation result
func NewValidationResult(isValid bool, errors ...error) ValidationResult {
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
func Invalid(errors ...error) ValidationResult {
	return ValidationResult{
		IsValid: false,
		Errors:  errors,
	}
}
