package service

import (
	"context"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
	"github.com/reearth/reearthx/log"
)

type GroupService struct {
	repo   repository.GroupRepository
	pubsub repository.PubSubRepository
}

func NewGroupService(repo repository.GroupRepository, pubsub repository.PubSubRepository) *GroupService {
	return &GroupService{
		repo:   repo,
		pubsub: pubsub,
	}
}

func (s *GroupService) Create(ctx context.Context, group *domain.Group) error {
	if err := s.repo.Create(ctx, group); err != nil {
		return err
	}

	// Create a dummy asset for event publishing
	asset := domain.NewAsset(domain.NewID(), group.Name(), 0, "")
	if err := s.pubsub.PublishAssetCreated(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish group created event: %v", err)
	}

	return nil
}

func (s *GroupService) Get(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *GroupService) Update(ctx context.Context, group *domain.Group) error {
	if err := s.repo.Update(ctx, group); err != nil {
		return err
	}

	// Create a dummy asset for event publishing
	asset := domain.NewAsset(domain.NewID(), group.Name(), 0, "")
	if err := s.pubsub.PublishAssetUpdated(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish group updated event: %v", err)
	}

	return nil
}

func (s *GroupService) Delete(ctx context.Context, id domain.GroupID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Create a dummy asset ID for event publishing
	assetID := domain.NewID()
	if err := s.pubsub.PublishAssetDeleted(ctx, assetID); err != nil {
		log.Errorfc(ctx, "failed to publish group deleted event: %v", err)
	}

	return nil
}

func (s *GroupService) List(ctx context.Context) ([]*domain.Group, error) {
	return s.repo.List(ctx)
}

func (s *GroupService) AssignPolicy(ctx context.Context, id domain.GroupID, policy string) error {
	group, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := group.AssignPolicy(policy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, group); err != nil {
		return err
	}

	if err := s.pubsub.PublishAssetUpdated(ctx, nil); err != nil {
		log.Errorfc(ctx, "failed to publish group policy updated event: %v", err)
	}

	return nil
}
