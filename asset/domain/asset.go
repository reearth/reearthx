package domain

import (
	"time"

	"github.com/reearth/reearthx/id"
)

type ID = id.AssetID
type GroupID = id.GroupID
type ProjectID = id.ProjectID
type WorkspaceID = id.WorkspaceID

var (
	NewID          = id.NewAssetID
	NewGroupID     = id.NewGroupID
	NewProjectID   = id.NewProjectID
	NewWorkspaceID = id.NewWorkspaceID

	MustID          = id.MustAssetID
	MustGroupID     = id.MustGroupID
	MustProjectID   = id.MustProjectID
	MustWorkspaceID = id.MustWorkspaceID

	IDFrom          = id.AssetIDFrom
	GroupIDFrom     = id.GroupIDFrom
	ProjectIDFrom   = id.ProjectIDFrom
	WorkspaceIDFrom = id.WorkspaceIDFrom

	IDFromRef          = id.AssetIDFromRef
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

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusActive     Status = "ACTIVE"
	StatusExtracting Status = "EXTRACTING"
	StatusError      Status = "ERROR"
)

type Asset struct {
	id          ID
	groupID     GroupID
	projectID   ProjectID
	workspaceID WorkspaceID
	name        string
	size        int64
	url         string
	contentType string
	status      Status
	error       string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewAsset(id ID, name string, size int64, contentType string) *Asset {
	now := time.Now()
	return &Asset{
		id:          id,
		name:        name,
		size:        size,
		contentType: contentType,
		status:      StatusPending,
		createdAt:   now,
		updatedAt:   now,
	}
}

// Getters
func (a *Asset) ID() ID                   { return a.id }
func (a *Asset) GroupID() GroupID         { return a.groupID }
func (a *Asset) ProjectID() ProjectID     { return a.projectID }
func (a *Asset) WorkspaceID() WorkspaceID { return a.workspaceID }
func (a *Asset) Name() string             { return a.name }
func (a *Asset) Size() int64              { return a.size }
func (a *Asset) URL() string              { return a.url }
func (a *Asset) ContentType() string      { return a.contentType }
func (a *Asset) Status() Status           { return a.status }
func (a *Asset) Error() string            { return a.error }
func (a *Asset) CreatedAt() time.Time     { return a.createdAt }
func (a *Asset) UpdatedAt() time.Time     { return a.updatedAt }

func (a *Asset) UpdateStatus(status Status, err string) {
	a.status = status
	a.error = err
	a.updatedAt = time.Now()
}

func (a *Asset) UpdateMetadata(name, url, contentType string) {
	if name != "" {
		a.name = name
	}
	if url != "" {
		a.url = url
	}
	if contentType != "" {
		a.contentType = contentType
	}
	a.updatedAt = time.Now()
}

func (a *Asset) MoveToWorkspace(workspaceID WorkspaceID) {
	a.workspaceID = workspaceID
	a.updatedAt = time.Now()
}

func (a *Asset) MoveToProject(projectID ProjectID) {
	a.projectID = projectID
	a.updatedAt = time.Now()
}
