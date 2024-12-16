package asset

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrEmptyWorkspaceID = errors.New("require workspace id")
	ErrEmptyURL         = errors.New("require valid url")
	ErrEmptySize        = errors.New("file size cannot be zero")
	ErrInvalidName      = errors.New("invalid file name")
	ErrInvalidID        = errors.New("invalid asset id")
)

// Asset represents a file resource in the system
type Asset struct {
	id          ID
	createdAt   time.Time
	workspace   WorkspaceID
	name        string // file name
	size        int64  // file size
	url         string
	contentType string
	coreSupport bool
}

// New creates a new Asset
func New(workspace WorkspaceID, name string, size int64, url, contentType string) (*Asset, error) {
	if err := validateAssetInput(workspace, name, size, url); err != nil {
		return nil, err
	}

	return &Asset{
		id:          NewID(),
		createdAt:   time.Now(),
		workspace:   workspace,
		name:        name,
		size:        size,
		url:         url,
		contentType: contentType,
		coreSupport: false,
	}, nil
}

// Validate input parameters for new asset
func validateAssetInput(workspace WorkspaceID, name string, size int64, url string) error {
	if workspace.IsEmpty() {
		return ErrEmptyWorkspaceID
	}
	if name == "" {
		return ErrInvalidName
	}
	if size <= 0 {
		return ErrEmptySize
	}
	if url == "" {
		return ErrEmptyURL
	}
	return nil
}

// Getters
func (a *Asset) ID() ID {
	return a.id
}

func (a *Asset) Workspace() WorkspaceID {
	return a.workspace
}

func (a *Asset) Name() string {
	return a.name
}

func (a *Asset) Size() int64 {
	return a.size
}

func (a *Asset) URL() string {
	return a.url
}

func (a *Asset) ContentType() string {
	return a.contentType
}

func (a *Asset) CoreSupport() bool {
	return a.coreSupport
}

func (a *Asset) CreatedAt() time.Time {
	if a == nil {
		return time.Time{}
	}
	return a.createdAt
}

// Setters
func (a *Asset) SetCoreSupport(support bool) {
	if a == nil {
		return
	}
	a.coreSupport = support
}

func (a *Asset) SetContentType(contentType string) {
	if a == nil {
		return
	}
	a.contentType = contentType
}

// Clone returns a deep copy of the Asset
func (a *Asset) Clone() *Asset {
	if a == nil {
		return nil
	}
	return &Asset{
		id:          a.id,
		createdAt:   a.createdAt,
		workspace:   a.workspace,
		name:        a.name,
		size:        a.size,
		url:         a.url,
		contentType: a.contentType,
		coreSupport: a.coreSupport,
	}
}

// Equals checks if two assets are equal
func (a *Asset) Equals(other *Asset) bool {
	if a == nil && other == nil {
		return true
	}
	if a == nil || other == nil {
		return false
	}
	return a.id == other.id &&
		a.workspace == other.workspace &&
		a.name == other.name &&
		a.size == other.size &&
		a.url == other.url &&
		a.contentType == other.contentType &&
		a.coreSupport == other.coreSupport &&
		a.createdAt.Equal(other.createdAt)
}
