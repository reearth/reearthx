package schema

import (
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"
)

func CreateCorrespondingField(
	sid id.SchemaID,
	mid id.ModelID,
	f *Field,
	inp CorrespondingField,
) (*Field, error) {
	if f == nil || f.typeProperty == nil || f.typeProperty.reference == nil {
		return nil, ErrInvalidType
	}

	tp := NewReference(mid, sid, f.ID().Ref(), nil).TypeProperty()

	cf, err := NewField(tp).
		NewID().
		Unique(false).
		Multiple(false).
		Required(inp.Required).
		Name(inp.Title).
		Description(inp.Description).
		Key(id.NewKey(inp.Key)).
		DefaultValue(nil).
		Build()
	if err != nil {
		return nil, err
	}

	f.typeProperty.reference.correspondingFieldID = cf.ID().Ref()

	return cf, nil
}

func FieldReferenceFromTypeProperty(tp *TypeProperty) (*FieldReference, bool) {
	if tp == nil {
		return nil, false
	}
	return tp.reference, tp.Type() == value.TypeReference && tp.reference != nil
}
