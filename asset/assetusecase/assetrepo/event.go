package assetrepo

import (
	"context"

	id "github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/asset/assetdomain/event"
)

type Event interface {
	FindByID(context.Context, id.EventID) (*event.Event[any], error)
	Save(context.Context, *event.Event[any]) error
}
