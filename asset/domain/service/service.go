package service

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/id"
)

// Storage defines the interface for asset storage operations
type Storage interface {
	Upload(ctx context.Context, workspaceID id.WorkspaceID, name string, content io.Reader) (string, int64, error)
	Download(ctx context.Context, url string) (io.ReadCloser, error)
	Delete(ctx context.Context, url string) error
}

// Extractor defines the interface for asset extraction operations
type Extractor interface {
	Extract(ctx context.Context, asset *entity.Asset) error
	IsExtractable(contentType string) bool
}

// AssetService defines the interface for asset domain service
type AssetService interface {
	Upload(ctx context.Context, workspaceID id.WorkspaceID, name string, content io.Reader) (*entity.Asset, error)
	Download(ctx context.Context, assetID id.ID) (io.ReadCloser, error)
	Extract(ctx context.Context, assetID id.ID) error
	Move(ctx context.Context, assetID id.ID, projectID id.ProjectID, groupID id.GroupID) error
	Delete(ctx context.Context, assetID id.ID) error
	SetEventPublisher(publisher event.Publisher)
}

// GroupService defines the interface for group domain service
type GroupService interface {
	Create(ctx context.Context, name string, policy string) (*entity.Group, error)
	Update(ctx context.Context, id id.GroupID, name string, policy string, description string) (*entity.Group, error)
	Delete(ctx context.Context, id id.GroupID) error
	SetEventPublisher(publisher event.Publisher)
}
