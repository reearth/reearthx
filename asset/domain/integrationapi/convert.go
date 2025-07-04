//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yml ../../schemas/integration.yml

package integrationapi

import (
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

var ErrUnsupportedEntity = rerror.NewE(i18n.T("unsupported entity"))

func New(obj any, v string) (res any, err error) {
	// note: version (v) is not used currently
	switch o := obj.(type) {
	case *event.Event[any]:
		res, err = NewEvent(o, v)
	case *asset.Asset:
		res = NewAsset(o, nil, true)
	case *asset.File:
		res = ToAssetFile(o, true)
	case *item.Item:
		res = NewItem(o, nil, nil)
	case item.Versioned:
		res = NewVersionedItem(o, nil, nil, nil, nil, nil, nil)
	case item.ItemModelSchema:
		res = NewItemModelSchema(o, nil)
	// TODO: add later
	// case *schema.Schema:
	// case *project.Project:
	// case *model.Model:
	// case *thread.Thread:
	// case *integration.Integration:
	// case *user.Workspace:
	// case *user.User:
	default:
		err = ErrUnsupportedEntity
	}

	return
}
