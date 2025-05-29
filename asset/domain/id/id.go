package id

import (
	"github.com/reearth/reearthx/idx"
)

type Type struct{}

func (Type) Type() string {
	return "asset"
}

type ID = idx.ID[Type]

type IntegrationIDType struct{}

func (IntegrationIDType) Type() string {
	return "integration"
}

type IntegrationID = idx.ID[IntegrationIDType]

type IntegrationIDList = idx.List[IntegrationIDType]

type ProjectIDType struct{}

func (ProjectIDType) Type() string {
	return "project"
}

type ProjectID = idx.ID[ProjectIDType]

type ProjectIDList = idx.List[ProjectIDType]

type workspaceIDType struct{}

func (workspaceIDType) Type() string {
	return "workspace"
}

type workspaceID = idx.ID[workspaceIDType]
type workspaceIDList = idx.List[workspaceIDType]

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
	return idx.New[Type]()
}

func From(id string) (ID, error) {
	return idx.From[Type](id)
}

func MustAssetID(id string) ID {
	return idx.Must[Type](id)
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

type List []ID

func (l List) Add(id ID) List {
	return append(l, id)
}

func (l List) Strings() []string {
	strings := make([]string, len(l))
	for i, id := range l {
		strings[i] = id.String()
	}
	return strings
}
