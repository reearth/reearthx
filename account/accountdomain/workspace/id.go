package workspace

import "github.com/reearth/reearthx/account/accountdomain"

type ID = accountdomain.WorkspaceID
type IDList = accountdomain.WorkspaceIDList
type UserID = accountdomain.UserID
type UserIDList = accountdomain.UserIDList
type IntegrationID = accountdomain.IntegrationID
type IntegrationIDList = accountdomain.IntegrationIDList

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
