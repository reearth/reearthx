package schema

import (
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/model"
	"github.com/reearth/reearthx/asset/domain/value"
)

type CorrespondingField struct {
	Title       string
	Key         string
	Description string
	Required    bool
}

type FieldReference struct {
	correspondingFieldID *id.FieldID
	correspondingField   *CorrespondingField // from user input only
	modelID              id.ModelID
	schemaID             id.SchemaID
}

func NewReference(
	mID id.ModelID,
	sID id.SchemaID,
	cfID *id.FieldID,
	cf *CorrespondingField,
) *FieldReference {
	return &FieldReference{
		modelID:              mID,
		schemaID:             sID,
		correspondingFieldID: cfID,
		correspondingField:   cf,
	}
}

func (f *FieldReference) TypeProperty() *TypeProperty {
	return &TypeProperty{
		t:         f.Type(),
		reference: f,
	}
}

func (f *FieldReference) Model() model.ID {
	return f.modelID
}

func (f *FieldReference) Schema() id.SchemaID {
	return f.schemaID
}

// CorrespondingField returns the corresponding field of this reference from user input.
// This is not stored in the database.
func (f *FieldReference) CorrespondingField() *CorrespondingField {
	return f.correspondingField
}

func (f *FieldReference) CorrespondingFieldID() *id.FieldID {
	return f.correspondingFieldID
}

func (f *FieldReference) IsTowWay() bool {
	return f.correspondingFieldID != nil
}

func (f *FieldReference) Type() value.Type {
	return value.TypeReference
}

func (f *FieldReference) Clone() *FieldReference {
	if f == nil {
		return nil
	}
	return &FieldReference{
		modelID:              f.modelID,
		schemaID:             f.schemaID,
		correspondingFieldID: f.correspondingFieldID,
		correspondingField:   f.correspondingField,
	}
}

func (f *FieldReference) Validate(v *value.Value) (err error) {
	v.Match(value.Match{
		Reference: func(a value.Reference) {
			_, ok := v.ValueReference()
			if !ok {
				err = ErrInvalidValue
			}
		},
		Default: func() {
			err = ErrInvalidValue
		},
	})
	return
}

func (f *FieldReference) ValidateMultiple(_ *value.Multiple) error {
	return nil
}
