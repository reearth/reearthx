package usecasex

import (
	"errors"

	"github.com/reearth/reearthx/account/accountdomain/workspace"
)

var ErrPolicyVioration = errors.New("policy violation")

type Operator struct {
	Workspaces []OperatableWorkspace
}

type OperatableWorkspace struct {
	Workspace workspace.ID
	Role      workspace.Role
}
