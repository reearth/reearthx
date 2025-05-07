package asset

import (
	"context"
	"io"
)

type ZipExtractor interface {
	Extract(ctx context.Context, assetID AssetID, reader io.ReaderAt, size int64) error
}
