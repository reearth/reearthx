package repository

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

// PubSubRepository defines the interface for publishing and subscribing to asset events
type PubSubRepository interface {
	// PublishAssetCreated publishes an asset created event
	PublishAssetCreated(ctx context.Context, asset *entity.Asset) error

	// PublishAssetUpdated publishes an asset updated event
	PublishAssetUpdated(ctx context.Context, asset *entity.Asset) error

	// PublishAssetDeleted publishes an asset deleted event
	PublishAssetDeleted(ctx context.Context, assetID id.ID) error

	// PublishAssetUploaded publishes an asset uploaded event
	PublishAssetUploaded(ctx context.Context, asset *entity.Asset) error

	// PublishAssetExtracted publishes an asset extraction status event
	PublishAssetExtracted(ctx context.Context, asset *entity.Asset) error

	// PublishAssetTransferred publishes an asset transferred event
	PublishAssetTransferred(ctx context.Context, asset *entity.Asset) error

	// Subscribe registers a handler for a specific event type
	// Use "*" as eventType to subscribe to all events
	Subscribe(eventType EventType, handler EventHandler)

	// Unsubscribe removes a handler for a specific event type
	Unsubscribe(eventType EventType, handler EventHandler)
}
