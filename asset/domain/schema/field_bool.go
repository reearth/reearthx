package schema

import "github.com/reearth/reearthx/asset/domain/value"

type FieldBool struct{}

func NewBool() *FieldBool {
	return &FieldBool{}
}

func (f *FieldBool) TypeProperty() *TypeProperty {
	return &TypeProperty{
		t:    f.Type(),
		bool: f,
	}
}

func (f *FieldBool) Type() value.Type {
	return value.TypeBool
}

func (f *FieldBool) Clone() *FieldBool {
	if f == nil {
		return nil
	}
	return &FieldBool{}
}

func (f *FieldBool) Validate(v *value.Value) (err error) {
	v.Match(value.Match{
		Bool: func(a value.Bool) {
			// ok
		},
		Default: func() {
			err = ErrInvalidValue
		},
	})
	return
}

func (f *FieldBool) ValidateMultiple(v *value.Multiple) error {
	return nil
}
