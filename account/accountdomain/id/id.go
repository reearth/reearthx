package id

import (
	"github.com/reearth/reearthx/idx"
)

type UserIDType struct{}
type WorkspaceIDType struct{}
type IntegrationIDType struct{}

func (*UserIDType) Type() string        { return "user" }
func (*WorkspaceIDType) Type() string   { return "workspace" }
func (*IntegrationIDType) Type() string { return "integration" }

type UserID = idx.ID[*UserIDType]
type WorkspaceID = idx.ID[*WorkspaceIDType]
type IntegrationID = idx.ID[*IntegrationIDType]

var NewUserID = idx.New[*UserIDType]
var MustUserID = idx.Must[*UserIDType]
var UserIDFrom = idx.From[*UserIDType]
var UserIDFromRef = idx.FromRef[*UserIDType]

var NewWorkspaceID = idx.New[*WorkspaceIDType]
var MustWorkspaceID = idx.Must[*WorkspaceIDType]
var WorkspaceIDFrom = idx.From[*WorkspaceIDType]
var WorkspaceIDFromRef = idx.FromRef[*WorkspaceIDType]

var NewIntegrationID = idx.New[*IntegrationIDType]
var MustIntegrationID = idx.Must[*IntegrationIDType]
var IntegrationIDFrom = idx.From[*IntegrationIDType]
var IntegrationIDFromRef = idx.FromRef[*IntegrationIDType]

var ErrInvalidID = idx.ErrInvalidID
