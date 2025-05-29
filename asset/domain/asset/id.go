package asset

import (
	"github.com/reearth/reearthx/idx"
)

type IDType struct{}

func (IDType) Type() string {
	return "asset"
}

type ID = idx.ID[IDType]

type IntegrationIDType struct{}

func (IntegrationIDType) Type() string {
	return "integration"
}

type IntegrationID = idx.ID[IntegrationIDType]

type OperatorIDType struct{}

func (OperatorIDType) Type() string {
	return "operator"
}

type ThreadIDType struct{}

func (ThreadIDType) Type() string {
	return "thread"
}

type ThreadID = idx.ID[ThreadIDType]

func NewAssetID() ID {
	return idx.New[IDType]()
}

func IdFrom(id string) (ID, error) {
	return idx.From[IDType](id)
}

func MustAssetID(id string) ID {
	return idx.Must[IDType](id)
}

type GroupIDType struct{}

func (GroupIDType) Type() string {
	return "group"
}

type GroupID = idx.ID[GroupIDType]

type GroupIDList = idx.List[GroupIDType]

func NewGroupID() GroupID {
	return idx.New[GroupIDType]()
}

func GroupIDFrom(id string) (GroupID, error) {
	return idx.From[GroupIDType](id)
}

func MustGroupID(id string) GroupID {
	return idx.Must[GroupIDType](id)
}

type WorkspaceIDType struct{}

func (WorkspaceIDType) Type() string {
	return "workspace"
}

type WorkspaceID = idx.ID[WorkspaceIDType]

type WorkspaceIDList = idx.List[WorkspaceIDType]

func NewWorkspaceID() WorkspaceID {
	return idx.New[WorkspaceIDType]()
}

func WorkspaceIDFrom(id string) (WorkspaceID, error) {
	return idx.From[WorkspaceIDType](id)
}

func MustWorkspaceID(id string) WorkspaceID {
	return idx.Must[WorkspaceIDType](id)
}

type PolicyIDType struct{}

func (PolicyIDType) Type() string {
	return "policy"
}

type PolicyID = idx.ID[PolicyIDType]

func NewPolicyID() PolicyID {
	return idx.New[PolicyIDType]()
}

func PolicyIDFrom(id string) (PolicyID, error) {
	return idx.From[PolicyIDType](id)
}

func MustPolicyID(id string) PolicyID {
	return idx.Must[PolicyIDType](id)
}
