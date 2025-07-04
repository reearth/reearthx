package item

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/stretchr/testify/assert"
)

func TestNewField(t *testing.T) {
	fId := id.NewFieldID()
	assert.Nil(t, NewField(fId, nil, nil))
	f := NewField(fId, value.TypeBool.Value(true).AsMultiple(), nil)
	assert.Equal(t, &Field{
		field: fId,
		value: value.TypeBool.Value(true).AsMultiple(),
	}, f)

	assert.Equal(t, value.TypeBool, f.Type())
}
