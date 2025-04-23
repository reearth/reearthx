package asset

import (
	"context"
	"errors"
)

type policyService struct {
	policyRepo PolicyRepository
}

func NewPolicyService(
	policyRepo PolicyRepository,
) PolicyService {
	return &policyService{
		policyRepo: policyRepo,
	}
}

func (s *policyService) CreatePolicy(ctx context.Context, name string, storageLimit int64) (*Policy, error) {
	if name == "" {
		return nil, errors.New("policy name cannot be empty")
	}

	if storageLimit <= 0 {
		return nil, errors.New("storage limit must be greater than zero")
	}

	policy := NewPolicy(name, storageLimit)

	if err := s.policyRepo.Save(ctx, policy); err != nil {
		return nil, err
	}

	return policy, nil
}

func (s *policyService) GetPolicy(ctx context.Context, id PolicyID) (*Policy, error) {
	policy, err := s.policyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, errors.New("policy not found")
	}
	return policy, nil
}

func (s *policyService) DeletePolicy(ctx context.Context, id PolicyID) error {
	policy, err := s.policyRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if policy == nil {
		return errors.New("policy not found")
	}

	return s.policyRepo.Delete(ctx, id)
}
