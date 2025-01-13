package domain_test

import (
	"errors"
	"testing"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewValidationResult(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
		errors  []error
		want    domain.ValidationResult
	}{
		{
			name:    "valid result without errors",
			isValid: true,
			errors:  nil,
			want: domain.ValidationResult{
				IsValid: true,
				Errors:  nil,
			},
		},
		{
			name:    "invalid result with errors",
			isValid: false,
			errors:  []error{errors.New("test error")},
			want: domain.ValidationResult{
				IsValid: false,
				Errors:  []error{errors.New("test error")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.NewValidationResult(tt.isValid, tt.errors...)
			assert.Equal(t, tt.want.IsValid, got.IsValid)
			if tt.errors == nil {
				assert.Empty(t, got.Errors)
			} else {
				assert.Equal(t, tt.want.Errors[0].Error(), got.Errors[0].Error())
			}
		})
	}
}

func TestValid(t *testing.T) {
	result := domain.Valid()
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Errors)
}

func TestInvalid(t *testing.T) {
	err := errors.New("test error")
	result := domain.Invalid(err)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, err.Error(), result.Errors[0].Error())
}
