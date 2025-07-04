package schema

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/stretchr/testify/assert"
)

func TestNewGroup(t *testing.T) {
	g := id.NewGroupID()
	assert.Equal(t, &FieldGroup{group: g}, NewGroup(g))
}

func TestFieldGroup_Group(t *testing.T) {
	g := id.NewGroupID()
	f := NewGroup(g)
	assert.Equal(t, g, f.Group())
}

func TestFieldGroup_Type(t *testing.T) {
	assert.Equal(t, value.TypeGroup, (&FieldGroup{}).Type())
}

func TestFieldGroup_TypeProperty(t *testing.T) {
	f := FieldGroup{}
	assert.Equal(t, &TypeProperty{
		t:     f.Type(),
		group: &f,
	}, (&f).TypeProperty())
}

func TestFieldGroup_Clone(t *testing.T) {
	g := id.NewGroupID()

	assert.Nil(t, (*FieldGroup)(nil).Clone())
	assert.Equal(t, &FieldGroup{group: g}, (&FieldGroup{group: g}).Clone())
}

func TestFieldGroup_Validate(t *testing.T) {
	gid := id.NewItemGroupID()
	assert.NoError(t, (&FieldGroup{}).Validate(value.TypeGroup.Value(gid)))
	assert.Error(t, (&FieldGroup{}).Validate(value.TypeGroup.Value("")))
}
