package repository

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
)

// EventType represents the type of asset event
type EventType string

const (
	EventTypeAssetCreated     EventType = "asset.created"
	EventTypeAssetUpdated     EventType = "asset.updated"
	EventTypeAssetDeleted     EventType = "asset.deleted"
	EventTypeAssetUploaded    EventType = "asset.uploaded"
	EventTypeAssetExtracted   EventType = "asset.extracted"
	EventTypeAssetTransferred EventType = "asset.transferred"
)

// AssetEvent represents an asset event
type AssetEvent struct {
	Type        EventType
	AssetID     id.ID
	WorkspaceID id.WorkspaceID
	ProjectID   id.ProjectID
	Status      string
	Error       string
}

// EventHandler is a function that handles asset events
type EventHandler func(ctx context.Context, event AssetEvent)
