package event

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID            = id.EventID
	UserID        = accountdomain.UserID
	IntegrationID = id.IntegrationID
)

var (
	NewID        = id.NewEventID
	MustID       = id.MustEventID
	IDFrom       = id.EventIDFrom
	IDFromRef    = id.EventIDFromRef
	ErrInvalidID = id.ErrInvalidID
)
