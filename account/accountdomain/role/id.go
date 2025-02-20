package role

import (
	"github.com/reearth/reearthx/account/accountdomain"
)

type ID = accountdomain.RoleID

var NewID = accountdomain.NewRoleID

var MustID = accountdomain.MustRoleID

var IDFrom = accountdomain.RoleIDFrom

var IDFromRef = accountdomain.RoleIDFromRef

var ErrInvalidID = accountdomain.ErrInvalidID
