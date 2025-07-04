package schema

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/id"

	"github.com/stretchr/testify/assert"
)

func TestFieldReferenceFromTypeProperty(t *testing.T) {
	// check that it returns true and correct field reference if type is reference
	mid1 := id.NewModelID()
	sid1 := id.NewSchemaID()
	fid1 := id.NewFieldID()
	f1 := NewField(
		NewReference(mid1, sid1, nil, nil).TypeProperty(),
	).ID(fid1).
		Key(id.RandomKey()).
		MustBuild()
	got1, ok := FieldReferenceFromTypeProperty(f1.TypeProperty())
	want1 := &FieldReference{
		modelID:              mid1,
		schemaID:             sid1,
		correspondingFieldID: nil,
		correspondingField:   nil,
	}
	assert.True(t, ok)
	assert.Equal(t, want1, got1)

	// check that it returns false and nil if type is not reference
	fid2 := id.NewFieldID()
	f2 := NewField(NewText(nil).TypeProperty()).ID(fid2).Key(id.RandomKey()).MustBuild()
	got2, ok := FieldReferenceFromTypeProperty(f2.TypeProperty())
	assert.False(t, ok)
	assert.Nil(t, got2)
}
