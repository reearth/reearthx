package i18n

import (
	"io"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestBundle_LoadMessageFileFS(t *testing.T) {
	fs := afero.NewMemMapFs()
	f, _ := fs.Create("en.yml")
	_, _ = io.WriteString(f, `test: test`)
	_ = f.Close()
	f, _ = fs.Create("ja.yml")
	_, _ = io.WriteString(f, `test: テスト`)
	_ = f.Close()

	b := NewBundle(language.English)
	b.MustLoadFS(afero.NewIOFS(fs), "en.yml", "ja.yml")
	assert.Equal(t, []language.Tag{language.English, language.Japanese}, b.LanguageTags())
	assert.Equal(t, "テスト", NewLocalizer(b, "ja").MustLocalize(&i18n.LocalizeConfig{
		MessageID: "test",
	}))
}
