package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewGroup(t *testing.T) {
	id := NewGroupID()
	g := NewGroup(id, "test-group")

	assert.Equal(t, id, g.ID())
	assert.Equal(t, "test-group", g.Name())
	assert.Empty(t, g.Policy())
	assert.Empty(t, g.Description())
	assert.NotZero(t, g.CreatedAt())
	assert.NotZero(t, g.UpdatedAt())
	assert.Equal(t, g.CreatedAt(), g.UpdatedAt())
}

func TestGroup_UpdateName(t *testing.T) {
	g := NewGroup(NewGroupID(), "test-group")
	createdAt := g.CreatedAt()
	time.Sleep(time.Millisecond)

	g.UpdateName("new-name")
	assert.Equal(t, "new-name", g.Name())
	assert.Equal(t, createdAt, g.CreatedAt())
	assert.True(t, g.UpdatedAt().After(createdAt))
}

func TestGroup_UpdateDescription(t *testing.T) {
	g := NewGroup(NewGroupID(), "test-group")
	createdAt := g.CreatedAt()
	time.Sleep(time.Millisecond)

	g.UpdateDescription("test description")
	assert.Equal(t, "test description", g.Description())
	assert.Equal(t, createdAt, g.CreatedAt())
	assert.True(t, g.UpdatedAt().After(createdAt))
}

func TestGroup_UpdatePolicy(t *testing.T) {
	g := NewGroup(NewGroupID(), "test-group")
	createdAt := g.CreatedAt()
	time.Sleep(time.Millisecond)

	g.UpdatePolicy("test-policy")
	assert.Equal(t, "test-policy", g.Policy())
	assert.Equal(t, createdAt, g.CreatedAt())
	assert.True(t, g.UpdatedAt().After(createdAt))
}
