package asset

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
	ID          ID
	GroupID     ID
	Name        string
	Size        int64
	URL         string
	ContentType string
	Status      Status
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateAssetInput struct {
	Name        string
	Size        int64
	ContentType string
}

type UpdateAssetInput struct {
	Name        *string
	URL         *string
	ContentType *string
	Status      Status
	Error       string
}
