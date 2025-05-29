package repo

import (
	"context"
	"github.com/reearth/reearthx/asset/domain/asset"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/usecasex"
)

type AssetRepository interface {
	SaveCMS(ctx context.Context, asset *asset.Asset) error
	FindByID(ctx context.Context, id asset.ID) (*asset.Asset, error)
	FindByUUID(ctx context.Context, uuid string) (*asset.Asset, error)
	FindByIDs(ctx context.Context, ids asset.IDList) ([]*asset.Asset, error)
	FindByIDList(ctx context.Context, ids asset.IDList) (asset.List, error)
	FindByGroup(ctx context.Context, groupID asset.GroupID, filter asset.Filter, sort asset.Sort, pagination asset.Pagination) ([]*asset.Asset, int64, error)
	FindByProject(ctx context.Context, groupID asset.GroupID, filter asset.Filter) (asset.List, *usecasex.PageInfo, error)
	Search(context.Context, asset.GroupID, asset.Filter) (asset.List, *usecasex.PageInfo, error) // cms
	Delete(ctx context.Context, id asset.ID) error
	DeleteMany(ctx context.Context, ids []asset.ID) error
	BatchDelete(ctx context.Context, ids asset.IDList) error
	UpdateExtractionStatus(ctx context.Context, id asset.ID, status asset.ExtractionStatus) error
	Filtered(asset.GroupFilter) AssetRepository                      // cms
	UpdateProject(ctx context.Context, from, to asset.GroupID) error // cms

	TotalSizeByWorkspace(context.Context, accountdomain.WorkspaceID) (int64, error) //viz
	RemoveByProjectWithFile(context.Context, asset.GroupID, any) error
	FindByWorkspaceProject(context.Context, accountdomain.WorkspaceID, *asset.GroupID, asset.Filter) ([]*asset.Asset, *usecasex.PageInfo, error)
	Save(ctx context.Context, asset *asset.Asset) error // viz and flow
}

// cms
type AssetFile interface {
	FindByID(context.Context, asset.ID) (*asset.File, error)
	FindByIDs(context.Context, asset.IDList) (map[asset.ID]*asset.File, error)
	Save(context.Context, asset.ID, *asset.File) error
	SaveFlat(context.Context, asset.ID, *asset.File, []*asset.File) error
}
