package event

import (
	"github.com/reearth/reearthx/account/accountdomain"
	id "github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/idx"
)

type ID = id.EventID
type UserID = accountdomain.UserID
type IntegrationID = id.IntegrationID

var NewID = id.NewEventID
var MustID = id.MustEventID
var IDFrom = id.EventIDFrom
var IDFromRef = id.EventIDFromRef
var ErrInvalidID = idx.ErrInvalidID
