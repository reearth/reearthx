package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAsset(t *testing.T) {
	id := NewID()
	a := NewAsset(id, "test.txt", 100, "text/plain")

	assert.Equal(t, id, a.ID())
	assert.Equal(t, "test.txt", a.Name())
	assert.Equal(t, int64(100), a.Size())
	assert.Equal(t, "text/plain", a.ContentType())
	assert.Equal(t, StatusPending, a.Status())
	assert.Empty(t, a.Error())
	assert.NotZero(t, a.CreatedAt())
	assert.NotZero(t, a.UpdatedAt())
	assert.Equal(t, a.CreatedAt(), a.UpdatedAt())
}

func TestAsset_UpdateStatus(t *testing.T) {
	a := NewAsset(NewID(), "test.txt", 100, "text/plain")
	createdAt := a.CreatedAt()
	time.Sleep(time.Millisecond)

	a.UpdateStatus(StatusError, "test error")
	assert.Equal(t, StatusError, a.Status())
	assert.Equal(t, "test error", a.Error())
	assert.Equal(t, createdAt, a.CreatedAt())
	assert.True(t, a.UpdatedAt().After(createdAt))
}

func TestAsset_UpdateMetadata(t *testing.T) {
	a := NewAsset(NewID(), "test.txt", 100, "text/plain")
	createdAt := a.CreatedAt()
	time.Sleep(time.Millisecond)

	a.UpdateMetadata("new.txt", "http://example.com", "application/json")
	assert.Equal(t, "new.txt", a.Name())
	assert.Equal(t, "http://example.com", a.URL())
	assert.Equal(t, "application/json", a.ContentType())
	assert.Equal(t, createdAt, a.CreatedAt())
	assert.True(t, a.UpdatedAt().After(createdAt))

	// Test partial update
	updatedAt := a.UpdatedAt()
	time.Sleep(time.Millisecond)
	a.UpdateMetadata("", "new-url", "")
	assert.Equal(t, "new.txt", a.Name())
	assert.Equal(t, "new-url", a.URL())
	assert.Equal(t, "application/json", a.ContentType())
	assert.True(t, a.UpdatedAt().After(updatedAt))
}

func TestAsset_MoveToWorkspace(t *testing.T) {
	a := NewAsset(NewID(), "test.txt", 100, "text/plain")
	createdAt := a.CreatedAt()
	time.Sleep(time.Millisecond)

	wsID := NewWorkspaceID()
	a.MoveToWorkspace(wsID)
	assert.Equal(t, wsID, a.WorkspaceID())
	assert.Equal(t, createdAt, a.CreatedAt())
	assert.True(t, a.UpdatedAt().After(createdAt))
}

func TestAsset_MoveToProject(t *testing.T) {
	a := NewAsset(NewID(), "test.txt", 100, "text/plain")
	createdAt := a.CreatedAt()
	time.Sleep(time.Millisecond)

	projID := NewProjectID()
	a.MoveToProject(projID)
	assert.Equal(t, projID, a.ProjectID())
	assert.Equal(t, createdAt, a.CreatedAt())
	assert.True(t, a.UpdatedAt().After(createdAt))
}

func TestAsset_Getters(t *testing.T) {
	id := NewID()
	groupID := NewGroupID()
	projectID := NewProjectID()
	workspaceID := NewWorkspaceID()
	now := time.Now()

	a := &Asset{
		id:          id,
		groupID:     groupID,
		projectID:   projectID,
		workspaceID: workspaceID,
		name:        "test.txt",
		size:        100,
		url:         "http://example.com",
		contentType: "text/plain",
		status:      StatusActive,
		error:       "test error",
		createdAt:   now,
		updatedAt:   now,
	}

	assert.Equal(t, id, a.ID())
	assert.Equal(t, groupID, a.GroupID())
	assert.Equal(t, projectID, a.ProjectID())
	assert.Equal(t, workspaceID, a.WorkspaceID())
	assert.Equal(t, "test.txt", a.Name())
	assert.Equal(t, int64(100), a.Size())
	assert.Equal(t, "http://example.com", a.URL())
	assert.Equal(t, "text/plain", a.ContentType())
	assert.Equal(t, StatusActive, a.Status())
	assert.Equal(t, "test error", a.Error())
	assert.Equal(t, now, a.CreatedAt())
	assert.Equal(t, now, a.UpdatedAt())
}

func TestMockNewID(t *testing.T) {
	id := NewID()
	cleanup := MockNewID(id)
	defer cleanup()

	assert.Equal(t, id, NewID())
	cleanup()
	assert.NotEqual(t, id, NewID())
}

func TestMockNewGroupID(t *testing.T) {
	id := NewGroupID()
	cleanup := MockNewGroupID(id)
	defer cleanup()

	assert.Equal(t, id, NewGroupID())
	cleanup()
	assert.NotEqual(t, id, NewGroupID())
}

func TestMockNewProjectID(t *testing.T) {
	id := NewProjectID()
	cleanup := MockNewProjectID(id)
	defer cleanup()

	assert.Equal(t, id, NewProjectID())
	cleanup()
	assert.NotEqual(t, id, NewProjectID())
}

func TestMockNewWorkspaceID(t *testing.T) {
	id := NewWorkspaceID()
	cleanup := MockNewWorkspaceID(id)
	defer cleanup()

	assert.Equal(t, id, NewWorkspaceID())
	cleanup()
	assert.NotEqual(t, id, NewWorkspaceID())
}
