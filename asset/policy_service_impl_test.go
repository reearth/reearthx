package asset

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePolicy(t *testing.T) {
	ctx := context.Background()
	policyRepo := new(MockPolicyRepository)

	service := NewPolicyService(policyRepo)

	t.Run("Create policy with valid parameters", func(t *testing.T) {
		policyRepo.On("Save", ctx, mock.AnythingOfType("*asset.Policy")).Return(nil)

		name := "Test Policy"
		storageLimit := int64(1024 * 1024 * 100)
		policy, err := service.CreatePolicy(ctx, name, storageLimit)

		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Equal(t, name, policy.Name)
		assert.Equal(t, storageLimit, policy.StorageLimit)
		assert.NotEqual(t, PolicyID{}, policy.ID)

		policyRepo.AssertExpectations(t)
	})

	t.Run("Create policy with empty name", func(t *testing.T) {
		name := ""
		storageLimit := int64(1024 * 1024 * 100)
		policy, err := service.CreatePolicy(ctx, name, storageLimit)

		assert.Error(t, err)
		assert.Nil(t, policy)
	})

	t.Run("Create policy with invalid storage limit", func(t *testing.T) {
		name := "Test Policy"
		storageLimit := int64(-1)
		policy, err := service.CreatePolicy(ctx, name, storageLimit)

		assert.Error(t, err)
		assert.Nil(t, policy)
	})
}

func TestGetPolicy(t *testing.T) {
	ctx := context.Background()
	policyRepo := new(MockPolicyRepository)

	service := NewPolicyService(policyRepo)

	t.Run("Get existing policy", func(t *testing.T) {
		policyID := NewPolicyID()
		expectedPolicy := &Policy{
			ID:           policyID,
			Name:         "Test Policy",
			StorageLimit: 1024 * 1024 * 100,
		}

		policyRepo.On("FindByID", ctx, policyID).Return(expectedPolicy, nil)

		policy, err := service.GetPolicy(ctx, policyID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPolicy, policy)

		policyRepo.AssertExpectations(t)
	})

	t.Run("Get nonexistent policy", func(t *testing.T) {
		policyID := NewPolicyID()

		policyRepo.On("FindByID", ctx, policyID).Return(nil, nil)

		policy, err := service.GetPolicy(ctx, policyID)

		assert.Error(t, err)
		assert.Nil(t, policy)

		policyRepo.AssertExpectations(t)
	})
}

func TestDeletePolicy(t *testing.T) {
	ctx := context.Background()
	policyRepo := new(MockPolicyRepository)

	service := NewPolicyService(policyRepo)

	t.Run("Delete existing policy", func(t *testing.T) {
		policyID := NewPolicyID()
		expectedPolicy := &Policy{
			ID:           policyID,
			Name:         "Test Policy",
			StorageLimit: 1024 * 1024 * 100,
		}

		policyRepo.On("FindByID", ctx, policyID).Return(expectedPolicy, nil)
		policyRepo.On("Delete", ctx, policyID).Return(nil)

		err := service.DeletePolicy(ctx, policyID)

		assert.NoError(t, err)

		policyRepo.AssertExpectations(t)
	})

	t.Run("Delete nonexistent policy", func(t *testing.T) {
		policyID := NewPolicyID()

		policyRepo.On("FindByID", ctx, policyID).Return(nil, nil)

		err := service.DeletePolicy(ctx, policyID)

		assert.Error(t, err)

		policyRepo.AssertExpectations(t)
	})
}
