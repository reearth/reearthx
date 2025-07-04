package schema

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/stretchr/testify/assert"
)

func TestNewURL(t *testing.T) {
	assert.Equal(t, &FieldURL{}, NewURL())
}

func TestFieldURL_Type(t *testing.T) {
	assert.Equal(t, value.TypeURL, (&FieldURL{}).Type())
}

func TestFieldURL_TypeProperty(t *testing.T) {
	f := FieldURL{}
	assert.Equal(t, &TypeProperty{
		t:   f.Type(),
		url: &f,
	}, (&f).TypeProperty())
}

func TestFieldURL_Clone(t *testing.T) {
	assert.Nil(t, (*FieldURL)(nil).Clone())
	assert.Equal(t, &FieldURL{}, (&FieldURL{}).Clone())
}

func TestFieldURL_Validate(t *testing.T) {
	assert.NoError(t, (&FieldURL{}).Validate(value.TypeURL.Value("https://example.com")))
	assert.Equal(t, ErrInvalidValue, (&FieldURL{}).Validate(value.TypeText.Value("")))
}
