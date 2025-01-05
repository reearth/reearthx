package repository

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset"
)

// Decompressor defines the interface for compression and decompression operations
type Decompressor interface {
	// DecompressAsync asynchronously decompresses a zip file identified by assetID
	DecompressAsync(ctx context.Context, assetID asset.ID) error

	// DecompressWithContent decompresses zip content directly without downloading
	DecompressWithContent(ctx context.Context, assetID asset.ID, content []byte) error

	// CompressAsync asynchronously compresses files into a zip archive
	CompressAsync(ctx context.Context, assetID asset.ID, files []asset.ID) error

	// CompressWithContent compresses the provided content into a zip archive
	CompressWithContent(ctx context.Context, assetID asset.ID, files map[string]io.Reader) error

	// GetCompressionStatus returns the current status of a compression/decompression operation
	GetStatus(ctx context.Context, assetID asset.ID) (asset.Status, error)

	// CancelOperation cancels an ongoing compression/decompression operation
	CancelOperation(ctx context.Context, assetID asset.ID) error
}
