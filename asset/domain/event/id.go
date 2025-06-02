package event

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type ID = id.EventID
type UserID = accountdomain.UserID
type IntegrationID = id.IntegrationID

var NewID = id.NewEventID
var MustID = id.MustEventID
var IDFrom = id.EventIDFrom
var IDFromRef = id.EventIDFromRef
var ErrInvalidID = id.ErrInvalidID
