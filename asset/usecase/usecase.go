package assetusecase

import (
	"context"
	"io"
	"time"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

// DeliverOptions contains options for asset delivery
type DeliverOptions struct {
	Transform   bool              // Whether to transform the content
	ContentType string            // Optional content type override
	Headers     map[string]string // Additional response headers
	MaxAge      int               // Cache control max age in seconds
	Disposition string            // Content disposition (inline/attachment)
}

// DecompressStatus represents the status of a decompression job
type DecompressStatus struct {
	JobID       string
	AssetID     id.ID
	Status      string  // "pending", "processing", "completed", "failed"
	Progress    float64 // 0-100
	Error       string
	StartedAt   time.Time
	CompletedAt time.Time
}

type Usecase interface {
	// CreateAsset creates a new asset
	CreateAsset(ctx context.Context, asset *entity.Asset) *Result
	// GetAsset retrieves an asset by ID
	GetAsset(ctx context.Context, id id.ID) *Result
	// UpdateAsset updates an existing asset
	UpdateAsset(ctx context.Context, asset *entity.Asset) *Result
	// DeleteAsset removes an asset by ID
	DeleteAsset(ctx context.Context, id id.ID) *Result
	// UploadAssetContent uploads content for an asset with the given ID
	UploadAssetContent(ctx context.Context, id id.ID, content io.Reader) *Result
	// DownloadAssetContent retrieves the content of an asset by ID
	DownloadAssetContent(ctx context.Context, id id.ID) *Result
	// GetAssetUploadURL generates a URL for uploading content to an asset
	GetAssetUploadURL(ctx context.Context, id id.ID) *Result
	// ListAssets returns all assets
	ListAssets(ctx context.Context) *Result
	// DecompressZipContent decompresses zip content and returns a channel of decompressed files
	DecompressZipContent(ctx context.Context, content []byte) *Result
	// CompressToZip compresses the provided files into a zip archive
	CompressToZip(ctx context.Context, files map[string]io.Reader) *Result
	// DeleteAllAssetsInGroup deletes all assets in a group
	DeleteAllAssetsInGroup(ctx context.Context, groupID id.GroupID) *Result
	// DeliverAsset proxies the asset content with optional transformations
	DeliverAsset(ctx context.Context, id id.ID, options *DeliverOptions) *Result
	// GetDecompressStatus gets the current status of an async decompression
	GetDecompressStatus(ctx context.Context, jobID string) *Result
	// ListDecompressJobs lists all active decompression jobs
	ListDecompressJobs(ctx context.Context) *Result
}
