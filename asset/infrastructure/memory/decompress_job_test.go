package memory

import (
	"context"
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	assetusecase "github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
)

func TestDecompressJobRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewDecompressJobRepository()

	t.Run("Save and Get", func(t *testing.T) {
		status := &assetusecase.DecompressStatus{
			JobID:     "job1",
			AssetID:   id.NewID(),
			Status:    "pending",
			Progress:  0,
			StartedAt: time.Now(),
		}

		// Test Save
		err := repo.Save(ctx, status)
		assert.NoError(t, err)

		// Test Get
		got, err := repo.Get(ctx, "job1")
		assert.NoError(t, err)
		assert.Equal(t, status, got)

		// Test Get non-existent
		_, err = repo.Get(ctx, "non-existent")
		assert.Equal(t, rerror.ErrNotFound, err)

		// Test Save nil
		err = repo.Save(ctx, nil)
		assert.Equal(t, rerror.ErrInvalidParams, err)
	})

	t.Run("List", func(t *testing.T) {
		repo := NewDecompressJobRepository()

		status1 := &assetusecase.DecompressStatus{
			JobID:     "job1",
			Status:    "processing",
			Progress:  50,
			StartedAt: time.Now(),
		}
		status2 := &assetusecase.DecompressStatus{
			JobID:     "job2",
			Status:    "completed",
			Progress:  100,
			StartedAt: time.Now(),
		}
		status3 := &assetusecase.DecompressStatus{
			JobID:     "job3",
			Status:    "pending",
			Progress:  0,
			StartedAt: time.Now(),
		}

		repo.Save(ctx, status1)
		repo.Save(ctx, status2)
		repo.Save(ctx, status3)

		// Should only return active jobs (not completed or failed)
		jobs, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, jobs, 2)
	})

	t.Run("Delete", func(t *testing.T) {
		repo := NewDecompressJobRepository()
		status := &assetusecase.DecompressStatus{
			JobID:     "job1",
			Status:    "pending",
			StartedAt: time.Now(),
		}

		repo.Save(ctx, status)

		// Test Delete
		err := repo.Delete(ctx, "job1")
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.Get(ctx, "job1")
		assert.Equal(t, rerror.ErrNotFound, err)

		// Test Delete non-existent
		err = repo.Delete(ctx, "non-existent")
		assert.Equal(t, rerror.ErrNotFound, err)
	})

	t.Run("UpdateProgress", func(t *testing.T) {
		repo := NewDecompressJobRepository()
		status := &assetusecase.DecompressStatus{
			JobID:     "job1",
			Status:    "processing",
			Progress:  0,
			StartedAt: time.Now(),
		}

		repo.Save(ctx, status)

		// Test UpdateProgress
		err := repo.UpdateProgress(ctx, "job1", 50.0)
		assert.NoError(t, err)

		// Verify progress update
		got, _ := repo.Get(ctx, "job1")
		assert.Equal(t, 50.0, got.Progress)

		// Test UpdateProgress non-existent
		err = repo.UpdateProgress(ctx, "non-existent", 50.0)
		assert.Equal(t, rerror.ErrNotFound, err)
	})

	t.Run("Complete and Fail", func(t *testing.T) {
		repo := NewDecompressJobRepository()
		status1 := &assetusecase.DecompressStatus{
			JobID:     "job1",
			Status:    "processing",
			StartedAt: time.Now(),
		}
		status2 := &assetusecase.DecompressStatus{
			JobID:     "job2",
			Status:    "processing",
			StartedAt: time.Now(),
		}

		repo.Save(ctx, status1)
		repo.Save(ctx, status2)

		// Test Complete
		err := repo.Complete(ctx, "job1")
		assert.NoError(t, err)
		got, _ := repo.Get(ctx, "job1")
		assert.Equal(t, "completed", got.Status)
		assert.Equal(t, 100.0, got.Progress)

		// Test Fail
		err = repo.Fail(ctx, "job2", "test error")
		assert.NoError(t, err)
		got, _ = repo.Get(ctx, "job2")
		assert.Equal(t, "failed", got.Status)
		assert.Equal(t, "test error", got.Error)

		// Test Complete non-existent
		err = repo.Complete(ctx, "non-existent")
		assert.Equal(t, rerror.ErrNotFound, err)

		// Test Fail non-existent
		err = repo.Fail(ctx, "non-existent", "error")
		assert.Equal(t, rerror.ErrNotFound, err)
	})
}
