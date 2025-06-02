package schema

import (
	"github.com/reearth/reearthx/asset/domain/value"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBool(t *testing.T) {
	assert.Equal(t, &FieldBool{}, NewBool())
}

func TestFieldBool_Type(t *testing.T) {
	assert.Equal(t, value.TypeBool, (&FieldBool{}).Type())
}

func TestFieldBool_TypeProperty(t *testing.T) {
	f := FieldBool{}
	assert.Equal(t, &TypeProperty{
		t:    f.Type(),
		bool: &f,
	}, (&f).TypeProperty())
}

func TestFieldBool_Clone(t *testing.T) {
	assert.Nil(t, (*FieldBool)(nil).Clone())
	assert.Equal(t, &FieldBool{}, (&FieldBool{}).Clone())
}

func TestFieldBool_Validate(t *testing.T) {
	assert.NoError(t, (&FieldBool{}).Validate(value.TypeBool.Value(true)))
	assert.Equal(t, ErrInvalidValue, (&FieldBool{}).Validate(value.TypeText.Value("")))
}
