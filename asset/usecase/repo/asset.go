package repo

import (
	"context"
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/file"
	"github.com/reearth/reearthx/asset/domain/id"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/usecasex"
)

type Asset interface {
	SaveCMS(ctx context.Context, asset *asset.Asset) error
	FindByID(ctx context.Context, id id.ID) (*asset.Asset, error)
	FindByUUID(ctx context.Context, uuid string) (*asset.Asset, error)
	FindByIDs(ctx context.Context, ids id.List) ([]*asset.Asset, error)
	FindByIDList(ctx context.Context, ids id.List) (asset.List, error)
	FindByGroup(ctx context.Context, groupID id.GroupID, filter asset.Filter, sort asset.Sort, pagination asset.Pagination) ([]*asset.Asset, int64, error)
	FindByProject(ctx context.Context, groupID id.GroupID, filter asset.Filter) (asset.List, *usecasex.PageInfo, error)
	Search(context.Context, id.GroupID, asset.Filter) (asset.List, *usecasex.PageInfo, error) // cms
	Delete(ctx context.Context, id id.ID) error
	DeleteMany(ctx context.Context, ids []id.ID) error
	BatchDelete(ctx context.Context, list id.List) error
	UpdateExtractionStatus(ctx context.Context, id id.ID, status asset.ExtractionStatus) error
	Filtered(any) Asset                                           // cms
	UpdateProject(ctx context.Context, from, to id.GroupID) error // cms

	TotalSizeByWorkspace(context.Context, accountdomain.WorkspaceID) (int64, error) //viz
	RemoveByProjectWithFile(context.Context, id.GroupID, any) error
	FindByWorkspaceProject(context.Context, accountdomain.WorkspaceID, *id.GroupID, asset.Filter) ([]*asset.Asset, *usecasex.PageInfo, error)
	Save(ctx context.Context, asset *asset.Asset) error // viz and flow
}

// cms
type AssetFile interface {
	FindByID(context.Context, id.ID) (*file.File, error)
	FindByIDs(context.Context, id.List) (map[id.ID]*file.File, error)
	Save(context.Context, id.ID, *file.File) error
	SaveFlat(context.Context, id.ID, *file.File, []*file.File) error
}
