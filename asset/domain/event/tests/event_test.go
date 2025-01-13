package event_test

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/stretchr/testify/assert"
)

func TestBaseEvent(t *testing.T) {
	before := time.Now()
	e := event.NewBaseEvent()
	after := time.Now()

	assert.True(t, e.OccurredAt().After(before) || e.OccurredAt().Equal(before))
	assert.True(t, e.OccurredAt().Before(after) || e.OccurredAt().Equal(after))
}

func TestAssetEvents(t *testing.T) {
	assetID := id.NewID()
	asset := entity.NewAsset(assetID, "test.jpg", 1024, "image/jpeg")

	t.Run("AssetCreated", func(t *testing.T) {
		e := event.NewAssetCreated(asset)
		assert.Equal(t, "asset.created", e.EventType())
		assert.Equal(t, asset, e.Asset)
		assert.NotZero(t, e.OccurredAt())
	})

	t.Run("AssetUpdated", func(t *testing.T) {
		e := event.NewAssetUpdated(asset)
		assert.Equal(t, "asset.updated", e.EventType())
		assert.Equal(t, asset, e.Asset)
		assert.NotZero(t, e.OccurredAt())
	})

	t.Run("AssetDeleted", func(t *testing.T) {
		e := event.NewAssetDeleted(assetID)
		assert.Equal(t, "asset.deleted", e.EventType())
		assert.Equal(t, assetID, e.AssetID)
		assert.NotZero(t, e.OccurredAt())
	})
}

func TestGroupEvents(t *testing.T) {
	groupID := id.NewGroupID()
	group := entity.NewGroup(groupID, "test-group")

	t.Run("GroupCreated", func(t *testing.T) {
		e := event.NewGroupCreated(group)
		assert.Equal(t, "group.created", e.EventType())
		assert.Equal(t, group, e.Group)
		assert.NotZero(t, e.OccurredAt())
	})

	t.Run("GroupUpdated", func(t *testing.T) {
		e := event.NewGroupUpdated(group)
		assert.Equal(t, "group.updated", e.EventType())
		assert.Equal(t, group, e.Group)
		assert.NotZero(t, e.OccurredAt())
	})

	t.Run("GroupDeleted", func(t *testing.T) {
		e := event.NewGroupDeleted(groupID)
		assert.Equal(t, "group.deleted", e.EventType())
		assert.Equal(t, groupID, e.GroupID)
		assert.NotZero(t, e.OccurredAt())
	})
}
