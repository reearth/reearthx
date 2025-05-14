package id

import "github.com/reearth/reearthx/idx"

type idAsset struct{}
type idGroup struct{}
type idProject struct{}
type idWorkspace struct{}

func (idAsset) Type() string     { return "asset" }
func (idGroup) Type() string     { return "group" }
func (idProject) Type() string   { return "project" }
func (idWorkspace) Type() string { return "workspace" }

type ID = idx.ID[idAsset]
type GroupID = idx.ID[idGroup]
type ProjectID = idx.ID[idProject]
type WorkspaceID = idx.ID[idWorkspace]

var (
	NewID          = idx.New[idAsset]
	NewGroupID     = idx.New[idGroup]
	NewProjectID   = idx.New[idProject]
	NewWorkspaceID = idx.New[idWorkspace]

	MustID          = idx.Must[idAsset]
	MustGroupID     = idx.Must[idGroup]
	MustProjectID   = idx.Must[idProject]
	MustWorkspaceID = idx.Must[idWorkspace]

	From            = idx.From[idAsset]
	GroupIDFrom     = idx.From[idGroup]
	ProjectIDFrom   = idx.From[idProject]
	WorkspaceIDFrom = idx.From[idWorkspace]

	FromRef            = idx.FromRef[idAsset]
	GroupIDFromRef     = idx.FromRef[idGroup]
	ProjectIDFromRef   = idx.FromRef[idProject]
	WorkspaceIDFromRef = idx.FromRef[idWorkspace]

	ErrInvalidID = idx.ErrInvalidID
)
