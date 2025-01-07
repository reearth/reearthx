package pubsub

import (
	"context"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
)

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, topic string, msg interface{}) error
}

// AssetPubSub handles publishing of asset events
type AssetPubSub struct {
	publisher Publisher
	topic     string
}

var _ repository.PubSubRepository = (*AssetPubSub)(nil)

// NewAssetPubSub creates a new AssetPubSub instance
func NewAssetPubSub(publisher Publisher, topic string) *AssetPubSub {
	return &AssetPubSub{
		publisher: publisher,
		topic:     topic,
	}
}

// PublishAssetCreated publishes an asset created event
func (p *AssetPubSub) PublishAssetCreated(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, repository.EventTypeAssetCreated, asset)
}

// PublishAssetUpdated publishes an asset updated event
func (p *AssetPubSub) PublishAssetUpdated(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, repository.EventTypeAssetUpdated, asset)
}

// PublishAssetDeleted publishes an asset deleted event
func (p *AssetPubSub) PublishAssetDeleted(ctx context.Context, assetID domain.ID) error {
	event := repository.AssetEvent{
		Type:    repository.EventTypeAssetDeleted,
		AssetID: assetID,
	}
	return p.publisher.Publish(ctx, p.topic, event)
}

// PublishAssetUploaded publishes an asset uploaded event
func (p *AssetPubSub) PublishAssetUploaded(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, repository.EventTypeAssetUploaded, asset)
}

// PublishAssetExtracted publishes an asset extraction status event
func (p *AssetPubSub) PublishAssetExtracted(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, repository.EventTypeAssetExtracted, asset)
}

// PublishAssetTransferred publishes an asset transferred event
func (p *AssetPubSub) PublishAssetTransferred(ctx context.Context, asset *domain.Asset) error {
	return p.publishAssetEvent(ctx, repository.EventTypeAssetTransferred, asset)
}

// publishAssetEvent is a helper function to publish asset events with common fields
func (p *AssetPubSub) publishAssetEvent(ctx context.Context, eventType repository.EventType, asset *domain.Asset) error {
	event := repository.AssetEvent{
		Type:        eventType,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}
	return p.publisher.Publish(ctx, p.topic, event)
}
