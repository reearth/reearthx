package domain

import (
	"time"
)

type ID string

func (id ID) String() string {
	return string(id)
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
	groupID     ID
	projectID   ID
	workspaceID ID
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
func (a *Asset) ID() ID               { return a.id }
func (a *Asset) GroupID() ID          { return a.groupID }
func (a *Asset) ProjectID() ID        { return a.projectID }
func (a *Asset) WorkspaceID() ID      { return a.workspaceID }
func (a *Asset) Name() string         { return a.name }
func (a *Asset) Size() int64          { return a.size }
func (a *Asset) URL() string          { return a.url }
func (a *Asset) ContentType() string  { return a.contentType }
func (a *Asset) Status() Status       { return a.status }
func (a *Asset) Error() string        { return a.error }
func (a *Asset) CreatedAt() time.Time { return a.createdAt }
func (a *Asset) UpdatedAt() time.Time { return a.updatedAt }

// Methods
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
