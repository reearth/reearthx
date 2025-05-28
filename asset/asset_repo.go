package asset

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/usecasex"
)

type AssetRepository interface {
	Save(ctx context.Context, asset *Asset) error
	FindByID(ctx context.Context, id AssetID) (*Asset, error)
	FindByUUID(ctx context.Context, uuid string) (*Asset, error)
	FindByIDs(ctx context.Context, ids AssetIDList) ([]*Asset, error)
	FindByIDList(ctx context.Context, ids AssetIDList) (List, error)
	FindByGroup(ctx context.Context, groupID GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*Asset, int64, error)
	FindByProject(ctx context.Context, groupID GroupID, filter AssetFilter) (List, *usecasex.PageInfo, error)
	Search(context.Context, GroupID, AssetFilter) (List, *usecasex.PageInfo, error) // cms
	Delete(ctx context.Context, id AssetID) error
	DeleteMany(ctx context.Context, ids []AssetID) error
	BatchDelete(ctx context.Context, ids AssetIDList) error
	UpdateExtractionStatus(ctx context.Context, id AssetID, status ExtractionStatus) error
	Filtered(GroupFilter) AssetRepository                      // cms
	UpdateProject(ctx context.Context, from, to GroupID) error // cms

	TotalSizeByWorkspace(context.Context, accountdomain.WorkspaceID) (int64, error) //viz
	RemoveByProjectWithFile(context.Context, GroupID, any) error
	FindByWorkspaceProject(context.Context, accountdomain.WorkspaceID, *GroupID, AssetFilter) ([]*Asset, *usecasex.PageInfo, error)
	SaveViz(ctx context.Context, asset *Asset) error // viz
}

// cms
type AssetFile interface {
	FindByID(context.Context, AssetID) (*File, error)
	FindByIDs(context.Context, AssetIDList) (map[AssetID]*File, error)
	Save(context.Context, AssetID, *File) error
	SaveFlat(context.Context, AssetID, *File, []*File) error
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

type AssetIDList []AssetID

func (l AssetIDList) Add(id AssetID) AssetIDList {
	return append(l, id)
}

func (l AssetIDList) Strings() []string {
	strings := make([]string, len(l))
	for i, id := range l {
		strings[i] = id.String()
	}
	return strings
}
