package usecasex

import (
	"errors"

	"github.com/reearth/reearthx/account/accountdomain"
)

var ErrPolicyVioration = errors.New("policy violation")

type Operator struct {
	Workspaces []OperatableWorkspace
}

type OperatableWorkspace struct {
	Workspace accountdomain.WorkspaceID
	Role      accountdomain.Role
}
