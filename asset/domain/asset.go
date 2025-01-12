package domain

import (
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
)

type Asset struct {
	id          id.ID
	groupID     id.GroupID
	projectID   id.ProjectID
	workspaceID id.WorkspaceID
	name        string
	size        int64
	url         string
	contentType string
	status      Status
	error       string
	createdAt   time.Time
	updatedAt   time.Time
}

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusActive     Status = "ACTIVE"
	StatusExtracting Status = "EXTRACTING"
	StatusError      Status = "ERROR"
)

func NewAsset(id id.ID, name string, size int64, contentType string) *Asset {
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

// ID Getters
func (a *Asset) ID() id.ID                   { return a.id }
func (a *Asset) GroupID() id.GroupID         { return a.groupID }
func (a *Asset) ProjectID() id.ProjectID     { return a.projectID }
func (a *Asset) WorkspaceID() id.WorkspaceID { return a.workspaceID }
func (a *Asset) Name() string                { return a.name }
func (a *Asset) Size() int64                 { return a.size }
func (a *Asset) URL() string                 { return a.url }
func (a *Asset) ContentType() string         { return a.contentType }
func (a *Asset) Status() Status              { return a.status }
func (a *Asset) Error() string               { return a.error }
func (a *Asset) CreatedAt() time.Time        { return a.createdAt }
func (a *Asset) UpdatedAt() time.Time        { return a.updatedAt }

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

func (a *Asset) MoveToWorkspace(workspaceID id.WorkspaceID) {
	a.workspaceID = workspaceID
	a.updatedAt = time.Now()
}

func (a *Asset) MoveToProject(projectID id.ProjectID) {
	a.projectID = projectID
	a.updatedAt = time.Now()
}

func (a *Asset) SetSize(size int64) {
	a.size = size
	a.updatedAt = time.Now()
}
