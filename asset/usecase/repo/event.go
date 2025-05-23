package repo

import (
	"context"

	id "github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/event"
)

type Event interface {
	FindByID(context.Context, id.EventID) (*event.Event[any], error)
	Save(context.Context, *event.Event[any]) error
}
