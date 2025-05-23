package repo

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/asset"
)

type AssetUpload interface {
	Save(ctx context.Context, upload *asset.Upload) error
	FindByID(ctx context.Context, uuid string) (*asset.Upload, error)
}
