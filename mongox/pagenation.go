package mongox

import (
	"errors"

	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
)

type pagination struct {
	Before *string
	After  *string
	First  *int
	Last   *int
}

func PaginationFrom(p *usecasex.Pagination) *pagination {
	if p == nil {
		return nil
	}
	return &pagination{
		Before: (*string)(p.Before),
		After:  (*string)(p.After),
		First:  p.First,
		Last:   p.Last,
	}
}

func (p *pagination) SortDirection() int {
	if p != nil && p.Last != nil {
		return -1
	}
	return 1
}

func (p *pagination) Parameters() (limit int64, op string, cursor *string, err error) {
	if first, after := p.First, p.After; first != nil {
		limit = int64(*first)
		op = "$gt"
		cursor = after
		return
	}
	if last, before := p.Last, p.Before; last != nil {
		limit = int64(*last)
		op = "$lt"
		cursor = before
		return
	}
	return 0, "", nil, errors.New("neither first nor last are set")
}

func (p *pagination) SortOptions(sort *string, key string) (any, string) {
	var sortOptions bson.D
	var sortKey = ""
	if sort != nil && len(*sort) > 0 && *sort != "id" {
		sortKey = *sort
		sortOptions = append(sortOptions, bson.E{Key: sortKey, Value: p.SortDirection()})
	}
	sortOptions = append(sortOptions, bson.E{Key: key, Value: p.SortDirection()})
	return sortOptions, sortKey
}
