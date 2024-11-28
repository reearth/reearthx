package mongox

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

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

func TestAndEmptyFilter(t *testing.T) {
	filter := bson.M{}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"b": bson.M{"c": 2}},
		},
	}
	actual := AddCondition(filter, "b", bson.M{"c": 2})
	assert.Equal(t, expected, actual)
}

func TestAndEmptyKey(t *testing.T) {
	filter := bson.M{"a": 1}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"a": 1},
			bson.M{"c": 3},
		},
		"a": 1,
	}
	actual := AddCondition(filter, "", bson.M{"c": 3})
	assert.Equal(t, expected, actual)
}

func TestAndExistingAddCondition(t *testing.T) {
	filter := bson.M{
		"$and": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
			bson.M{"c": bson.M{"d": 3}},
		},
	}
	actual := AddCondition(filter, "c", bson.M{"d": 3})
	assert.Equal(t, expected, actual)
}

func TestAndWithOrAndEmptyKey(t *testing.T) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"a": 1},
					bson.M{"b": 2},
				},
			},
			bson.M{"c": 3},
		},
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	actual := AddCondition(filter, "", bson.M{"c": 3})
	assert.Equal(t, expected, actual)
}

func TestAndComplexFilter(t *testing.T) {
	filter := bson.M{
		"$and": bson.A{
			bson.M{"x": 10},
			bson.M{"y": 20},
		},
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"x": 10},
			bson.M{"y": 20},
			bson.M{"c": bson.M{"d": 3}},
		},
		"$or": bson.A{bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	actual := AddCondition(filter, "c", bson.M{"d": 3})
	assert.Equal(t, expected, actual)
}

func TestAndNilFilter(t *testing.T) {
	var filter interface{}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"b": bson.M{"c": 2}},
		},
	}
	actual := AddCondition(filter, "b", bson.M{"c": 2})
	assert.Equal(t, expected, actual)
}

func TestAndEmptySliceCondition(t *testing.T) {
	filter := bson.M{"a": 1}
	expected := bson.M{"a": 1}
	actual := AddCondition(filter, "b", bson.A{})
	assert.Equal(t, expected, actual)
}

func TestAndProjectRefetchFilter(t *testing.T) {
	team := "team_id_example"
	last := "last_project_id"
	updatedat := 1654849072592
	filter := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"deleted": false},
					bson.M{"deleted": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"coresupport": true},
					bson.M{"coresupport": bson.M{"$exists": false}},
				},
			},
		},
		"team": team,
	}

	condition := bson.M{
		"$or": bson.A{
			bson.M{"updatedat": bson.M{"$lt": updatedat}},
			bson.M{"id": bson.M{"$lt": last}, "updatedat": updatedat},
		},
	}

	expected := bson.M{
		"$and": bson.A{
			bson.M{"$or": bson.A{
				bson.M{"deleted": false},
				bson.M{"deleted": bson.M{"$exists": false}},
			},
			},
			bson.M{"$or": bson.A{
				bson.M{"coresupport": true},
				bson.M{"coresupport": bson.M{"$exists": false}},
			},
			},
			bson.M{"$or": bson.A{
				bson.M{"updatedat": bson.M{"$lt": updatedat}},
				bson.M{"id": bson.M{"$lt": last}, "updatedat": updatedat},
			},
			},
		},
		"team": team,
	}

	actual := AddCondition(filter, "", condition)
	assert.Equal(t, expected, actual)
}

func TestAddConditionEmptyFilter(t *testing.T) {
	filter := bson.M{}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"b": bson.M{"c": 2}},
		},
	}
	actual := AddCondition(filter, "b", bson.M{"c": 2})
	assert.Equal(t, expected, actual)
}

func TestAddConditionEmptyKey(t *testing.T) {
	filter := bson.M{"a": 1}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"a": 1},
			bson.M{"c": 3},
		},
		"a": 1,
	}
	actual := AddCondition(filter, "", bson.M{"c": 3})
	assert.Equal(t, expected, actual)
}

func TestAddConditionExistingAddCondition(t *testing.T) {
	filter := bson.M{
		"$and": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
			bson.M{"c": bson.M{"d": 3}},
		},
	}
	actual := AddCondition(filter, "c", bson.M{"d": 3})
	assert.Equal(t, expected, actual)
}

func TestAddConditionWithOrAndEmptyKey(t *testing.T) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"a": 1},
					bson.M{"b": 2},
				},
			},
			bson.M{"c": 3},
		},
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	actual := AddCondition(filter, "", bson.M{"c": 3})
	assert.Equal(t, expected, actual)
}

func TestAddConditionComplexFilter(t *testing.T) {
	filter := bson.M{
		"$and": bson.A{
			bson.M{"x": 10},
			bson.M{"y": 20},
		},
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"x": 10},
			bson.M{"y": 20},
			bson.M{"c": bson.M{"d": 3}},
		},
		"$or": bson.A{
			bson.M{"a": 1},
			bson.M{"b": 2},
		},
	}
	actual := AddCondition(filter, "c", bson.M{"d": 3})
	assert.Equal(t, expected, actual)
}

func TestAddConditionNilFilter(t *testing.T) {
	var filter interface{}
	expected := bson.M{
		"$and": bson.A{
			bson.M{"b": bson.M{"c": 2}},
		},
	}
	actual := AddCondition(filter, "b", bson.M{"c": 2})
	assert.Equal(t, expected, actual)
}

func TestAddConditionEmptySliceCondition(t *testing.T) {
	filter := bson.M{"a": 1}
	expected := bson.M{"a": 1}
	actual := AddCondition(filter, "b", bson.A{})
	assert.Equal(t, expected, actual)
}

func TestAddConditionProjectRefetchFilter_bsonM(t *testing.T) {
	team := "team_id_example"
	last := "last_project_id"
	updatedat := 1111111111111
	filter := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"deleted": false},
					bson.M{"deleted": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"coresupport": true},
					bson.M{"coresupport": bson.M{"$exists": false}},
				},
			},
		},
		"team": team,
	}

	condition := bson.M{
		"$or": bson.A{
			bson.M{"updatedat": bson.M{"$lt": updatedat}},
			bson.M{"id": bson.M{"$lt": last}, "updatedat": updatedat},
		},
	}

	expected := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"deleted": false},
					bson.M{"deleted": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"coresupport": true},
					bson.M{"coresupport": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"updatedat": bson.M{"$lt": updatedat}},
					bson.M{"id": bson.M{"$lt": last}, "updatedat": updatedat},
				},
			},
		},
		"team": team,
	}

	actual := AddCondition(filter, "", condition)
	JSONEqAny(t, expected, actual)
	assert.Equal(t, expected, actual)
}

func TestAddConditionProjectRefetchFilter_bsonA(t *testing.T) {
	team := "team_id_example"
	last := "last_project_id"
	updatedat := 1111111111111
	filter := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"deleted": false},
					bson.M{"deleted": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"coresupport": true},
					bson.M{"coresupport": bson.M{"$exists": false}},
				},
			},
		},
		"team": team,
	}

	condition := bson.M{
		"$or": []bson.M{
			{"updatedat": bson.M{"$lt": updatedat}},
			{"id": bson.M{"$lt": last}, "updatedat": updatedat},
		},
	}
	expected := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"deleted": false},
					bson.M{"deleted": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"coresupport": true},
					bson.M{"coresupport": bson.M{"$exists": false}},
				},
			},
			bson.M{
				"$or": []bson.M{
					{"updatedat": bson.M{"$lt": updatedat}},
					{"id": bson.M{"$lt": last}, "updatedat": updatedat},
				},
			},
		},
		"team": team,
	}

	actual := AddCondition(filter, "", condition)
	JSONEqAny(t, expected, actual)
	assert.Equal(t, expected, actual)
}

func TestAddConditionProjectRefetchFilterWithKeyword(t *testing.T) {
	team := "team_id_example"
	last := "last_project_id"
	updatedat := 1654849072592

	filter := bson.M{
		"$and": []bson.M{
			{
				"$and": []bson.M{
					{
						"$or": []bson.M{
							{"deleted": false},
							{"deleted": bson.M{"$exists": false}},
						},
					},
					{
						"$or": []bson.M{
							{"coresupport": true},
							{"coresupport": bson.M{"$exists": false}},
						},
					},
				},
			},
			{"team": team},
			{"name": bson.M{"$regex": bson.M{"pattern": ".*test.*", "options": "i"}}},
		},
	}

	condition := bson.M{
		"$or": []bson.M{
			{"updatedat": bson.M{"$lt": updatedat}},
			{"id": bson.M{"$lt": last}, "updatedat": updatedat},
		},
	}

	expected := bson.M{
		"$and": []bson.M{
			{
				"$and": []bson.M{
					{
						"$or": []bson.M{
							{"deleted": false}, {"deleted": bson.M{"$exists": false}},
						},
					},
					{
						"$or": []bson.M{
							{"coresupport": true}, {"coresupport": bson.M{"$exists": false}},
						},
					},
				},
			},
			{"team": team},
			{
				"name": bson.M{
					"$regex": bson.M{
						"pattern": ".*test.*",
						"options": "i",
					},
				},
			},
			{
				"$or": []bson.M{
					{"updatedat": bson.M{"$lt": updatedat}},
					{"id": bson.M{"$lt": last}, "updatedat": updatedat},
				},
			},
		},
	}

	actual := AddCondition(filter, "", condition)
	JSONEqAny(t, expected, actual)
	assert.Equal(t, expected, actual)
}

func JSONEqAny(t *testing.T, expected interface{}, actual interface{}) {
	e, err := json.Marshal(expected)
	if err != nil {
		fmt.Println(err.Error())
	}
	a, err := json.Marshal(actual)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.JSONEq(t, string(e), string(a))
}
