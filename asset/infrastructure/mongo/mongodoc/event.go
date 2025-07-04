package mongodoc

import (
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/operator"
	"github.com/reearth/reearthx/mongox"
)

type EventDocument struct {
	Timestamp   time.Time
	User        *string
	Integration *string
	ID          string
	Type        string
	Object      Document
	Machine     bool
}

func NewEvent(e *event.Event[any]) (*EventDocument, string, error) {
	eId := e.ID().String()
	objDoc, _, err := NewDocument(e.Object())
	if err != nil {
		return nil, "", err
	}
	return &EventDocument{
		ID:          eId,
		Timestamp:   e.Timestamp(),
		User:        e.Operator().User().StringRef(),
		Integration: e.Operator().Integration().StringRef(),
		Machine:     e.Operator().Machine(),
		Type:        string(e.Type()),
		Object:      objDoc,
	}, eId, nil
}

func NewEvents(e event.List) ([]*EventDocument, []string, error) {
	res := make([]*EventDocument, 0, len(e))
	ids := make([]string, 0, len(e))
	for _, d := range e {
		if d == nil {
			continue
		}
		r, rid, err := NewEvent(d)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, r)
		ids = append(ids, rid)
	}
	return res, ids, nil
}

func (d *EventDocument) Model() (*event.Event[any], error) {
	eID, err := event.IDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	m, err := ModelFrom(d.Object)
	if err != nil {
		return nil, err
	}

	var o operator.Operator
	switch {
	case d.User != nil:
		if uid := accountdomain.UserIDFromRef(d.User); uid != nil {
			o = operator.OperatorFromUser(*uid)
		}
	case d.Integration != nil:
		if iid := id.IntegrationIDFromRef(d.Integration); iid != nil {
			o = operator.OperatorFromIntegration(*iid)
		}
	case d.Machine:
		o = operator.OperatorFromMachine()
	}

	e, err := event.New[any]().
		ID(eID).
		Type(event.Type(d.Type)).
		Timestamp(d.Timestamp).
		Operator(o).
		Object(m).
		Build()
	if err != nil {
		return nil, err
	}

	return e, nil
}

type EventConsumer = mongox.SliceFuncConsumer[*EventDocument, *event.Event[any]]

func NewEventConsumer() *EventConsumer {
	return NewConsumer[*EventDocument, *event.Event[any]]()
}
