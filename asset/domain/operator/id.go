package operator

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
	ErrInvalidID     = id.ErrInvalidID
	NewIntegrationID = id.NewIntegrationID
)
