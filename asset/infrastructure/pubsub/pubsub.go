package pubsub

import (
	"context"
	"reflect"
	"sync"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
	"github.com/reearth/reearthx/log"
)

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, topic string, msg interface{}) error
}

type subscription struct {
	eventType repository.EventType
	handler   repository.EventHandler
}

// AssetPubSub handles publishing and subscribing to asset events
type AssetPubSub struct {
	publisher     Publisher
	topic         string
	mu            sync.RWMutex
	subscriptions []subscription
}

var _ repository.PubSubRepository = (*AssetPubSub)(nil)

// NewAssetPubSub creates a new AssetPubSub instance
func NewAssetPubSub(publisher Publisher, topic string) *AssetPubSub {
	return &AssetPubSub{
		publisher: publisher,
		topic:     topic,
	}
}

// Subscribe registers a handler for a specific event type
func (p *AssetPubSub) Subscribe(eventType repository.EventType, handler repository.EventHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.subscriptions = append(p.subscriptions, subscription{
		eventType: eventType,
		handler:   handler,
	})
}

// Unsubscribe removes a handler for a specific event type
func (p *AssetPubSub) Unsubscribe(eventType repository.EventType, handler repository.EventHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerValue := reflect.ValueOf(handler)
	for i := len(p.subscriptions) - 1; i >= 0; i-- {
		s := p.subscriptions[i]
		if s.eventType == eventType && reflect.ValueOf(s.handler) == handlerValue {
			p.subscriptions = append(p.subscriptions[:i], p.subscriptions[i+1:]...)
		}
	}
}

// notify notifies all subscribers of an event
func (p *AssetPubSub) notify(ctx context.Context, event repository.AssetEvent) {
	p.mu.RLock()
	subs := make([]subscription, len(p.subscriptions))
	copy(subs, p.subscriptions)
	p.mu.RUnlock()

	for _, sub := range subs {
		if sub.eventType == event.Type || sub.eventType == "*" {
			sub.handler(ctx, event)
		}
	}
}

// PublishAssetCreated publishes an asset created event
func (p *AssetPubSub) PublishAssetCreated(ctx context.Context, asset *domain.Asset) error {
	event := repository.AssetEvent{
		Type:        repository.EventTypeAssetCreated,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}

	if err := p.publisher.Publish(ctx, p.topic, event); err != nil {
		log.Errorfc(ctx, "failed to publish asset created event: %v", err)
		return err
	}

	p.notify(ctx, event)
	return nil
}

// PublishAssetUpdated publishes an asset updated event
func (p *AssetPubSub) PublishAssetUpdated(ctx context.Context, asset *domain.Asset) error {
	event := repository.AssetEvent{
		Type:        repository.EventTypeAssetUpdated,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}

	if err := p.publisher.Publish(ctx, p.topic, event); err != nil {
		log.Errorfc(ctx, "failed to publish asset updated event: %v", err)
		return err
	}

	p.notify(ctx, event)
	return nil
}

// PublishAssetDeleted publishes an asset deleted event
func (p *AssetPubSub) PublishAssetDeleted(ctx context.Context, assetID domain.ID) error {
	event := repository.AssetEvent{
		Type:    repository.EventTypeAssetDeleted,
		AssetID: assetID,
	}

	if err := p.publisher.Publish(ctx, p.topic, event); err != nil {
		log.Errorfc(ctx, "failed to publish asset deleted event: %v", err)
		return err
	}

	p.notify(ctx, event)
	return nil
}

// PublishAssetUploaded publishes an asset uploaded event
func (p *AssetPubSub) PublishAssetUploaded(ctx context.Context, asset *domain.Asset) error {
	event := repository.AssetEvent{
		Type:        repository.EventTypeAssetUploaded,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}

	if err := p.publisher.Publish(ctx, p.topic, event); err != nil {
		log.Errorfc(ctx, "failed to publish asset uploaded event: %v", err)
		return err
	}

	p.notify(ctx, event)
	return nil
}

// PublishAssetExtracted publishes an asset extraction status event
func (p *AssetPubSub) PublishAssetExtracted(ctx context.Context, asset *domain.Asset) error {
	event := repository.AssetEvent{
		Type:        repository.EventTypeAssetExtracted,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}

	if err := p.publisher.Publish(ctx, p.topic, event); err != nil {
		log.Errorfc(ctx, "failed to publish asset extracted event: %v", err)
		return err
	}

	p.notify(ctx, event)
	return nil
}

// PublishAssetTransferred publishes an asset transferred event
func (p *AssetPubSub) PublishAssetTransferred(ctx context.Context, asset *domain.Asset) error {
	event := repository.AssetEvent{
		Type:        repository.EventTypeAssetTransferred,
		AssetID:     asset.ID(),
		WorkspaceID: asset.WorkspaceID(),
		ProjectID:   asset.ProjectID(),
		Status:      asset.Status(),
		Error:       asset.Error(),
	}

	if err := p.publisher.Publish(ctx, p.topic, event); err != nil {
		log.Errorfc(ctx, "failed to publish asset transferred event: %v", err)
		return err
	}

	p.notify(ctx, event)
	return nil
}
