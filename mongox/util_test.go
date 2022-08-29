package mongox

import (
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	mongotest.Env = "REEARTH_DB"
}

func TestConvertDToM(t *testing.T) {
	assert.Equal(t, bson.M{"a": "b"}, DToM(bson.M{"a": "b"}))
	assert.Equal(t, bson.M{"a": "b"}, DToM(bson.D{{Key: "a", Value: "b"}}))
	assert.Equal(t, []interface{}{bson.M{"a": "b"}}, DToM([]bson.D{{{Key: "a", Value: "b"}}}))
	assert.Equal(t, []interface{}{bson.M{"a": "b"}}, DToM([]bson.M{{"a": "b"}}))
	assert.Equal(t, []interface{}{bson.M{"a": "b"}}, DToM(bson.A{bson.D{{Key: "a", Value: "b"}}}))
	assert.Equal(t, []interface{}{bson.M{"a": "b"}}, DToM([]interface{}{bson.D{{Key: "a", Value: "b"}}}))
}

func TestAppendI(t *testing.T) {
	assert.Equal(t, []interface{}{bson.M{"a": "b"}, "x"}, AppendI([]bson.M{{"a": "b"}}, "x"))
	assert.Equal(t, []interface{}{bson.D{{Key: "a", Value: "b"}}, "x"}, AppendI([]bson.D{{{Key: "a", Value: "b"}}}, "x"))
	assert.Equal(t, []interface{}{bson.D{{Key: "a", Value: "b"}}, "x"}, AppendI(bson.A{bson.D{{Key: "a", Value: "b"}}}, "x"))
	assert.Equal(t, []interface{}{bson.D{{Key: "a", Value: "b"}}, "x"}, AppendI([]interface{}{bson.D{{Key: "a", Value: "b"}}}, "x"))
}

func TestAppendE(t *testing.T) {
	assert.Equal(t, bson.M{"a": "b", "c": "d"}, AppendE(bson.M{"a": "b"}, bson.E{Key: "c", Value: "d"}))
	assert.Equal(t, bson.D{{Key: "a", Value: "b"}, {Key: "c", Value: "d"}}, AppendE(bson.D{{Key: "a", Value: "b"}}, bson.E{Key: "c", Value: "d"}))
	assert.Equal(t, []bson.M{}, AppendE([]bson.M{}, bson.E{Key: "c", Value: "d"}))
}

func TestGetE(t *testing.T) {
	assert.Equal(t, "b", GetE(bson.M{"a": "b"}, "a"))
	assert.Nil(t, GetE(bson.M{"a": "b"}, "b"))
	assert.Equal(t, "b", GetE(bson.D{{Key: "a", Value: "b"}}, "a"))
	assert.Nil(t, GetE(bson.D{{Key: "a", Value: "b"}}, "b"))
	assert.Nil(t, GetE(bson.A{}, "b"))
}

func TestAnd(t *testing.T) {
	assert.Equal(t, bson.M{"x": "y"}, And(bson.M{}, "x", "y"))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "x", "y"))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", nil))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", bson.M(nil)))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", bson.D(nil)))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", bson.A(nil)))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", []bson.M(nil)))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", []bson.D(nil)))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", []bson.A(nil)))
	assert.Equal(t, bson.M{"x": "z"}, And(bson.M{"x": "z"}, "", []interface{}(nil)))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.M{"$or": []bson.M{{"a": "b"}}},
			bson.M{"x": "y"},
		},
	}, And(bson.M{"$or": []bson.M{{"a": "b"}}}, "x", "y"))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.M{"a": "b"},
			bson.M{"x": "y"},
		},
	}, And(bson.M{"$and": []bson.M{{"a": "b"}}}, "x", "y"))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.M{"a": "b"},
			bson.M{"x": "y"},
		},
	}, And(bson.M{"$and": []interface{}{bson.M{"a": "b"}}}, "x", "y"))

	assert.Equal(t, bson.D{{Key: "x", Value: "y"}}, And(bson.D{}, "x", "y"))
	assert.Equal(t, bson.D{{Key: "x", Value: "z"}}, And(bson.D{{Key: "x", Value: "z"}}, "x", "y"))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.D{{Key: "$or", Value: []bson.M{{"a": "b"}}}},
			bson.M{"x": "y"},
		},
	}, And(bson.D{{Key: "$or", Value: []bson.M{{"a": "b"}}}}, "x", "y"))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.M{"a": "b"},
			bson.M{"x": "y"},
		},
	}, And(bson.D{{Key: "$and", Value: []bson.M{{"a": "b"}}}}, "x", "y"))

	assert.Equal(t, bson.M{"$and": []interface{}{bson.M{}, "y"}}, And(bson.M{}, "", "y"))
	assert.Equal(t, bson.M{"$and": []interface{}{bson.D{}, "y"}}, And(bson.D{}, "", "y"))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.D{{Key: "$or", Value: []bson.M{{"a": "b"}}}},
			"y",
		},
	}, And(bson.D{{Key: "$or", Value: []bson.M{{"a": "b"}}}}, "", "y"))
	assert.Equal(t, bson.M{
		"$and": []interface{}{
			bson.M{"a": "b"},
			"y",
		},
	}, And(bson.D{{Key: "$and", Value: []bson.M{{"a": "b"}}}}, "", "y"))
}
