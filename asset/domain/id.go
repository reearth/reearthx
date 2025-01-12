package domain

import (
	"github.com/reearth/reearthx/asset/domain/id"
)

type ID = id.ID
type GroupID = id.GroupID
type ProjectID = id.ProjectID
type WorkspaceID = id.WorkspaceID

var (
	NewID          = id.NewID
	NewGroupID     = id.NewGroupID
	NewProjectID   = id.NewProjectID
	NewWorkspaceID = id.NewWorkspaceID

	MustID          = id.MustID
	MustGroupID     = id.MustGroupID
	MustProjectID   = id.MustProjectID
	MustWorkspaceID = id.MustWorkspaceID

	IDFrom          = id.IDFrom
	GroupIDFrom     = id.GroupIDFrom
	ProjectIDFrom   = id.ProjectIDFrom
	WorkspaceIDFrom = id.WorkspaceIDFrom

	IDFromRef          = id.IDFromRef
	GroupIDFromRef     = id.GroupIDFromRef
	ProjectIDFromRef   = id.ProjectIDFromRef
	WorkspaceIDFromRef = id.WorkspaceIDFromRef

	ErrInvalidID = id.ErrInvalidID
)

func MockNewID(i ID) func() {
	original := NewID
	NewID = func() ID { return i }
	return func() { NewID = original }
}

func MockNewGroupID(i GroupID) func() {
	original := NewGroupID
	NewGroupID = func() GroupID { return i }
	return func() { NewGroupID = original }
}

func MockNewProjectID(i ProjectID) func() {
	original := NewProjectID
	NewProjectID = func() ProjectID { return i }
	return func() { NewProjectID = original }
}

func MockNewWorkspaceID(i WorkspaceID) func() {
	original := NewWorkspaceID
	NewWorkspaceID = func() WorkspaceID { return i }
	return func() { NewWorkspaceID = original }
}
