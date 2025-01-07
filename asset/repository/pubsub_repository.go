package repository

import (
	"context"

	"github.com/reearth/reearthx/asset/domain"
)

// EventType represents the type of asset event
type EventType string

const (
	// Asset events
	EventTypeAssetCreated     EventType = "ASSET_CREATED"
	EventTypeAssetUpdated     EventType = "ASSET_UPDATED"
	EventTypeAssetDeleted     EventType = "ASSET_DELETED"
	EventTypeAssetUploaded    EventType = "ASSET_UPLOADED"
	EventTypeAssetExtracted   EventType = "ASSET_EXTRACTED"
	EventTypeAssetTransferred EventType = "ASSET_TRANSFERRED"
)

// AssetEvent represents an event related to an asset
type AssetEvent struct {
	Type        EventType          `json:"type"`
	AssetID     domain.ID          `json:"asset_id"`
	WorkspaceID domain.WorkspaceID `json:"workspace_id,omitempty"`
	ProjectID   domain.ProjectID   `json:"project_id,omitempty"`
	Status      domain.Status      `json:"status,omitempty"`
	Error       string             `json:"error,omitempty"`
}

// EventHandler is a function that handles asset events
type EventHandler func(ctx context.Context, event AssetEvent)

// PubSubRepository defines the interface for publishing and subscribing to asset events
type PubSubRepository interface {
	// PublishAssetCreated publishes an asset created event
	PublishAssetCreated(ctx context.Context, asset *domain.Asset) error

	// PublishAssetUpdated publishes an asset updated event
	PublishAssetUpdated(ctx context.Context, asset *domain.Asset) error

	// PublishAssetDeleted publishes an asset deleted event
	PublishAssetDeleted(ctx context.Context, assetID domain.ID) error

	// PublishAssetUploaded publishes an asset uploaded event
	PublishAssetUploaded(ctx context.Context, asset *domain.Asset) error

	// PublishAssetExtracted publishes an asset extraction status event
	PublishAssetExtracted(ctx context.Context, asset *domain.Asset) error

	// PublishAssetTransferred publishes an asset transferred event
	PublishAssetTransferred(ctx context.Context, asset *domain.Asset) error

	// Subscribe registers a handler for a specific event type
	// Use "*" as eventType to subscribe to all events
	Subscribe(eventType EventType, handler EventHandler)

	// Unsubscribe removes a handler for a specific event type
	Unsubscribe(eventType EventType, handler EventHandler)
}
