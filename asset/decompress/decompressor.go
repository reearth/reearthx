package decompress

import (
	"context"

	"github.com/reearth/reearthx/asset"
)

type Decompressor interface {
	DecompressAsync(ctx context.Context, assetID asset.ID) error
}
