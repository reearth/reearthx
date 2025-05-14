package assetrepo

import (
	"context"

	id "github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/asset/assetdomain/asset"
	"github.com/reearth/reearthx/usecasex"
)

type AssetFilter struct {
	Sort       *usecasex.Sort
	Keyword    *string
	Pagination *usecasex.Pagination
}

type Asset interface {
	// Filtered returns a filtered list of assets based on the provided filter.
	Filtered(ProjectFilter) Asset

	// FindByProject retrieves assets by project ID with optional filtering.
	FindByProject(ctx context.Context, projectID id.ProjectID, filter AssetFilter) ([]*asset.Asset, *usecasex.PageInfo, error)

	// FindByID retrieves a single asset by its ID.
	FindByID(ctx context.Context, assetID id.AssetID) (*asset.Asset, error)

	// FindByIDs retrieves multiple assets by their IDs.
	FindByIDs(ctx context.Context, assetIDs id.AssetIDList) ([]*asset.Asset, error)

	// Save creates or updates an asset.
	Save(ctx context.Context, a *asset.Asset) error

	// Delete removes an asset by its ID.
	Delete(ctx context.Context, assetID id.AssetID) error
}

type AssetFile interface {
	// FindByID retrieves a file by asset ID.
	FindByID(ctx context.Context, assetID id.AssetID) (*asset.File, error)

	// Save creates or updates a file associated with an asset.
	Save(ctx context.Context, assetID id.AssetID, file *asset.File) error

	// SaveFlat saves a file and its related files.
	SaveFlat(ctx context.Context, assetID id.AssetID, file *asset.File, relatedFiles []*asset.File) error
}
