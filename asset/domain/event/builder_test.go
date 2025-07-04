package event

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/operator"
	"github.com/reearth/reearthx/asset/domain/project"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/stretchr/testify/assert"
)

var (
	u = user.New().NewID().Email("hoge@example.com").Name("John").MustBuild()
	a = asset.New().NewID().Project(project.NewID()).NewUUID().
		Thread(id.NewThreadID().Ref()).Size(100).CreatedByUser(u.ID()).
		MustBuild()
)

func TestBuilder(t *testing.T) {
	now := time.Now()
	id := NewID()

	ev := New[*asset.Asset]().ID(id).Timestamp(now).
		Type(AssetCreate).Operator(operator.OperatorFromUser(u.ID())).Object(a).MustBuild()
	ev2 := New[*asset.Asset]().NewID().Timestamp(now).
		Type(AssetDecompress).Operator(operator.OperatorFromUser(u.ID())).Object(a).MustBuild()

	// ev1
	assert.Equal(t, id, ev.ID())
	assert.Equal(t, Type(AssetCreate), ev.Type())
	assert.Equal(t, operator.OperatorFromUser(u.ID()), ev.Operator())
	assert.Equal(t, a, ev.Object())

	// ev2
	assert.NotNil(t, ev2.ID())

	ev3, err := New[*asset.Asset]().Build()
	assert.Equal(t, ErrInvalidID, err)
	assert.Nil(t, ev3)
}
