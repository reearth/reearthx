package asset

import (
	"context"
	"errors"
)

var _ GroupService = &groupService{}

type groupService struct {
	groupRepo GroupRepository
}

func NewGroupService(
	groupRepo GroupRepository,
) GroupService {
	return &groupService{
		groupRepo: groupRepo,
	}
}

func (s *groupService) CreateGroup(ctx context.Context, name string) (*Group, error) {
	if name == "" {
		return nil, errors.New("group name cannot be empty")
	}

	group := NewGroup(name)

	if err := s.groupRepo.Save(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) GetGroup(ctx context.Context, id GroupID) (*Group, error) {
	group, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func (s *groupService) DeleteGroup(ctx context.Context, id GroupID) error {
	group, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	return s.groupRepo.Delete(ctx, id)
}

func (s *groupService) AssignPolicy(ctx context.Context, groupID GroupID, policyID *PolicyID) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	return s.groupRepo.UpdatePolicy(ctx, groupID, policyID)
}
