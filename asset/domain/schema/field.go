package schema

import (
	"fmt"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

var ErrValueRequired = rerror.NewE(i18n.T("value is required"))

type Field struct {
	updatedAt    time.Time
	defaultValue *value.Multiple
	typeProperty *TypeProperty
	name         string
	description  string
	key          id.Key
	order        int
	id           FieldID
	unique       bool
	multiple     bool
	required     bool
}

func (f *Field) ID() FieldID {
	return f.id
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) SetName(name string) {
	f.name = name
}

func (f *Field) Description() string {
	return f.description
}

func (f *Field) Order() int {
	return f.order
}

func (f *Field) SetDescription(description string) {
	f.description = description
}

func (f *Field) SetOrder(o int) {
	f.order = o
}

func (f *Field) DefaultValue() *value.Multiple {
	return f.defaultValue
}

func (f *Field) SetDefaultValue(v *value.Multiple) error {
	if v == nil {
		f.defaultValue = nil
		return nil
	}

	if v.Type() != f.Type() {
		return ErrInvalidValue
	}
	if err := f.ValidateValue(v); err != nil {
		return err
	}
	f.defaultValue = v
	return nil
}

func (f *Field) Key() id.Key {
	return f.key
}

func (f *Field) SetKey(key id.Key) error {
	if !key.IsValid() {
		return &rerror.Error{
			Label: ErrInvalidKey,
			Err:   fmt.Errorf("%s", key.String()),
		}
	}
	f.key = key
	return nil
}

func (f *Field) Unique() bool {
	return f.unique
}

func (f *Field) Multiple() bool {
	return f.multiple
}

func (f *Field) Required() bool {
	return f.required
}

func (f *Field) SetRequired(req bool) {
	f.required = req
}

func (f *Field) SetUnique(unique bool) {
	f.unique = unique
}

func (f *Field) SetMultiple(m bool) {
	f.multiple = m
}

func (f *Field) CreatedAt() time.Time {
	return f.id.Timestamp()
}

func (f *Field) UpdatedAt() time.Time {
	if f.updatedAt.IsZero() {
		return f.id.Timestamp()
	}
	return f.updatedAt
}

func (f *Field) Type() value.Type {
	return f.typeProperty.Type()
}

func (f *Field) TypeProperty() *TypeProperty {
	return f.typeProperty
}

func (f *Field) SetTypeProperty(tp *TypeProperty) error {
	if tp == nil {
		return ErrInvalidType
	}
	if !f.defaultValue.IsEmpty() {
		for _, v := range f.defaultValue.Values() {
			if err := tp.Validate(v); err != nil {
				return err
			}
		}
	}
	f.typeProperty = tp
	return nil
}

func (f *Field) Clone() *Field {
	if f == nil {
		return nil
	}

	return &Field{
		id:           f.id,
		name:         f.name,
		description:  f.description,
		key:          f.key,
		order:        f.order,
		unique:       f.unique,
		multiple:     f.multiple,
		required:     f.required,
		updatedAt:    f.updatedAt,
		typeProperty: f.typeProperty.Clone(),
		defaultValue: f.defaultValue.Clone(),
	}
}

// Validate the Multiple value against the Field schema
// if its multiple it will return only the first error
func (f *Field) Validate(m *value.Multiple) error {
	if f.required && m.IsEmpty() {
		return ErrValueRequired
	}
	return f.ValidateValue(m)
}

func (f *Field) ValidateValue(m *value.Multiple) error {
	if m.IsEmpty() {
		return nil
	}
	if !f.multiple && m.Len() > 1 {
		return ErrInvalidValue
	}
	for _, v := range m.Values() {
		if err := f.typeProperty.Validate(v); err != nil {
			return err
		}
	}
	return f.typeProperty.ValidateMultiple(m)
}

func (f *Field) IsGeometryField() bool {
	return f.Type() == value.TypeGeometryObject || f.Type() == value.TypeGeometryEditor
}

func (f *Field) SupportsPointField() bool {
	var supported bool
	f.TypeProperty().Match(TypePropertyMatch{
		GeometryObject: func(f *FieldGeometryObject) {
			supported = f.SupportedTypes().Has(GeometryObjectSupportedTypePoint)
		},
		GeometryEditor: func(f *FieldGeometryEditor) {
			supported = f.SupportedTypes().Has(GeometryEditorSupportedTypePoint)
		},
	})
	return supported
}
