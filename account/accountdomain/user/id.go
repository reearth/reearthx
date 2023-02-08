package user

import (
	"github.com/reearth/reearthx/account/accountdomain/id"
)

type ID = id.UserID
type WorkspaceID = id.WorkspaceID

var NewID = id.NewUserID
var NewWorkspaceID = id.NewWorkspaceID

var IDFrom = id.UserIDFrom
var WorkspaceIDFrom = id.WorkspaceIDFrom

var IDFromRef = id.UserIDFromRef
var WorkspaceIDFromRef = id.WorkspaceIDFromRef

var ErrInvalidID = id.ErrInvalidID
