package asset

import (
	"context"
	"io"
)

type Repository interface {
	// Create creates a new asset
	Create(ctx context.Context, asset *Asset) error
	// Read returns an asset by ID
	Read(ctx context.Context, id ID) (*Asset, error)
	// Update updates an existing asset
	Update(ctx context.Context, asset *Asset) error
	// Delete removes an asset by ID
	Delete(ctx context.Context, id ID) error
	// List returns all assets
	List(ctx context.Context) ([]*Asset, error)

	// Existing file operations
	FetchFile(ctx context.Context, id ID) (io.ReadCloser, error)
	Upload(ctx context.Context, id ID, file io.Reader) error
	GetUploadURL(ctx context.Context, id ID) (string, error)
}
