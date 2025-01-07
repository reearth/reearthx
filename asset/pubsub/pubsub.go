package pubsub

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

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, topic string, msg interface{}) error
}

// AssetPubSub handles publishing of asset events
type AssetPubSub struct {
	publisher Publisher
	topic     string
}

// NewAssetPubSub creates a new AssetPubSub instance
func NewAssetPubSub(publisher Publisher, topic string) *AssetPubSub {
	return &AssetPubSub{
		publisher: publisher,
		topic:     topic,
	}
}

// PublishAssetCreated publishes an asset created event
func (p *AssetPubSub) PublishAssetCreated(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, EventTypeAssetCreated, asset)
}

// PublishAssetUpdated publishes an asset updated event
func (p *AssetPubSub) PublishAssetUpdated(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, EventTypeAssetUpdated, asset)
}

// PublishAssetDeleted publishes an asset deleted event
func (p *AssetPubSub) PublishAssetDeleted(ctx context.Context, assetID domain.ID) error {
	event := AssetEvent{
		Type:    EventTypeAssetDeleted,
		AssetID: assetID,
	}
	return p.publisher.Publish(ctx, p.topic, event)
}

// PublishAssetUploaded publishes an asset uploaded event
func (p *AssetPubSub) PublishAssetUploaded(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, EventTypeAssetUploaded, asset)
}

// PublishAssetExtracted publishes an asset extraction status event
func (p *AssetPubSub) PublishAssetExtracted(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, EventTypeAssetExtracted, asset)
}

// PublishAssetTransferred publishes an asset transferred event
func (p *AssetPubSub) PublishAssetTransferred(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, EventTypeAssetTransferred, asset)
}

// publishAssetEvent is a helper function to publish asset events with common fields
func (p *AssetPubSub) publishAssetEvent(ctx context.Context, eventType EventType, asset *domain.Asset) error {
	event := AssetEvent{
		Type:        eventType,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}
	return p.publisher.Publish(ctx, p.topic, event)
}
