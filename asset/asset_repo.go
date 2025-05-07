package asset

import (
	"context"
)

type AssetRepository interface {
	Save(ctx context.Context, asset *Asset) error
	FindByID(ctx context.Context, id AssetID) (*Asset, error)
	FindByIDs(ctx context.Context, ids []AssetID) ([]*Asset, error)
	FindByGroup(ctx context.Context, groupID GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*Asset, int64, error)
	Delete(ctx context.Context, id AssetID) error
	DeleteMany(ctx context.Context, ids []AssetID) error
	UpdateExtractionStatus(ctx context.Context, id AssetID, status ExtractionStatus) error
}

type AssetFilter struct {
	Keyword string
}

type AssetSortType string

const (
	AssetSortTypeDate AssetSortType = "DATE"
	AssetSortTypeSize AssetSortType = "SIZE"
	AssetSortTypeName AssetSortType = "NAME"
)

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

type AssetSort struct {
	By        AssetSortType
	Direction SortDirection
}

type Pagination struct {
	Offset int64
	Limit  int64
}
