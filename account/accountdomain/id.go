package accountdomain

import (
	"github.com/reearth/reearthx/account/accountdomain/accountid"
	"github.com/reearth/reearthx/idx"
)

type UserIDType struct{}
type WorkspaceIDType struct{}
type IntegrationIDType struct{}

func (*UserIDType) Type() string        { return "user" }
func (*WorkspaceIDType) Type() string   { return "workspace" }
func (*IntegrationIDType) Type() string { return "integration" }

type RawUserID = idx.ID[*UserIDType]
type RawWorkspaceID = idx.ID[*WorkspaceIDType]
type RawIntegrationID = idx.ID[*IntegrationIDType]

var NewRawUserID = idx.New[*UserIDType]
var MustRawUserID = idx.Must[*UserIDType]
var RawUserIDFrom = idx.From[*UserIDType]
var RawUserIDFromRef = idx.FromRef[*UserIDType]

var NewRawWorkspaceID = idx.New[*WorkspaceIDType]
var MustRawWorkspaceID = idx.Must[*WorkspaceIDType]
var RawWorkspaceIDFrom = idx.From[*WorkspaceIDType]
var RawWorkspaceIDFromRef = idx.FromRef[*WorkspaceIDType]

var NewRawIntegrationID = idx.New[*IntegrationIDType]
var MustRawIntegrationID = idx.Must[*IntegrationIDType]
var RawIntegrationIDFrom = idx.From[*IntegrationIDType]
var RawIntegrationIDFromRef = idx.FromRef[*IntegrationIDType]

type UserID = accountid.ID[*UserIDType]
type WorkspaceID = accountid.ID[*WorkspaceIDType]
type IntegrationID = accountid.ID[*IntegrationIDType]

var NewUserID = accountid.New[*UserIDType]
var UserIDFrom = accountid.Parse[*UserIDType]
var MustUserID = accountid.Must[*UserIDType]
var GenerateUserID = accountid.Generate[*UserIDType]

var NewWorkspaceID = accountid.New[*WorkspaceIDType]
var MustWorkspaceID = accountid.Must[*WorkspaceIDType]
var GenerateWorkspaceID = accountid.Generate[*WorkspaceIDType]
var WorkspaceIDFrom = accountid.Parse[*WorkspaceIDType]

var NewIntegrationID = accountid.New[*IntegrationIDType]
var MustIntegrationID = accountid.Must[*IntegrationIDType]
var GenerateIntegrationID = accountid.Generate[*IntegrationIDType]
var IntegrationIDFrom = accountid.Parse[*IntegrationIDType]

var ErrInvalidID = idx.ErrInvalidID
