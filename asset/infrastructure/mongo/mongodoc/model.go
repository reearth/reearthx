package mongodoc

import (
	"time"

	"github.com/reearth/reearthx/asset/domain/model"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/mongox"
)

type ModelDocument struct {
	UpdatedAt   time.Time
	Metadata    *string
	ID          string
	Name        string
	Description string
	Key         string
	Project     string
	Schema      string
	Order       int
	Public      bool
}

func NewModel(model *model.Model) (*ModelDocument, string) {
	mId := model.ID().String()
	return &ModelDocument{
		ID:          mId,
		Name:        model.Name(),
		Description: model.Description(),
		Key:         model.Key().String(),
		Public:      model.Public(),
		Metadata:    model.Metadata().StringRef(),
		Project:     model.Project().String(),
		Schema:      model.Schema().String(),
		UpdatedAt:   model.UpdatedAt(),
		Order:       model.Order(),
	}, mId
}

func NewModels(models model.List) ([]*ModelDocument, []string) {
	res := make([]*ModelDocument, 0, len(models))
	ids := make([]string, 0, len(models))
	for _, d := range models {
		if d == nil {
			continue
		}
		r, rid := NewModel(d)
		res = append(res, r)
		ids = append(ids, rid)
	}
	return res, ids
}

func (d *ModelDocument) Model() (*model.Model, error) {
	mId, err := id.ModelIDFrom(d.ID)
	if err != nil {
		return nil, err
	}
	pId, err := id.ProjectIDFrom(d.Project)
	if err != nil {
		return nil, err
	}
	sId, err := id.SchemaIDFrom(d.Schema)
	if err != nil {
		return nil, err
	}

	return model.New().
		ID(mId).
		Name(d.Name).
		Description(d.Description).
		UpdatedAt(d.UpdatedAt).
		Key(id.NewKey(d.Key)).
		Public(d.Public).
		Project(pId).
		Metadata(id.SchemaIDFromRef(d.Metadata)).
		Schema(sId).
		Order(d.Order).
		Build()
}

type ModelConsumer = mongox.SliceFuncConsumer[*ModelDocument, *model.Model]

func NewModelConsumer() *ModelConsumer {
	return NewConsumer[*ModelDocument, *model.Model]()
}
