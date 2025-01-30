package repository

import (
	"context"

	assetusecase "github.com/reearth/reearthx/asset/usecase"
)

// DecompressJobRepository defines the interface for storing decompression job status
type DecompressJobRepository interface {
	// Save saves or updates a decompress job status
	Save(ctx context.Context, status *assetusecase.DecompressStatus) error

	// Get retrieves a decompress job status by ID
	Get(ctx context.Context, jobID string) (*assetusecase.DecompressStatus, error)

	// List retrieves all active decompress jobs
	List(ctx context.Context) ([]*assetusecase.DecompressStatus, error)

	// Delete removes a decompress job status
	Delete(ctx context.Context, jobID string) error

	// UpdateProgress updates the progress of a decompress job
	UpdateProgress(ctx context.Context, jobID string, progress float64) error

	// Complete marks a decompress job as completed
	Complete(ctx context.Context, jobID string) error

	// Fail marks a decompress job as failed with an error message
	Fail(ctx context.Context, jobID string, err string) error
}
