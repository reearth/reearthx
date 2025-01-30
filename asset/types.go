package asset

import (
	"time"
)

type ID string

type Asset struct {
	ID          ID
	Name        string
	Size        int64
	URL         string
	ContentType string
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
}
