package asset

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func TestCreateGroup(t *testing.T) {
// 	ctx := context.Background()
// 	groupRepo := new(MockGroupRepository)

// 	service := NewGroupService(groupRepo)

// 	t.Run("Create group", func(t *testing.T) {
// 		groupRepo.On("Save", ctx, mock.AnythingOfType("*asset.Group")).Return(nil)

// 		name := "Test Group"
// 		group, err := service.CreateGroup(ctx, name)

// 		assert.NoError(t, err)
// 		assert.NotNil(t, group)
// 		assert.Equal(t, name, group.Name)
// 		assert.NotEqual(t, GroupID{}, group.ID)
// 		assert.WithinDuration(t, time.Now(), group.CreatedAt, 2*time.Second)
// 		assert.Nil(t, group.PolicyID)

// 		groupRepo.AssertExpectations(t)
// 	})

// 	t.Run("Create group with empty name", func(t *testing.T) {
// 		name := ""
// 		group, err := service.CreateGroup(ctx, name)

// 		assert.Error(t, err)
// 		assert.Nil(t, group)
// 		assert.Equal(t, "group name cannot be empty", err.Error())
// 	})
// }

// func TestGetGroup(t *testing.T) {
// 	ctx := context.Background()
// 	groupRepo := new(MockGroupRepository)

// 	service := NewGroupService(groupRepo)

// 	t.Run("Get existing group", func(t *testing.T) {
// 		groupID := NewGroupID()
// 		expectedGroup := &Group{
// 			ID:        groupID,
// 			Name:      "Test Group",
// 			CreatedAt: time.Now(),
// 		}

// 		groupRepo.On("FindByID", ctx, groupID).Return(expectedGroup, nil)

// 		group, err := service.GetGroup(ctx, groupID)

// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedGroup, group)

// 		groupRepo.AssertExpectations(t)
// 	})

// 	t.Run("Get nonexistent group", func(t *testing.T) {
// 		groupID := NewGroupID()

// 		groupRepo.On("FindByID", ctx, groupID).Return(nil, nil)

// 		group, err := service.GetGroup(ctx, groupID)

// 		assert.Error(t, err)
// 		assert.Nil(t, group)
// 		assert.Equal(t, "group not found", err.Error())

// 		groupRepo.AssertExpectations(t)
// 	})
// }

// func TestAssignPolicy(t *testing.T) {
// 	ctx := context.Background()
// 	groupRepo := new(MockGroupRepository)

// 	service := NewGroupService(groupRepo)

// 	t.Run("Assign policy to existing group", func(t *testing.T) {
// 		groupID := NewGroupID()
// 		policyID := NewPolicyID()

// 		group := &Group{
// 			ID:        groupID,
// 			Name:      "Test Group",
// 			CreatedAt: time.Now(),
// 		}

// 		groupRepo.On("FindByID", ctx, groupID).Return(group, nil)
// 		groupRepo.On("UpdatePolicy", ctx, groupID, &policyID).Return(nil)

// 		err := service.AssignPolicy(ctx, groupID, &policyID)

// 		assert.NoError(t, err)

// 		groupRepo.AssertExpectations(t)
// 	})

// 	t.Run("Remove policy from group", func(t *testing.T) {
// 		groupID := NewGroupID()

// 		group := &Group{
// 			ID:        groupID,
// 			Name:      "Test Group",
// 			CreatedAt: time.Now(),
// 			PolicyID:  new(PolicyID),
// 		}

// 		groupRepo.On("FindByID", ctx, groupID).Return(group, nil)
// 		groupRepo.On("UpdatePolicy", ctx, groupID, (*PolicyID)(nil)).Return(nil)

// 		err := service.AssignPolicy(ctx, groupID, nil)

// 		assert.NoError(t, err)

// 		groupRepo.AssertExpectations(t)
// 	})

// 	t.Run("Assign policy to nonexistent group", func(t *testing.T) {
// 		groupID := NewGroupID()
// 		policyID := NewPolicyID()

// 		groupRepo.On("FindByID", ctx, groupID).Return(nil, nil)

// 		err := service.AssignPolicy(ctx, groupID, &policyID)

// 		assert.Error(t, err)
// 		assert.Equal(t, "group not found", err.Error())

// 		groupRepo.AssertExpectations(t)
// 	})
// }

// func TestDeleteGroup(t *testing.T) {
// 	ctx := context.Background()
// 	groupRepo := new(MockGroupRepository)

// 	service := NewGroupService(groupRepo)

// 	t.Run("Delete existing group", func(t *testing.T) {
// 		groupID := NewGroupID()

// 		group := &Group{
// 			ID:        groupID,
// 			Name:      "Test Group",
// 			CreatedAt: time.Now(),
// 		}

// 		groupRepo.On("FindByID", ctx, groupID).Return(group, nil)
// 		groupRepo.On("Delete", ctx, groupID).Return(nil)

// 		err := service.DeleteGroup(ctx, groupID)

// 		assert.NoError(t, err)

// 		groupRepo.AssertExpectations(t)
// 	})

// 	t.Run("Delete nonexistent group", func(t *testing.T) {
// 		groupID := NewGroupID()

// 		groupRepo.On("FindByID", ctx, groupID).Return(nil, nil)

// 		err := service.DeleteGroup(ctx, groupID)

// 		assert.Error(t, err)
// 		assert.Equal(t, "group not found", err.Error())

// 		groupRepo.AssertExpectations(t)
// 	})
// }

// type MockPolicyRepository struct {
// 	mock.Mock
// }

// func (m *MockPolicyRepository) Save(ctx context.Context, policy *Policy) error {
// 	args := m.Called(ctx, policy)
// 	return args.Error(0)
// }

// func (m *MockPolicyRepository) FindByID(ctx context.Context, id PolicyID) (*Policy, error) {
// 	args := m.Called(ctx, id)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*Policy), args.Error(1)
// }

// func (m *MockPolicyRepository) Delete(ctx context.Context, id PolicyID) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }
