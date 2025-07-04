package schema

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/stretchr/testify/assert"
)

func TestNewDateTime(t *testing.T) {
	assert.Equal(t, &FieldDateTime{}, NewDateTime())
}

func TestFieldDateTime_Type(t *testing.T) {
	assert.Equal(t, value.TypeDateTime, (&FieldDateTime{}).Type())
}

func TestFieldDateTime_TypeProperty(t *testing.T) {
	f := FieldDateTime{}
	assert.Equal(t, &TypeProperty{
		t:        f.Type(),
		dateTime: &f,
	}, (&f).TypeProperty())
}

func TestFieldDateTime_Clone(t *testing.T) {
	assert.Nil(t, (*FieldDateTime)(nil).Clone())
	assert.Equal(t, &FieldDateTime{}, (&FieldDateTime{}).Clone())
}

func TestFieldDateTime_Validate(t *testing.T) {
	now := time.Now()
	assert.NoError(t, (&FieldDateTime{}).Validate(value.TypeDateTime.Value(now)))
	assert.Equal(t, ErrInvalidValue, (&FieldDateTime{}).Validate(value.TypeText.Value("")))
}
