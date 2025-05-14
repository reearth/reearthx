package assetusecase

import (
	"fmt"

	"github.com/reearth/reearthx/asset/domain/validation"
)

// ResultCode represents the status of an operation
type ResultCode string

const (
	ResultCodeSuccess ResultCode = "SUCCESS"
	ResultCodeError   ResultCode = "ERROR"
)

// Result represents the result of a use case operation
type Result struct {
	Code    ResultCode
	Data    interface{}
	Errors  []*Error
	Message string
}

// Error represents an error in the use case layer
type Error struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewResult creates a new success result with data
func NewResult(data interface{}) *Result {
	return &Result{
		Code: ResultCodeSuccess,
		Data: data,
	}
}

// NewErrorResult creates a new error result
func NewErrorResult(code string, message string, details map[string]interface{}) *Result {
	return &Result{
		Code: ResultCodeError,
		Errors: []*Error{
			{
				Code:    code,
				Message: message,
				Details: details,
			},
		},
	}
}

// NewValidationErrorResult creates a new validation error result
func NewValidationErrorResult(validationErrors []*validation.Error) *Result {
	errors := make([]*Error, len(validationErrors))
	for i, ve := range validationErrors {
		errors[i] = &Error{
			Code:    "VALIDATION_ERROR",
			Message: ve.Error(),
			Details: map[string]interface{}{
				"field": ve.Field,
			},
		}
	}

	return &Result{
		Code:   ResultCodeError,
		Errors: errors,
	}
}

// IsSuccess returns true if the result represents a successful operation
func (r *Result) IsSuccess() bool {
	return r.Code == ResultCodeSuccess
}

// GetError returns the first error if any
func (r *Result) GetError() error {
	if len(r.Errors) > 0 {
		return r.Errors[0]
	}
	return nil
}

// WithMessage adds a message to the result
func (r *Result) WithMessage(message string) *Result {
	r.Message = message
	return r
}

// Error implements the error interface for Result
func (r *Result) Error() string {
	if r == nil || len(r.Errors) == 0 {
		return ""
	}
	return r.Errors[0].Error()
}
