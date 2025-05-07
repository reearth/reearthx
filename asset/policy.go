package asset

import (
	"time"
)

type Policy struct {
	ID           PolicyID
	Name         string
	StorageLimit int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewPolicy(name string, storageLimit int64) *Policy {
	now := time.Now()
	return &Policy{
		ID:           NewPolicyID(),
		Name:         name,
		StorageLimit: storageLimit,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (p *Policy) Clone() *Policy {
	if p == nil {
		return nil
	}

	clone := *p
	return &clone
}

func (p *Policy) UpdateStorageLimit(storageLimit int64) {
	p.StorageLimit = storageLimit
	p.UpdatedAt = time.Now()
}

func (p *Policy) UpdateName(name string) {
	p.Name = name
	p.UpdatedAt = time.Now()
}
