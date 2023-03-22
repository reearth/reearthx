package mongox

import (
	"reflect"
	"strings"

	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const prefix = "re_"

type Index struct {
	Name   string
	Key    bson.D
	Unique bool
	Filter bson.M `bson:"partialFilterExpression"`
}

func IndexFromKey(key string, unique bool) Index {
	return Index{
		Name:   prefix + key,
		Key:    toKeyBSON(key),
		Unique: unique,
	}
}

func IndexFromKeys(keys []string, unique bool) []Index {
	return lo.Map(keys, func(k string, _ int) Index {
		return IndexFromKey(k, unique)
	})
}

func toKeyBSON(key string) bson.D {
	return lo.Map(
		strings.Split(key, ","),
		func(k string, _ int) bson.E {
			k = strings.TrimSpace(k)
			v := 1
			if strings.HasPrefix(k, "!") {
				v = -1
			}
			k = strings.TrimPrefix(k, "!")
			return bson.E{
				Key:   strings.TrimSpace(k),
				Value: v,
			}
		},
	)
}

func (i Index) Normalize() Index {
	b, err := bson.Marshal(i)
	if err != nil {
		return i
	}
	i2 := Index{}
	if err := bson.Unmarshal(b, &i2); err != nil {
		return i
	}
	return i2
}

func (i Index) Model() mongo.IndexModel {
	o := options.Index().SetName(i.Name)
	if i.Unique {
		o.SetUnique(i.Unique)
	}
	if i.Filter != nil {
		o.SetPartialFilterExpression(i.Filter)
	}
	return mongo.IndexModel{
		Keys:    i.Key,
		Options: o,
	}
}

func (i Index) Equal(j Index) bool {
	e := reflect.DeepEqual(i, j)
	return e
}

type IndexList []Index

func (l IndexList) Names() []string {
	return lo.Map(l, func(i Index, _ int) string { return i.Name })
}

func (l IndexList) NamesWithoutPrefix() []string {
	return lo.Map(l, func(i Index, _ int) string { return strings.TrimPrefix(i.Name, prefix) })
}

func (l IndexList) Models() []mongo.IndexModel {
	return lo.Map(l, func(i Index, _ int) mongo.IndexModel { return i.Model() })
}

func (l IndexList) Normalize() []Index {
	return lo.Map(l, func(i Index, _ int) Index { return i.Normalize() })
}

func (l IndexList) AddNamePrefix() IndexList {
	return lo.Map(l, func(i Index, _ int) Index {
		if i.Name != "" && !strings.HasPrefix(i.Name, prefix) {
			i.Name = prefix + i.Name
		}
		return i
	})
}

func (l IndexList) RemoveDefaultIndex() IndexList {
	return lo.Filter(l, func(i Index, _ int) bool {
		return i.Name != "_id_"
	})
}

type IndexResult util.DiffResult[Index]

func (i IndexResult) AddedNames() []string {
	return IndexList(i.Added).NamesWithoutPrefix()
}

func (i IndexResult) UpdatedNames() []string {
	return IndexList(util.DiffResult[Index](i).UpdatedNext()).NamesWithoutPrefix()
}

func (i IndexResult) DeletedNames() []string {
	return IndexList(i.Deleted).NamesWithoutPrefix()
}
