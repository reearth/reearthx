package operator

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type ID = id.EventID
type UserID = accountdomain.UserID
type IntegrationID = id.IntegrationID

var ErrInvalidID = id.ErrInvalidID
var NewIntegrationID = id.NewIntegrationID
