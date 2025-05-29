package repo

import (
	"context"
	asset2 "github.com/reearth/reearthx/asset/domain/asset"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/usecasex"
)

type AssetRepository interface {
	SaveCMS(ctx context.Context, asset *asset2.Asset) error
	FindByID(ctx context.Context, id asset2.AssetID) (*asset2.Asset, error)
	FindByUUID(ctx context.Context, uuid string) (*asset2.Asset, error)
	FindByIDs(ctx context.Context, ids AssetIDList) ([]*asset2.Asset, error)
	FindByIDList(ctx context.Context, ids AssetIDList) (asset2.List, error)
	FindByGroup(ctx context.Context, groupID asset2.GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*asset2.Asset, int64, error)
	FindByProject(ctx context.Context, groupID asset2.GroupID, filter AssetFilter) (asset2.List, *usecasex.PageInfo, error)
	Search(context.Context, asset2.GroupID, AssetFilter) (asset2.List, *usecasex.PageInfo, error) // cms
	Delete(ctx context.Context, id asset2.AssetID) error
	DeleteMany(ctx context.Context, ids []asset2.AssetID) error
	BatchDelete(ctx context.Context, ids AssetIDList) error
	UpdateExtractionStatus(ctx context.Context, id asset2.AssetID, status asset2.ExtractionStatus) error
	Filtered(asset2.GroupFilter) AssetRepository                      // cms
	UpdateProject(ctx context.Context, from, to asset2.GroupID) error // cms

	TotalSizeByWorkspace(context.Context, accountdomain.WorkspaceID) (int64, error) //viz
	RemoveByProjectWithFile(context.Context, asset2.GroupID, any) error
	FindByWorkspaceProject(context.Context, accountdomain.WorkspaceID, *asset2.GroupID, AssetFilter) ([]*asset2.Asset, *usecasex.PageInfo, error)
	Save(ctx context.Context, asset *asset2.Asset) error // viz and flow
}

// cms
type AssetFile interface {
	FindByID(context.Context, asset2.AssetID) (*asset2.File, error)
	FindByIDs(context.Context, AssetIDList) (map[asset2.AssetID]*asset2.File, error)
	Save(context.Context, asset2.AssetID, *asset2.File) error
	SaveFlat(context.Context, asset2.AssetID, *asset2.File, []*asset2.File) error
}

type AssetFilter struct {
	Sort         *usecasex.Sort
	Keyword      *string
	Pagination   *usecasex.Pagination
	ContentTypes []string
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

type AssetIDList []asset2.AssetID

func (l AssetIDList) Add(id asset2.AssetID) AssetIDList {
	return append(l, id)
}

func (l AssetIDList) Strings() []string {
	strings := make([]string, len(l))
	for i, id := range l {
		strings[i] = id.String()
	}
	return strings
}
