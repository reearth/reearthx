package accountdomain

import (
	"github.com/reearth/reearthx/idx"
)

type User struct{}
type Workspace struct{}
type Integration struct{}

func (User) Type() string        { return "user" }
func (Workspace) Type() string   { return "workspace" }
func (Integration) Type() string { return "integration" }

type UserID = idx.ID[User]
type WorkspaceID = idx.ID[Workspace]
type IntegrationID = idx.ID[Integration]

type UserIDList = idx.List[User]
type WorkspaceIDList = idx.List[Workspace]
type IntegrationIDList = idx.List[Integration]

var NewUserID = idx.New[User]
var MustUserID = idx.Must[User]
var UserIDFrom = idx.From[User]
var UserIDFromRef = idx.FromRef[User]

var NewWorkspaceID = idx.New[Workspace]
var MustWorkspaceID = idx.Must[Workspace]
var WorkspaceIDFrom = idx.From[Workspace]
var WorkspaceIDFromRef = idx.FromRef[Workspace]

var NewIntegrationID = idx.New[Integration]
var MustIntegrationID = idx.Must[Integration]
var IntegrationIDFrom = idx.From[Integration]
var IntegrationIDFromRef = idx.FromRef[Integration]

var ErrInvalidID = idx.ErrInvalidID
