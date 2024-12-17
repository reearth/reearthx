// asset/asset.go
package asset

import (
	"errors"
	"time"
)

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrEmptyWorkspaceID = errors.New("workspace id is required")
	ErrEmptyURL         = errors.New("valid url is required")
	ErrEmptySize        = errors.New("file size cannot be zero")
)

// Asset represents a file stored in the system
type Asset struct {
	id          ID
	workspaceID WorkspaceID
	name        string
	size        int64
	url         string
	contentType string
	status      Status // For tracking extraction status
	metadata    Metadata
	createdAt   time.Time
}

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
)

type Metadata struct {
	IsArchive      bool            `json:"isArchive,omitempty"`
	ExtractedFiles []ExtractedFile `json:"extractedFiles,omitempty"`
}

type ExtractedFile struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

// Repository defines operations for asset persistence
type Repository interface {
	Save(Asset) error
	Find(ID) (*Asset, error)
	FindByWorkspace(WorkspaceID) ([]Asset, error)
	Remove(ID) error
}

// Storage defines operations for file storage
type Storage interface {
	Upload(file []byte, contentType string) (string, error)
	Download(url string) ([]byte, error)
	Delete(url string) error
	GetSignedURL(key string, contentType string) (string, error)
}

// Extractor handles archive extraction
type Extractor interface {
	Extract(Asset) error
	GetStatus(ID) (Status, error)
}
