package pubsub

import (
	"context"

	"github.com/reearth/reearthx/asset/domain"
)

type AssetEvent struct {
	Type    string    `json:"type"`
	AssetID domain.ID `json:"asset_id"`
}

type Publisher interface {
	Publish(ctx context.Context, topic string, msg interface{}) error
}

type AssetPubSub struct {
	publisher Publisher
	topic     string
}

func NewAssetPubSub(publisher Publisher, topic string) *AssetPubSub {
	return &AssetPubSub{
		publisher: publisher,
		topic:     topic,
	}
}

func (p *AssetPubSub) PublishAssetEvent(ctx context.Context, eventType string, assetID domain.ID) error {
	event := AssetEvent{
		Type:    eventType,
		AssetID: assetID,
	}
	return p.publisher.Publish(ctx, p.topic, event)
}
