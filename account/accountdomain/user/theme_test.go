package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThemeFrom(t *testing.T) {
	assert.Equal(t, ThemeDark, ThemeFrom("dark"))
	assert.Equal(t, ThemeLight, ThemeFrom("light"))
	assert.Equal(t, ThemeDark, ThemeFrom("DARK"))
	assert.Equal(t, ThemeLight, ThemeFrom("LIGHT"))
	assert.Equal(t, ThemeDefault, ThemeFrom(""))
	assert.Equal(t, ThemeDefault, ThemeFrom("a"))
}
