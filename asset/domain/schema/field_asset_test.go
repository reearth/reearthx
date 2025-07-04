package schema

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"

	"github.com/stretchr/testify/assert"
)

func TestNewAsset(t *testing.T) {
	assert.Equal(t, &FieldAsset{}, NewAsset())
}

func TestFieldAsset_Type(t *testing.T) {
	assert.Equal(t, value.TypeAsset, (&FieldAsset{}).Type())
}

func TestFieldAsset_TypeProperty(t *testing.T) {
	f := FieldAsset{}
	assert.Equal(t, &TypeProperty{
		t:     f.Type(),
		asset: &f,
	}, (&f).TypeProperty())
}

func TestFieldAsset_Clone(t *testing.T) {
	assert.Nil(t, (*FieldAsset)(nil).Clone())
	assert.Equal(t, &FieldAsset{}, (&FieldAsset{}).Clone())
}

func TestFieldAsset_Validate(t *testing.T) {
	aid := id.NewAssetID()
	assert.NoError(t, (&FieldAsset{}).Validate(value.TypeAsset.Value(aid)))
	assert.Equal(t, ErrInvalidValue, (&FieldAsset{}).Validate(value.TypeText.Value("")))
}
