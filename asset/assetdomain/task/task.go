package task

import (
	"github.com/reearth/reearthx/asset/assetdomain/event"
	"github.com/reearth/reearthx/asset/assetdomain/integration"
)

type Payload struct {
	DecompressAsset *DecompressAssetPayload
	CompressAsset   *CompressAssetPayload
	Webhook         *WebhookPayload
}

type DecompressAssetPayload struct {
	AssetID string
	Path    string
}

func (t *DecompressAssetPayload) Payload() Payload {
	return Payload{
		DecompressAsset: t,
	}
}

type CompressAssetPayload struct {
	AssetID string
}

func (t *CompressAssetPayload) Payload() Payload {
	return Payload{
		CompressAsset: t,
	}
}

type WebhookPayload struct {
	Webhook  *integration.Webhook
	Event    *event.Event[any]
	Override any
}

func (t WebhookPayload) Payload() Payload {
	return Payload{
		Webhook: &t,
	}
}
