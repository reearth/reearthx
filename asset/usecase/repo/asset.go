package repo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/usecase/gateway"
	"github.com/reearth/reearthx/asset/usecase/interfaces"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/id"

	"github.com/reearth/reearthx/usecasex"
)

type AssetFilter struct {
	Sort         *usecasex.Sort
	SortType     *asset.SortType
	Keyword      *string
	Pagination   *usecasex.Pagination
	ContentTypes []string
}

type Asset interface {
	Filtered(ProjectFilter) Asset
	FindByID(context.Context, id.AssetID) (*asset.Asset, error)
	FindByUUID(context.Context, string) (*asset.Asset, error)
	FindByIDs(context.Context, id.AssetIDList) (asset.List, error)
	Search(context.Context, id.ProjectID, AssetFilter) (asset.List, *usecasex.PageInfo, error)
	Save(context.Context, *asset.Asset) error
	Delete(context.Context, id.AssetID) error // save as remove at viz
	BatchDelete(context.Context, id.AssetIDList) error

	TotalSizeByWorkspace(context.Context, accountdomain.WorkspaceID) (int64, error)
	RemoveByProjectWithFile(context.Context, id.ProjectID, gateway.File) error // viz
	FindByWorkspaceProject(
		context.Context,
		accountdomain.WorkspaceID,
		*id.ProjectID,
		AssetFilter,
	) ([]*asset.Asset, *usecasex.PageInfo, error) // viz
	FindByWorkspace(
		context.Context,
		accountdomain.WorkspaceID,
		AssetFilter,
	) ([]*asset.Asset, *interfaces.PageBasedInfo, error) // flow
}

// cms
type AssetFile interface {
	FindByID(context.Context, id.AssetID) (*asset.File, error)
	FindByIDs(context.Context, id.AssetIDList) (map[id.AssetID]*asset.File, error)
	Save(context.Context, id.AssetID, *asset.File) error
	SaveFlat(context.Context, id.AssetID, *asset.File, []*asset.File) error
}
