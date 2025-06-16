package integration

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID        = id.IntegrationID
	WebhookID = id.WebhookID
	UserID    = accountdomain.UserID
	ModelID   = id.ModelID
)

var (
	NewID        = id.NewIntegrationID
	NewWebhookID = id.NewWebhookID
	MustID       = id.MustIntegrationID
	IDFrom       = id.IntegrationIDFrom
	IDFromRef    = id.IntegrationIDFromRef
	ErrInvalidID = id.ErrInvalidID
)
