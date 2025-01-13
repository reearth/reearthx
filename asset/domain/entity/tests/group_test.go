package entity_test

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/stretchr/testify/assert"
)

func TestNewGroup(t *testing.T) {
	groupID := id.NewGroupID()
	name := "test-group"

	group := entity.NewGroup(groupID, name)

	assert.Equal(t, groupID, group.ID())
	assert.Equal(t, name, group.Name())
	assert.Empty(t, group.Policy())
	assert.Empty(t, group.Description())
	assert.NotZero(t, group.CreatedAt())
	assert.NotZero(t, group.UpdatedAt())
	assert.Equal(t, group.CreatedAt(), group.UpdatedAt())
}

func TestGroup_UpdateName(t *testing.T) {
	group := entity.NewGroup(id.NewGroupID(), "test-group")
	initialUpdatedAt := group.UpdatedAt()
	time.Sleep(time.Millisecond)

	// Test valid name update
	err := group.UpdateName("new-name")
	assert.NoError(t, err)
	assert.Equal(t, "new-name", group.Name())
	assert.True(t, group.UpdatedAt().After(initialUpdatedAt))

	// Test empty name
	err = group.UpdateName("")
	assert.Equal(t, domain.ErrEmptyGroupName, err)
	assert.Equal(t, "new-name", group.Name()) // Name should not change
}

func TestGroup_UpdatePolicy(t *testing.T) {
	group := entity.NewGroup(id.NewGroupID(), "test-group")
	initialUpdatedAt := group.UpdatedAt()
	time.Sleep(time.Millisecond)

	// Test valid policy update
	err := group.UpdatePolicy("new-policy")
	assert.NoError(t, err)
	assert.Equal(t, "new-policy", group.Policy())
	assert.True(t, group.UpdatedAt().After(initialUpdatedAt))

	// Test empty policy
	err = group.UpdatePolicy("")
	assert.Equal(t, domain.ErrEmptyPolicy, err)
	assert.Equal(t, "new-policy", group.Policy()) // Policy should not change
}

func TestGroup_UpdateDescription(t *testing.T) {
	group := entity.NewGroup(id.NewGroupID(), "test-group")
	initialUpdatedAt := group.UpdatedAt()
	time.Sleep(time.Millisecond)

	// Test description update
	err := group.UpdateDescription("new description")
	assert.NoError(t, err)
	assert.Equal(t, "new description", group.Description())
	assert.True(t, group.UpdatedAt().After(initialUpdatedAt))

	// Test empty description (should be allowed)
	err = group.UpdateDescription("")
	assert.NoError(t, err)
	assert.Empty(t, group.Description())
}
