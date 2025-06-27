package group

import (
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/samber/lo"
)

type (
	ID        = id.GroupID
	ProjectID = id.ProjectID
	SchemaID  = id.SchemaID
)

var (
	NewID        = id.NewGroupID
	MustID       = id.MustGroupID
	IDFrom       = id.GroupIDFrom
	IDFromRef    = id.GroupIDFromRef
	ErrInvalidID = id.ErrInvalidID
)

type IDOrKey string

func (i IDOrKey) ID() *ID {
	return IDFromRef(lo.ToPtr(string(i)))
}

func (i IDOrKey) Key() *string {
	if i.ID() == nil {
		return lo.ToPtr(string(i))
	}
	return nil
}
