package service

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGroupRepo struct {
	mock.Mock
}

func (m *mockGroupRepo) FindByID(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *mockGroupRepo) FindByIDs(ctx context.Context, ids []domain.GroupID) ([]*domain.Group, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*domain.Group), args.Error(1)
}

func (m *mockGroupRepo) List(ctx context.Context) ([]*domain.Group, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Group), args.Error(1)
}

func (m *mockGroupRepo) Create(ctx context.Context, group *domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *mockGroupRepo) Update(ctx context.Context, group *domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *mockGroupRepo) Delete(ctx context.Context, id domain.GroupID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockPubSub struct {
	mock.Mock
}

func (m *mockPubSub) PublishAssetCreated(ctx context.Context, asset *domain.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockPubSub) PublishAssetUpdated(ctx context.Context, asset *domain.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockPubSub) PublishAssetDeleted(ctx context.Context, id domain.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPubSub) PublishAssetUploaded(ctx context.Context, asset *domain.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockPubSub) PublishAssetExtracted(ctx context.Context, asset *domain.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockPubSub) PublishAssetTransferred(ctx context.Context, asset *domain.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockPubSub) Subscribe(eventType repository.EventType, handler repository.EventHandler) {
	m.Called(eventType, handler)
}

func (m *mockPubSub) Unsubscribe(eventType repository.EventType, handler repository.EventHandler) {
	m.Called(eventType, handler)
}

func TestGroupService_Create(t *testing.T) {
	ctx := context.Background()
	repo := new(mockGroupRepo)
	pubsub := new(mockPubSub)
	service := NewGroupService(repo, pubsub)

	group := domain.NewGroup(domain.NewGroupID(), "test-group")

	repo.On("Create", ctx, group).Return(nil)
	pubsub.On("PublishAssetCreated", ctx, mock.AnythingOfType("*domain.Asset")).Return(nil)

	err := service.Create(ctx, group)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	pubsub.AssertExpectations(t)
}

func TestGroupService_Get(t *testing.T) {
	ctx := context.Background()
	repo := new(mockGroupRepo)
	pubsub := new(mockPubSub)
	service := NewGroupService(repo, pubsub)

	id := domain.NewGroupID()
	group := domain.NewGroup(id, "test-group")

	repo.On("FindByID", ctx, id).Return(group, nil)

	result, err := service.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, group, result)
	repo.AssertExpectations(t)
}

func TestGroupService_AssignPolicy(t *testing.T) {
	ctx := context.Background()
	repo := new(mockGroupRepo)
	pubsub := new(mockPubSub)
	service := NewGroupService(repo, pubsub)

	id := domain.NewGroupID()
	group := domain.NewGroup(id, "test-group")
	policy := "test-policy"

	repo.On("FindByID", ctx, id).Return(group, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Group")).Return(nil)
	pubsub.On("PublishAssetUpdated", ctx, mock.AnythingOfType("*domain.Asset")).Return(nil)

	err := service.AssignPolicy(ctx, id, policy)
	assert.NoError(t, err)
	assert.Equal(t, policy, group.Policy())
	repo.AssertExpectations(t)
	pubsub.AssertExpectations(t)
}
