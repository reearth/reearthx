package workspacesettings

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type ID = accountdomain.WorkspaceID
type ResourceID = id.ResourceID

var NewID = accountdomain.NewWorkspaceID
var NewResourceID = id.NewResourceID

var MustID = accountdomain.MustWorkspaceID
var MustResourceID = id.MustResourceID

var IDFrom = accountdomain.WorkspaceIDFrom
var ResourceIDFrom = id.ResourceIDFrom

var IDFromRef = accountdomain.WorkspaceIDFromRef
var ResourceIDFromRef = id.ResourceIDFromRef

var ErrInvalidID = id.ErrInvalidID
