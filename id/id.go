package id

import "github.com/reearth/reearthx/idx"

type Asset struct{}
type Group struct{}
type Project struct{}
type Workspace struct{}

func (Asset) Type() string     { return "asset" }
func (Group) Type() string     { return "group" }
func (Project) Type() string   { return "project" }
func (Workspace) Type() string { return "workspace" }

type AssetID = idx.ID[Asset]
type GroupID = idx.ID[Group]
type ProjectID = idx.ID[Project]
type WorkspaceID = idx.ID[Workspace]

var NewAssetID = idx.New[Asset]
var NewGroupID = idx.New[Group]
var NewProjectID = idx.New[Project]
var NewWorkspaceID = idx.New[Workspace]

var MustAssetID = idx.Must[Asset]
var MustGroupID = idx.Must[Group]
var MustProjectID = idx.Must[Project]
var MustWorkspaceID = idx.Must[Workspace]

var AssetIDFrom = idx.From[Asset]
var GroupIDFrom = idx.From[Group]
var ProjectIDFrom = idx.From[Project]
var WorkspaceIDFrom = idx.From[Workspace]

var AssetIDFromRef = idx.FromRef[Asset]
var GroupIDFromRef = idx.FromRef[Group]
var ProjectIDFromRef = idx.FromRef[Project]
var WorkspaceIDFromRef = idx.FromRef[Workspace]

type AssetIDList = idx.List[Asset]
type GroupIDList = idx.List[Group]
type ProjectIDList = idx.List[Project]
type WorkspaceIDList = idx.List[Workspace]

var AssetIDListFrom = idx.ListFrom[Asset]
var GroupIDListFrom = idx.ListFrom[Group]
var ProjectIDListFrom = idx.ListFrom[Project]
var WorkspaceIDListFrom = idx.ListFrom[Workspace]

type AssetIDSet = idx.Set[Asset]
type GroupIDSet = idx.Set[Group]
type ProjectIDSet = idx.Set[Project]
type WorkspaceIDSet = idx.Set[Workspace]

var NewAssetIDSet = idx.NewSet[Asset]
var NewGroupIDSet = idx.NewSet[Group]
var NewProjectIDSet = idx.NewSet[Project]
var NewWorkspaceIDSet = idx.NewSet[Workspace]

var ErrInvalidID = idx.ErrInvalidID
