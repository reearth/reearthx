package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/operator"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
)

func TestEvent_Save(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond).UTC()
	u := user.New().NewID().Email("hoge@example.com").Name("John").MustBuild()
	a := asset.New().NewID().Thread(id.NewThreadID().Ref()).NewUUID().
		Project(project.NewID()).Size(100).CreatedAt(now).CreatedByUser(u.ID()).MustBuild()
	eID := event.NewID()
	ev := event.New[any]().ID(eID).Timestamp(now).Type(event.AssetCreate).
		Operator(operator.OperatorFromUser(u.ID())).Object(a).MustBuild()

	initDB := mongotest.Connect(t)

	client := mongox.NewClientWithDatabase(initDB(t))
	r := NewEvent(client)
	ctx := context.Background()
	err := r.Save(ctx, ev)
	assert.NoError(t, err)

	got, err := r.FindByID(ctx, eID)
	assert.Equal(t, ev, got)
	assert.NoError(t, err)
}
