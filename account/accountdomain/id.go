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

type Role struct{}
type Permittable struct{}

func (Role) Type() string        { return "role" }
func (Permittable) Type() string { return "permittable" }

type RoleID = idx.ID[Role]
type PermittableID = idx.ID[Permittable]

var NewRoleID = idx.New[Role]
var NewPermittableID = idx.New[Permittable]

var MustRoleID = idx.Must[Role]
var MustPermittableID = idx.Must[Permittable]

var RoleIDFrom = idx.From[Role]
var PermittableIDFrom = idx.From[Permittable]

var RoleIDFromRef = idx.FromRef[Role]
var PermittableIDFromRef = idx.FromRef[Permittable]

type RoleIDList = idx.List[Role]
type PermittableIDList = idx.List[Permittable]

var RoleIDListFrom = idx.ListFrom[Role]
var PermittableIDListFrom = idx.ListFrom[Permittable]

type RoleIDSet = idx.Set[Role]
type PermittableIDSet = idx.Set[Permittable]

var NewRoleIDSet = idx.NewSet[Role]
var NewPermittableIDSet = idx.NewSet[Permittable]
