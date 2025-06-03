package mongodoc

import (
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/asset/infrastructure/mongo/mongogit"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)

type ItemDocument struct {
	Timestamp            time.Time
	Thread               *string
	User                 *string
	Integration          *string
	MetadataItem         *string
	OriginalItem         *string
	UpdatedByUser        *string
	UpdatedByIntegration *string
	ID                   string
	Project              string
	Schema               string
	ModelID              string
	Fields               []ItemFieldDocument
	Assets               []string `bson:"assets,omitempty"`
	IsMetadata           bool
}

type ItemFieldDocument struct {
	ItemGroup *string
	F         string        `bson:"f,omitempty"`
	V         ValueDocument `bson:"v,omitempty"`
}

type ItemConsumer = mongox.SliceFuncConsumer[*ItemDocument, *item.Item]

func NewItemConsumer() *ItemConsumer {
	return NewConsumer[*ItemDocument, *item.Item]()
}

type VersionedItemConsumer = mongox.SliceFuncConsumer[*mongogit.Document[*ItemDocument], *version.Value[*item.Item]]

func NewVersionedItemConsumer() *VersionedItemConsumer {
	return mongox.NewSliceFuncConsumer(func(d *mongogit.Document[*ItemDocument]) (*version.Value[*item.Item], error) {
		itm, err := d.Data.Model()
		if err != nil {
			return nil, err
		}

		v := mongogit.ToValue(d.Meta, itm)
		return v, nil
	})
}

func NewItem(i *item.Item) (*ItemDocument, string) {
	itmId := i.ID().String()
	return &ItemDocument{
		ID:           itmId,
		Schema:       i.Schema().String(),
		ModelID:      i.Model().String(),
		Project:      i.Project().String(),
		MetadataItem: i.MetadataItem().StringRef(),
		OriginalItem: i.OriginalItem().StringRef(),
		Fields: lo.FilterMap(i.Fields(), func(f *item.Field, _ int) (ItemFieldDocument, bool) {
			v := NewMultipleValue(f.Value())
			if v == nil {
				return ItemFieldDocument{}, false
			}

			return ItemFieldDocument{
				ItemGroup: f.ItemGroup().StringRef(),
				F:         f.FieldID().String(),
				V:         *v,
			}, true
		}),
		Timestamp:            i.Timestamp(),
		User:                 i.User().StringRef(),
		UpdatedByUser:        i.UpdatedByUser().StringRef(),
		UpdatedByIntegration: i.UpdatedByIntegration().StringRef(),
		Integration:          i.Integration().StringRef(),
		Assets:               i.AssetIDs().Strings(),
		IsMetadata:           i.IsMetadata(),
		Thread:               i.Thread().StringRef(),
	}, itmId
}

func (d *ItemDocument) Model() (*item.Item, error) {
	itmId, err := id.ItemIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	sid, err := id.SchemaIDFrom(d.Schema)
	if err != nil {
		return nil, err
	}

	mid, err := id.ModelIDFrom(d.ModelID)
	if err != nil {
		return nil, err
	}

	pid, err := id.ProjectIDFrom(d.Project)
	if err != nil {
		return nil, err
	}

	fields, err := util.TryMap(d.Fields, func(f ItemFieldDocument) (*item.Field, error) {
		sf, err := item.FieldIDFrom(f.F)
		if err != nil {
			return nil, err
		}
		ig := id.ItemGroupIDFromRef(f.ItemGroup)
		return item.NewField(sf, f.V.MultipleValue(), ig), nil
	})
	if err != nil {
		return nil, err
	}

	ib := item.New().
		ID(itmId).
		Project(pid).
		Schema(sid).
		UpdatedByUser(accountdomain.UserIDFromRef(d.UpdatedByUser)).
		UpdatedByIntegration(id.IntegrationIDFromRef(d.UpdatedByIntegration)).
		Model(mid).
		MetadataItem(id.ItemIDFromRef(d.MetadataItem)).
		OriginalItem(id.ItemIDFromRef(d.OriginalItem)).
		IsMetadata(d.IsMetadata).
		Fields(fields).
		Timestamp(d.Timestamp).
		Thread(id.ThreadIDFromRef(d.Thread))

	if uId := accountdomain.UserIDFromRef(d.User); uId != nil {
		ib = ib.User(*uId)
	}

	if iId := id.IntegrationIDFromRef(d.Integration); iId != nil {
		ib = ib.Integration(*iId)
	}

	return ib.Build()
}

func NewItems(items item.List) ([]*ItemDocument, []string) {
	res := make([]*ItemDocument, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, d := range items {
		if d == nil {
			continue
		}
		r, itmId := NewItem(d)
		res = append(res, r)
		ids = append(ids, itmId)
	}
	return res, ids
}
