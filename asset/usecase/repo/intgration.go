package repo

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/integration"

	"github.com/reearth/reearthx/account/accountdomain"
)

type Integration interface {
	FindByIDs(context.Context, id.IntegrationIDList) (integration.List, error)
	FindByUser(context.Context, accountdomain.UserID) (integration.List, error)
	FindByID(context.Context, id.IntegrationID) (*integration.Integration, error)
	FindByToken(context.Context, string) (*integration.Integration, error)
	Save(context.Context, *integration.Integration) error
	Remove(context.Context, id.IntegrationID) error
	RemoveMany(context.Context, id.IntegrationIDList) error
}
