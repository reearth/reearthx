package asset

import "context"

// AssetFileRepository handles asset file operations in the storage
type AssetFileRepository interface {
	// Init initializes the repository (creates indexes, etc.)
	Init(ctx context.Context) error

	// FindByID finds a file by asset ID
	FindByID(ctx context.Context, id AssetID) (*File, error)

	// FindByIDs finds multiple files by asset IDs
	FindByIDs(ctx context.Context, ids []AssetID) (map[AssetID]*File, error)

	// Save saves a file for an asset
	Save(ctx context.Context, id AssetID, file *File) error

	// SaveFlat saves files in a flat structure for an asset
	SaveFlat(ctx context.Context, id AssetID, parent *File, files []*File) error

	// Delete removes file data for an asset
	Delete(ctx context.Context, id AssetID) error
}
