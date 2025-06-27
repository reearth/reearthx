package repo

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/id"
)

type Event interface {
	FindByID(context.Context, id.EventID) (*event.Event[any], error)
	Save(context.Context, *event.Event[any]) error
	SaveAll(context.Context, event.List) error
}
