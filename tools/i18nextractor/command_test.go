package i18nextractor

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var gofile = `package reearth
import (i "github.com/reearth/reearthx/i18n")
var hello = i.T("hello")
var goodbye = i.T("good.bye")
var seeyou = &i.Message{ID:"seeyou", Other: "seeyou"}
`

func TestCommand(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "aaa/aaa.go", []byte(gofile), 0666)
	_ = afero.WriteFile(fs, "ja.yml", []byte("hello: こんにちは\n"), 0666)

	assert.NoError(t, (&Config{
		Lang:   []string{"en", "ja"},
		Input:  fs,
		Output: fs,
	}).execute())

	en, _ := afero.ReadFile(fs, "en.yml")
	assert.Equal(t, "good.bye: \"\"\nhello: \"\"\nseeyou: seeyou\n", string(en))

	ja, _ := afero.ReadFile(fs, "ja.yml")
	assert.Equal(t, "good.bye: \"\"\nhello: こんにちは\nseeyou: seeyou\n", string(ja))
}
