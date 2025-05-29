package id

import (
	"github.com/reearth/reearthx/idx"
)

// Asset type and ID
type Type struct{}

func (Type) Type() string {
	return "asset"
}

type ID = idx.ID[Type]
type List []ID

func NewAssetID() ID {
	return idx.New[Type]()
}

func From(id string) (ID, error) {
	return idx.From[Type](id)
}

func MustAssetID(id string) ID {
	return idx.Must[Type](id)
}

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

// Group type and ID
type Group struct{}

func (Group) Type() string {
	return "group"
}

type GroupID = idx.ID[Group]
type GroupIDList = idx.List[Group]

func NewGroupID() GroupID {
	return idx.New[Group]()
}

func GroupIDFrom(id string) (GroupID, error) {
	return idx.From[Group](id)
}

func MustGroupID(id string) GroupID {
	return idx.Must[Group](id)
}

// Integration type and ID
type Integration struct{}

func (Integration) Type() string {
	return "integration"
}

type IntegrationID = idx.ID[Integration]
type IntegrationIDList = idx.List[Integration]

func NewIntegrationID() IntegrationID {
	return idx.New[Integration]()
}

func IntegrationIDFrom(id string) (IntegrationID, error) {
	return idx.From[Integration](id)
}

func MustIntegrationID(id string) IntegrationID {
	return idx.Must[Integration](id)
}

func IntegrationIDFromRef(id *string) *IntegrationID {
	return idx.FromRef[Integration](id)
}

// Model type and ID
type ModelIDType struct{}

func (ModelIDType) Type() string {
	return "model"
}

type ModelID = idx.ID[ModelIDType]

func NewModelID() ModelID {
	return idx.New[ModelIDType]()
}

func ModelIDFrom(id string) (ModelID, error) {
	return idx.From[ModelIDType](id)
}

func MustModelID(id string) ModelID {
	return idx.Must[ModelIDType](id)
}

// Operator type
type Operator struct{}

func (Operator) Type() string {
	return "operator"
}

// Policy type and ID
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

// Project type and ID
type Project struct{}

func (Project) Type() string {
	return "project"
}

type ProjectID = idx.ID[Project]
type ProjectIDList = idx.List[Project]

func NewProjectID() ProjectID {
	return idx.New[Project]()
}

// Thread type and ID
type Thread struct{}

func (Thread) Type() string {
	return "thread"
}

type ThreadID = idx.ID[Thread]

// Webhook type and ID
type WebhookIDType struct{}

func (WebhookIDType) Type() string {
	return "webhook"
}

type WebhookID = idx.ID[WebhookIDType]

func NewWebhookID() WebhookID {
	return idx.New[WebhookIDType]()
}

func WebhookIDFrom(id string) (WebhookID, error) {
	return idx.From[WebhookIDType](id)
}

func MustWebhookID(id string) WebhookID {
	return idx.Must[WebhookIDType](id)
}

// Workspace type and ID
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

// Legacy workspace
type workspace struct{}

func (workspace) Type() string {
	return "workspace"
}

type workspaceID = idx.ID[workspace]
type workspaceIDList = idx.List[workspace]

// Error
var ErrInvalidID = idx.ErrInvalidID
