package workspace

import (
	"github.com/reearth/reearthx/account/accountdomain/id"
)

type ID = id.WorkspaceID
type UserID = id.UserID
type IntegrationID = id.IntegrationID

var NewID = id.NewWorkspaceID
var NewUserID = id.NewUserID
var NewIntegrationID = id.NewIntegrationID

var IDFrom = id.WorkspaceIDFrom

var IDFromRef = id.WorkspaceIDFromRef

var ErrInvalidID = id.ErrInvalidID

type PolicyID string

func (id PolicyID) Ref() *PolicyID {
	return &id
}

func (id PolicyID) String() string {
	return string(id)
}
