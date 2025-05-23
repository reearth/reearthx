package event

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain/user"
	id "github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/operator"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	u := user.New().NewID().Email("hoge@example.com").Name("John").MustBuild()
	a := asset.New().NewID().Thread(id.NewThreadID().Ref()).NewUUID().
		Project(project.NewID()).Size(100).CreatedByUser(u.ID()).MustBuild()
	now := time.Now()
	eID := NewID()
	ev := New[*asset.Asset]().ID(eID).Timestamp(now).Type(AssetCreate).
		Operator(operator.OperatorFromUser(u.ID())).Object(a).MustBuild()

	assert.Equal(t, eID, ev.ID())
	assert.Equal(t, Type(AssetCreate), ev.Type())
	assert.Equal(t, operator.OperatorFromUser(u.ID()), ev.Operator())
	assert.Equal(t, a, ev.Object())
	assert.Equal(t, now, ev.Timestamp())
	assert.Equal(t, ev, ev.Clone())
	assert.NotSame(t, ev, ev.Clone())
}
