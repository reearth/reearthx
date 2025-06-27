package request

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID          = id.RequestID
	WorkspaceID = id.WorkspaceID
	ProjectID   = id.ProjectID
	ItemID      = id.ItemID
	UserID      = accountdomain.UserID
	UserIDList  = accountdomain.UserIDList
	ThreadID    = id.ThreadID
)

var (
	NewID          = id.NewRequestID
	NewWorkspaceID = accountdomain.NewWorkspaceID
	NewProjectID   = id.NewProjectID
	NewThreadID    = id.NewThreadID
	NewUserID      = accountdomain.NewUserID
	NewItemID      = id.NewItemID
	MustID         = id.MustRequestID
	IDFrom         = id.RequestIDFrom
	IDFromRef      = id.RequestIDFromRef
)

var ErrInvalidID = id.ErrInvalidID
