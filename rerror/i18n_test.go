package rerror

import (
	"testing"

	"github.com/reearth/reearthx/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestE(t *testing.T) {
	b := i18n.NewBundle(language.Japanese)
	l := i18n.NewLocalizer(b, "ja")
	b.MustAddMessages(
		language.Japanese,
		&i18n.Message{ID: "hello", Other: "こんにちは"},
		&i18n.Message{ID: "hello: %s", Other: "こんにちは: %s"},
		&i18n.Message{ID: "hello: %w", Other: "こんにちは: %w"},
		&i18n.Message{ID: IDErrInternal, Other: "内部エラー"},
		&i18n.Message{ID: IDErrNotFound, Other: "見つかりませんでした"},
	)

	e1 := NewE(&i18n.Message{ID: "hello"})
	assert.Equal(t, "こんにちは", e1.LocalizeError(l).Error())
	assert.Equal(t, "hello", e1.Error())

	e2 := NewE(&i18n.Message{ID: "hello2"})
	assert.Equal(t, "hello2", e2.LocalizeError(l).Error())
	assert.Equal(t, "hello2", e2.Error())

	e3 := FmtE(&i18n.Message{ID: "hello: %s"}, "aaa")
	assert.Equal(t, "こんにちは: aaa", e3.LocalizeError(l).Error())
	assert.Equal(t, "hello: aaa", e3.Error())

	e4 := FmtE(&i18n.Message{ID: "hello: %w"}, e1)
	assert.Equal(t, "こんにちは: こんにちは", e4.LocalizeError(l).Error())
	assert.Equal(t, "hello: hello", e4.Error())
	assert.Same(t, e1, e4.Unwrap())

	e5 := FmtE(&i18n.Message{ID: "hello: %w"}, e4)
	assert.Equal(t, "こんにちは: こんにちは: こんにちは", e5.LocalizeError(l).Error())
	assert.Equal(t, "hello: hello: hello", e5.Error())
	assert.Same(t, e4, e5.Unwrap())

	e7 := ErrInternalBy(e1)
	assert.Equal(t, "内部エラー", Localize(l, e7).Error())
	assert.Equal(t, "internal", e7.Error())
	assert.True(t, IsInternal(e7))
	assert.Same(t, e1, UnwrapErrInternal(e7))

	e8 := ErrNotFound
	assert.Equal(t, "見つかりませんでした", Localize(l, e8).Error())
	assert.Equal(t, "not found", e8.Error())
}
