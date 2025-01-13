package event

import (
	"time"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

// Event represents a domain event
type Event interface {
	EventType() string
	OccurredAt() time.Time
}

// BaseEvent contains common event fields
type BaseEvent struct {
	occurredAt time.Time
}

func NewBaseEvent() BaseEvent {
	return BaseEvent{occurredAt: time.Now()}
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Asset Events
type AssetCreated struct {
	BaseEvent
	Asset *entity.Asset
}

func NewAssetCreated(asset *entity.Asset) *AssetCreated {
	return &AssetCreated{
		BaseEvent: NewBaseEvent(),
		Asset:     asset,
	}
}

func (e AssetCreated) EventType() string { return "asset.created" }

type AssetUpdated struct {
	BaseEvent
	Asset *entity.Asset
}

func NewAssetUpdated(asset *entity.Asset) *AssetUpdated {
	return &AssetUpdated{
		BaseEvent: NewBaseEvent(),
		Asset:     asset,
	}
}

func (e AssetUpdated) EventType() string { return "asset.updated" }

type AssetDeleted struct {
	BaseEvent
	AssetID id.ID
}

func NewAssetDeleted(assetID id.ID) *AssetDeleted {
	return &AssetDeleted{
		BaseEvent: NewBaseEvent(),
		AssetID:   assetID,
	}
}

func (e AssetDeleted) EventType() string { return "asset.deleted" }

// Group Events
type GroupCreated struct {
	BaseEvent
	Group *entity.Group
}

func NewGroupCreated(group *entity.Group) *GroupCreated {
	return &GroupCreated{
		BaseEvent: NewBaseEvent(),
		Group:     group,
	}
}

func (e GroupCreated) EventType() string { return "group.created" }

type GroupUpdated struct {
	BaseEvent
	Group *entity.Group
}

func NewGroupUpdated(group *entity.Group) *GroupUpdated {
	return &GroupUpdated{
		BaseEvent: NewBaseEvent(),
		Group:     group,
	}
}

func (e GroupUpdated) EventType() string { return "group.updated" }

type GroupDeleted struct {
	BaseEvent
	GroupID id.GroupID
}

func NewGroupDeleted(groupID id.GroupID) *GroupDeleted {
	return &GroupDeleted{
		BaseEvent: NewBaseEvent(),
		GroupID:   groupID,
	}
}

func (e GroupDeleted) EventType() string { return "group.deleted" }
