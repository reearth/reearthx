package value

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/util"
	"github.com/stretchr/testify/assert"
)

func TestValue_Match(t *testing.T) {
	var res any
	(&Value{t: TypeText, v: "aaa"}).Match(Match{Text: func(v string) { res = v }})
	assert.Equal(t, "aaa", res)

	res = nil
	(&Value{t: TypeTextArea, v: "aaa"}).Match(Match{TextArea: func(v string) { res = v }})
	assert.Equal(t, "aaa", res)

	res = nil
	(&Value{t: TypeRichText, v: "aaa"}).Match(Match{RichText: func(v string) { res = v }})
	assert.Equal(t, "aaa", res)

	res = nil
	(&Value{t: TypeMarkdown, v: "#aaa"}).Match(Match{Markdown: func(v string) { res = v }})
	assert.Equal(t, "#aaa", res)

	res = nil
	now := util.Now()
	(&Value{t: TypeDateTime, v: now}).Match(Match{DateTime: func(v DateTime) { res = v }})
	assert.Equal(t, now, res)

	res = nil
	aid := id.NewAssetID()
	(&Value{t: TypeAsset, v: aid}).Match(Match{Asset: func(v Asset) { res = v }})
	assert.Equal(t, aid, res)

	res = nil
	(&Value{t: TypeNumber, v: 5.0}).Match(Match{Number: func(v Number) { res = v }})
	assert.Equal(t, 5.0, res)

	res = nil
	(&Value{t: TypeInteger, v: int64(5)}).Match(Match{Integer: func(v Integer) { res = v }})
	assert.Equal(t, int64(5), res)

	res = nil
	(&Value{t: TypeBool, v: true}).Match(Match{Text: func(v string) { res = v }})
	assert.Nil(t, res)

	res = nil
	(&Value{t: TypeBool}).Match(Match{Default: func() { res = "default" }})
	assert.Equal(t, "default", res)

	res = nil
	g := `{
				"type": "Point",
				"coordinates": [102.0, 0.5]
			}`
	(&Value{t: TypeGeometryObject, v: g}).Match(Match{GeometryObject: func(v string) { res = v }})
	assert.Equal(t, g, res)

	res = nil
	ge := `{
				"type": "Point",
				"coordinates": [102.0, 0.5]
			}`
	(&Value{t: TypeGeometryEditor, v: ge}).Match(Match{GeometryEditor: func(v string) { res = v }})
	assert.Equal(t, ge, res)
}

func TestOptional_Match(t *testing.T) {
	var res any
	(&Optional{t: TypeText, v: &Value{t: TypeText, v: "aaa"}}).Match(
		OptionalMatch{Match: Match{Text: func(v string) { res = v }}},
	)
	assert.Equal(t, "aaa", res)

	res = nil
	(&Optional{t: TypeBool}).Match(OptionalMatch{None: func() { res = "none" }})
	assert.Equal(t, "none", res)

	res = nil
	(&Optional{t: TypeBool}).Match(OptionalMatch{Match: Match{Default: func() { res = "default" }}})
	assert.Equal(t, "default", res)
}
