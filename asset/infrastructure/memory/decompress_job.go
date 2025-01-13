package memory

import (
	"context"
	"sync"

	"github.com/reearth/reearthx/asset/repository"
	assetusecase "github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/rerror"
)

var _ repository.DecompressJobRepository = (*DecompressJobRepository)(nil)

// DecompressJobRepository is an in-memory implementation of repository.DecompressJobRepository
type DecompressJobRepository struct {
	mu   sync.RWMutex
	jobs map[string]*assetusecase.DecompressStatus
}

// NewDecompressJobRepository creates a new in-memory decompress job repository
func NewDecompressJobRepository() *DecompressJobRepository {
	return &DecompressJobRepository{
		jobs: make(map[string]*assetusecase.DecompressStatus),
	}
}

// Save saves or updates a decompress job status
func (r *DecompressJobRepository) Save(ctx context.Context, status *assetusecase.DecompressStatus) error {
	if status == nil {
		return rerror.ErrInvalidParams
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.jobs[status.JobID] = status
	return nil
}

// Get retrieves a decompress job status by ID
func (r *DecompressJobRepository) Get(ctx context.Context, jobID string) (*assetusecase.DecompressStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if status, ok := r.jobs[jobID]; ok {
		return status, nil
	}
	return nil, rerror.ErrNotFound
}

// List retrieves all active decompress jobs
func (r *DecompressJobRepository) List(ctx context.Context) ([]*assetusecase.DecompressStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	jobs := make([]*assetusecase.DecompressStatus, 0, len(r.jobs))
	for _, status := range r.jobs {
		if status.Status != "completed" && status.Status != "failed" {
			jobs = append(jobs, status)
		}
	}
	return jobs, nil
}

// Delete removes a decompress job status
func (r *DecompressJobRepository) Delete(ctx context.Context, jobID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.jobs[jobID]; !ok {
		return rerror.ErrNotFound
	}

	delete(r.jobs, jobID)
	return nil
}

// UpdateProgress updates the progress of a decompress job
func (r *DecompressJobRepository) UpdateProgress(ctx context.Context, jobID string, progress float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	status, ok := r.jobs[jobID]
	if !ok {
		return rerror.ErrNotFound
	}

	status.Progress = progress
	return nil
}

// Complete marks a decompress job as completed
func (r *DecompressJobRepository) Complete(ctx context.Context, jobID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	status, ok := r.jobs[jobID]
	if !ok {
		return rerror.ErrNotFound
	}

	status.Status = "completed"
	status.Progress = 100
	return nil
}

// Fail marks a decompress job as failed with an error message
func (r *DecompressJobRepository) Fail(ctx context.Context, jobID string, err string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	status, ok := r.jobs[jobID]
	if !ok {
		return rerror.ErrNotFound
	}

	status.Status = "failed"
	status.Error = err
	return nil
}
