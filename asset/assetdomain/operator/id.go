package operator

import (
	"github.com/reearth/reearthx/account/accountdomain"
	id "github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/idx"
)

type ID = id.EventID
type UserID = accountdomain.UserID
type IntegrationID = id.IntegrationID

var ErrInvalidID = idx.ErrInvalidID
var NewIntegrationID = id.NewIntegrationID
