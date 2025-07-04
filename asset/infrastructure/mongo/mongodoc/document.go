package mongodoc

import (
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/integration"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/asset/domain/model"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/reearth/reearthx/asset/domain/thread"
	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/asset/infrastructure/mongo/mongogit"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"go.mongodb.org/mongo-driver/bson"
)

type Type string

type Document struct {
	Type   Type
	Object bson.Raw
}

var (
	ErrInvalidObject = rerror.NewE(i18n.T("invalid object"))
	ErrInvalidDoc    = rerror.NewE(i18n.T("invalid document"))
)

func NewDocument(obj any) (doc Document, id string, err error) {
	var res any
	var ty Type

	switch m := obj.(type) {
	case *asset.Asset:
		ty = "asset"
		res, id = NewAsset(m)
	case *item.Item:
		ty = "item"
		res, id = NewItem(m)
	case item.Versioned:
		ty = "item"
		res, id = NewItem(m.Value())
		res = mongogit.NewDocument(version.ValueFrom(m, res))
	case *schema.Schema:
		ty = "schema"
		res, id = NewSchema(m)
	case *project.Project:
		ty = "project"
		res, id = NewProject(m)
	case *model.Model:
		ty = "model"
		res, id = NewModel(m)
	case *thread.Thread:
		ty = "thread"
		res, id = NewThread(m)
	case *integration.Integration:
		ty = "integration"
		res, id = NewIntegration(m)
	default:
		err = ErrInvalidObject
		return
	}

	raw, err := bson.Marshal(res)
	if err != nil {
		return
	}
	return Document{Object: raw, Type: ty}, id, nil
}

func ModelFrom(obj Document) (res any, err error) {
	switch obj.Type {
	case "asset":
		var d *AssetDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	case "item":
		var d *ItemDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	case "schema":
		var d *SchemaDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	case "project":
		var d *ProjectDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	case "model":
		var d *ModelDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	case "thread":
		var d *ThreadDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	case "integration":
		var d *IntegrationDocument
		if err = bson.Unmarshal(obj.Object, &d); err == nil {
			res, err = d.Model()
		}
	default:
		err = ErrInvalidDoc
	}
	return
}
