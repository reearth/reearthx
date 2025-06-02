package workspacesettings

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID         = accountdomain.WorkspaceID
	ResourceID = id.ResourceID
)

var (
	NewID         = accountdomain.NewWorkspaceID
	NewResourceID = id.NewResourceID
)

var (
	MustID         = accountdomain.MustWorkspaceID
	MustResourceID = id.MustResourceID
)

var (
	IDFrom         = accountdomain.WorkspaceIDFrom
	ResourceIDFrom = id.ResourceIDFrom
)

var (
	IDFromRef         = accountdomain.WorkspaceIDFromRef
	ResourceIDFromRef = id.ResourceIDFromRef
)

var ErrInvalidID = id.ErrInvalidID
