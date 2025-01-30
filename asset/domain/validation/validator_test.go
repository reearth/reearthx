package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError("name", "field is required")
	assert.Equal(t, "name: field is required", err.Error())
}

func TestValidationResult(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		result := Valid()
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Invalid", func(t *testing.T) {
		err := NewValidationError("name", "field is required")
		result := Invalid(err)
		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
	})
}

func TestRequiredRule(t *testing.T) {
	ctx := context.Background()
	rule := &RequiredRule{Field: "test"}

	t.Run("nil value", func(t *testing.T) {
		err := rule.Validate(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, "test: field is required", err.Error())
	})

	t.Run("empty string", func(t *testing.T) {
		err := rule.Validate(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, "test: field is required", err.Error())
	})

	t.Run("non-empty string", func(t *testing.T) {
		err := rule.Validate(ctx, "value")
		assert.NoError(t, err)
	})
}

func TestMaxLengthRule(t *testing.T) {
	ctx := context.Background()
	rule := &MaxLengthRule{Field: "test", MaxLength: 5}

	t.Run("string within limit", func(t *testing.T) {
		err := rule.Validate(ctx, "12345")
		assert.NoError(t, err)
	})

	t.Run("string exceeding limit", func(t *testing.T) {
		err := rule.Validate(ctx, "123456")
		assert.Error(t, err)
		assert.Equal(t, "test: length must not exceed 5 characters", err.Error())
	})

	t.Run("non-string value", func(t *testing.T) {
		err := rule.Validate(ctx, 123)
		assert.NoError(t, err)
	})
}

func TestValidationContext(t *testing.T) {
	ctx := context.Background()

	t.Run("multiple rules passing", func(t *testing.T) {
		validationCtx := NewValidationContext(
			&RequiredRule{Field: "name"},
			&MaxLengthRule{Field: "name", MaxLength: 10},
		)

		result := validationCtx.Validate(ctx, map[string]interface{}{
			"name": "test",
		})

		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("multiple rules failing", func(t *testing.T) {
		validationCtx := NewValidationContext(
			&RequiredRule{Field: "name"},
			&MaxLengthRule{Field: "name", MaxLength: 5},
		)

		result := validationCtx.Validate(ctx, map[string]interface{}{
			"name": "too long name",
		})

		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, "name: length must not exceed 5 characters", result.Errors[0].Error())
	})
}
