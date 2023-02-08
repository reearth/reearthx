package workspace

import "github.com/reearth/reearthx/account/accountdomain"

type ID = accountdomain.WorkspaceID
type UserID = accountdomain.UserID
type IntegrationID = accountdomain.IntegrationID

var NewID = accountdomain.NewWorkspaceID
var NewUserID = accountdomain.NewUserID
var NewIntegrationID = accountdomain.NewIntegrationID

var IDFrom = accountdomain.WorkspaceIDFrom

var IDFromRef = accountdomain.WorkspaceIDFromRef

var ErrInvalidID = accountdomain.ErrInvalidID

type PolicyID string

func (id PolicyID) Ref() *PolicyID {
	return &id
}

func (id PolicyID) String() string {
	return string(id)
}
