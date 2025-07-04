package schema

import "github.com/reearth/reearthx/asset/domain/value"

type FieldURL struct{}

func NewURL() *FieldURL {
	return &FieldURL{}
}

func (f *FieldURL) TypeProperty() *TypeProperty {
	return &TypeProperty{
		t:   f.Type(),
		url: f,
	}
}

func (*FieldURL) Type() value.Type {
	return value.TypeURL
}

func (f *FieldURL) Clone() *FieldURL {
	if f == nil {
		return nil
	}
	return &FieldURL{}
}

func (f *FieldURL) Validate(v *value.Value) (err error) {
	v.Match(value.Match{
		URL: func(a value.URL) {
			// ok
		},
		Default: func() {
			err = ErrInvalidValue
		},
	})
	return
}

func (f *FieldURL) ValidateMultiple(v *value.Multiple) (err error) {
	return nil
}
