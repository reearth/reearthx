package schema

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewMarkdown(t *testing.T) {
	assert.Equal(
		t,
		&FieldMarkdown{s: &FieldString{t: value.TypeMarkdown, maxLength: lo.ToPtr(1)}},
		NewMarkdown(lo.ToPtr(1)),
	)
}

func TestFieldMarkdown_Type(t *testing.T) {
	assert.Equal(
		t,
		value.TypeMarkdown,
		(&FieldMarkdown{s: &FieldString{t: value.TypeMarkdown}}).Type(),
	)
}

func TestFieldMarkdown_TypeProperty(t *testing.T) {
	f := FieldMarkdown{}
	assert.Equal(t, &TypeProperty{
		t:        f.Type(),
		markdown: &f,
	}, (&f).TypeProperty())
}

func TestFieldMarkdown_Clone(t *testing.T) {
	assert.Nil(t, (*FieldMarkdown)(nil).Clone())
	assert.Equal(t, &FieldMarkdown{}, (&FieldMarkdown{}).Clone())
}

func TestFieldMarkdown_Validate(t *testing.T) {
	assert.NoError(
		t,
		(&FieldMarkdown{s: &FieldString{t: value.TypeMarkdown}}).Validate(
			value.TypeMarkdown.Value("aaa"),
		),
	)
	assert.Equal(
		t,
		ErrInvalidValue,
		(&FieldMarkdown{s: &FieldString{t: value.TypeMarkdown}}).Validate(value.TypeText.Value("")),
	)
}
