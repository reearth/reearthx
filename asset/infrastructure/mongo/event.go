package mongo

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/infrastructure/mongo/mongodoc"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/mongox"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	eventIndexes       = []string{"user", "integration"}
	eventUniqueIndexes = []string{"id"}
)

type Event struct {
	client *mongox.Collection
}

func NewEvent(client *mongox.Client) repo.Event {
	return &Event{client: client.WithCollection("event")}
}

func (r *Event) Init() error {
	return createIndexes(context.Background(), r.client, eventIndexes, eventUniqueIndexes)
}

func (r *Event) FindByID(ctx context.Context, eventID id.EventID) (*event.Event[any], error) {
	return r.findOne(ctx, bson.M{
		"id": eventID.String(),
	})
}

func (r *Event) Save(ctx context.Context, ev *event.Event[any]) error {
	doc, eID, err := mongodoc.NewEvent(ev)
	if err != nil {
		return err
	}
	return r.client.SaveOne(ctx, eID, doc)
}

func (r *Event) SaveAll(ctx context.Context, ev event.List) error {
	doc, eID, err := mongodoc.NewEvents(ev)
	if err != nil {
		return err
	}
	return r.client.SaveAll(ctx, eID, lo.ToAnySlice(doc))
}

func (r *Event) findOne(ctx context.Context, filter any) (*event.Event[any], error) {
	c := mongodoc.NewEventConsumer()
	if err := r.client.FindOne(ctx, filter, c); err != nil {
		return nil, err
	}
	return c.Result[0], nil
}
