package entity_test

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/stretchr/testify/assert"
)

func TestNewAsset(t *testing.T) {
	assetID := id.NewID()
	name := "test.jpg"
	size := int64(1024)
	contentType := "image/jpeg"

	asset := entity.NewAsset(assetID, name, size, contentType)

	assert.Equal(t, assetID, asset.ID())
	assert.Equal(t, name, asset.Name())
	assert.Equal(t, size, asset.Size())
	assert.Equal(t, contentType, asset.ContentType())
	assert.Equal(t, entity.StatusPending, asset.Status())
	assert.Empty(t, asset.Error())
	assert.NotZero(t, asset.CreatedAt())
	assert.NotZero(t, asset.UpdatedAt())
	assert.Equal(t, asset.CreatedAt(), asset.UpdatedAt())
}

func TestAsset_UpdateStatus(t *testing.T) {
	asset := entity.NewAsset(id.NewID(), "test.jpg", 1024, "image/jpeg")
	initialUpdatedAt := asset.UpdatedAt()
	time.Sleep(time.Millisecond) // Ensure time difference

	asset.UpdateStatus(entity.StatusError, "test error")

	assert.Equal(t, entity.StatusError, asset.Status())
	assert.Equal(t, "test error", asset.Error())
	assert.True(t, asset.UpdatedAt().After(initialUpdatedAt))
}

func TestAsset_UpdateMetadata(t *testing.T) {
	asset := entity.NewAsset(id.NewID(), "test.jpg", 1024, "image/jpeg")
	initialUpdatedAt := asset.UpdatedAt()
	time.Sleep(time.Millisecond)

	newName := "new.jpg"
	newURL := "https://example.com/new.jpg"
	newContentType := "image/png"

	asset.UpdateMetadata(newName, newURL, newContentType)

	assert.Equal(t, newName, asset.Name())
	assert.Equal(t, newURL, asset.URL())
	assert.Equal(t, newContentType, asset.ContentType())
	assert.True(t, asset.UpdatedAt().After(initialUpdatedAt))

	// Test partial update
	asset.UpdateMetadata("", "new-url", "")
	assert.Equal(t, newName, asset.Name())
	assert.Equal(t, "new-url", asset.URL())
	assert.Equal(t, newContentType, asset.ContentType())
}

func TestAsset_MoveToWorkspace(t *testing.T) {
	asset := entity.NewAsset(id.NewID(), "test.jpg", 1024, "image/jpeg")
	initialUpdatedAt := asset.UpdatedAt()
	time.Sleep(time.Millisecond)

	workspaceID := id.NewWorkspaceID()
	asset.MoveToWorkspace(workspaceID)

	assert.Equal(t, workspaceID, asset.WorkspaceID())
	assert.True(t, asset.UpdatedAt().After(initialUpdatedAt))
}

func TestAsset_MoveToProject(t *testing.T) {
	asset := entity.NewAsset(id.NewID(), "test.jpg", 1024, "image/jpeg")
	initialUpdatedAt := asset.UpdatedAt()
	time.Sleep(time.Millisecond)

	projectID := id.NewProjectID()
	asset.MoveToProject(projectID)

	assert.Equal(t, projectID, asset.ProjectID())
	assert.True(t, asset.UpdatedAt().After(initialUpdatedAt))
}

func TestAsset_MoveToGroup(t *testing.T) {
	asset := entity.NewAsset(id.NewID(), "test.jpg", 1024, "image/jpeg")
	initialUpdatedAt := asset.UpdatedAt()
	time.Sleep(time.Millisecond)

	groupID := id.NewGroupID()
	asset.MoveToGroup(groupID)

	assert.Equal(t, groupID, asset.GroupID())
	assert.True(t, asset.UpdatedAt().After(initialUpdatedAt))
}

func TestAsset_SetSize(t *testing.T) {
	asset := entity.NewAsset(id.NewID(), "test.jpg", 1024, "image/jpeg")
	initialUpdatedAt := asset.UpdatedAt()
	time.Sleep(time.Millisecond)

	newSize := int64(2048)
	asset.SetSize(newSize)

	assert.Equal(t, newSize, asset.Size())
	assert.True(t, asset.UpdatedAt().After(initialUpdatedAt))
}
