package mongox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestIndexFromKey(t *testing.T) {
	assert.Equal(t, Index{
		Name: "re_a,b.c",
		Key: bson.D{
			{Key: "a", Value: 1},
			{Key: "b.c", Value: 1},
		},
	}, IndexFromKey("a,b.c", false))

	assert.Equal(t, []Index{{
		Name:   "re_!a,b",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b", Value: 1},
		},
	}}, IndexFromKeys([]string{"!a,b"}, true))
}

func TestCaseInsensitiveIndexFromKey(t *testing.T) {
	assert.Equal(t, Index{
		Name: "re_a,b.c",
		Key: bson.D{
			{Key: "a", Value: 1},
			{Key: "b.c", Value: 1},
		},
		CaseInsensitive: true,
	}, CaseInsensitiveIndexFromKey("a,b.c", false))
}

func TestToKeyBSON(t *testing.T) {
	assert.Equal(
		t,
		bson.D{{Key: "a", Value: 1}},
		toKeyBSON("a"),
	)
	assert.Equal(
		t,
		bson.D{{Key: "b.a", Value: 1}, {Key: "b", Value: 1}},
		toKeyBSON("b.a,b"),
	)
}

func TestIndex_Model(t *testing.T) {
	assert.Equal(t, mongo.IndexModel{
		Keys: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Options: options.Index().SetName("aaa").SetUnique(true).SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2,
		}).SetPartialFilterExpression(bson.M{
			"a": "A",
			"b": "B",
		}),
	}, Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"a": "A",
			"b": "B",
		},
		CaseInsensitive: true,
	}.Model())
}

func TestIndex_Normalize(t *testing.T) {
	assert.Equal(t, Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: int32(-1)},
			{Key: "b.c", Value: int32(1)},
		},
		Filter: bson.M{
			"a": float64(1.1),
		},
	}, Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"a": 1.1,
		},
	}.Normalize())
}

func TestIndex_Equal(t *testing.T) {
	assert.True(t, Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"a": "A",
			"b": "B",
		},
	}.Equal(Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"b": "B",
			"a": "A",
		},
	}))

	assert.False(t, Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"a": "A",
			"b": "C",
		},
	}.Equal(Index{
		Name:   "aaa",
		Unique: true,
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"a": "A",
			"b": "B",
		},
	}))

	assert.False(t, Index{
		Name: "aaa",
		Key: bson.D{
			{Key: "b.c", Value: 1},
			{Key: "a", Value: -1},
		},
		Filter: bson.M{
			"a": "A",
			"b": "C",
		},
	}.Equal(Index{
		Name: "aaa",
		Key: bson.D{
			{Key: "a", Value: -1},
			{Key: "b.c", Value: 1},
		},
		Filter: bson.M{
			"a": "A",
			"b": "C",
		},
	}))
}

func TestIndexDocument_Names(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, IndexList{{Name: "a"}, {Name: "b"}}.Names())
}

func TestIndexDocument_NamesWithoutPrefix(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, IndexList{{Name: "a"}, {Name: "re_b"}}.NamesWithoutPrefix())
}

func TestIndexDocument_Models(t *testing.T) {
	assert.Equal(t, []mongo.IndexModel{
		{Keys: bson.D{{Key: "a", Value: 1}}, Options: options.Index().SetName("a")},
		{Keys: bson.D(nil), Options: options.Index().SetName("b")},
	}, IndexList{{Name: "a", Key: bson.D{{Key: "a", Value: 1}}}, {Name: "b"}}.Models())
}

func TestIndexDocument_AddNamePrefix(t *testing.T) {
	assert.Equal(t, IndexList{
		{Name: "re_b"}, {Name: "re_c"}, {Name: ""},
	}, IndexList{{Name: "re_b"}, {Name: "c"}, {Name: ""}}.AddNamePrefix())
}

func TestIndexDocument_RemoveDefaultIndex(t *testing.T) {
	assert.Equal(t, IndexList{
		{Name: "re_b"}, {Name: "c"},
	}, IndexList{{Name: "_id_"}, {Name: "re_b"}, {Name: "c"}}.RemoveDefaultIndex())
}
