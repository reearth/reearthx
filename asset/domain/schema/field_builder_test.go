package schema

import (
	"fmt"
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"

	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
)

func TestNewField(t *testing.T) {
	// ok
	now := time.Now()
	tp := NewText(nil).TypeProperty()
	dv := tp.Type().Value("aaa")
	fId := id.NewFieldID()
	k := id.RandomKey()
	assert.Equal(
		t,
		&Field{
			id:           fId,
			name:         "name",
			description:  "a",
			key:          k,
			unique:       true,
			multiple:     true,
			required:     true,
			typeProperty: tp,
			order:        3,
			updatedAt:    now,
			defaultValue: dv.AsMultiple(),
		},
		NewField(tp).
			ID(fId).
			Name("name").
			Description("a").
			Key(k).
			Multiple(true).
			Unique(true).
			Required(true).
			DefaultValue(dv.AsMultiple()).
			Order(3).
			UpdatedAt(now).
			Type(tp).
			MustBuild(),
	)

	f := NewField(tp).
		ID(fId).
		RandomKey().
		Type(tp).
		MustBuild()
	assert.NotNil(t, f.Key())

	// error: invalid id
	_, err := NewField(tp).Build()
	assert.Equal(t, ErrInvalidID, err)

	// error: invalid type
	_, err = NewField(nil).NewID().Build()
	assert.Equal(t, ErrInvalidType, err)

	// error: invalid key
	_, err = NewField(tp).NewID().Build()
	assert.Equal(t, &rerror.Error{
		Label: ErrInvalidKey,
		Err:   fmt.Errorf("%s", ""),
	}, err)

	// error: invalid default value
	_, err = NewField(NewText(nil).TypeProperty()).
		NewID().
		Key(k).
		DefaultValue(value.TypeBool.Value(true).AsMultiple()).
		Build()
	assert.Equal(t, ErrInvalidValue, err)

	assert.Panics(t, func() {
		_ = NewField(tp).MustBuild()
	})
}
