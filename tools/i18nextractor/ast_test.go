package i18nextractor

import (
	"testing"

	"github.com/reearth/reearthx/i18n"
	"github.com/stretchr/testify/assert"
)

func TestExtractMessages(t *testing.T) {
	buf := []byte(`package reearth
import (i "github.com/reearth/reearthx/i18n")
var a = i.T("aaa")
var b = &i.Message{ID:"bbb", Other:"ccc"}
var c = &i.Message{ID:"ccc"}
func main() {
	nested := struct{Name *i18n.Message}{Name:i.T("ddd")}
}
`)

	msgs, err := extractMessages(buf)
	assert.NoError(t, err)
	assert.Equal(t, []*i18n.Message{
		{ID: "aaa"},
		{ID: "bbb", Other: "ccc"},
		{ID: "ccc"},
		{ID: "ddd"},
	}, msgs)
}
