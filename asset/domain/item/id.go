package item

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID            = id.ItemID
	IDList        = id.ItemIDList
	ProjectID     = id.ProjectID
	SchemaID      = id.SchemaID
	FieldID       = id.FieldID
	FieldIDList   = id.FieldIDList
	ModelID       = id.ModelID
	ThreadID      = id.ThreadID
	UserID        = accountdomain.UserID
	IntegrationID = id.IntegrationID
	AssetID       = id.AssetID
	AssetIDList   = id.AssetIDList
	ItemGroupID   = id.ItemGroupID
)

var (
	NewID       = id.NewItemID
	NewThreadID = id.NewThreadID
)

var (
	MustID       = id.MustItemID
	MustThreadID = id.MustThreadID
)

var (
	IDFrom          = id.ItemIDFrom
	IDFromRef       = id.ItemIDFromRef
	ThreadIDFrom    = id.ThreadIDFrom
	ThreadIDFromRef = id.ThreadIDFromRef
)

var (
	NewFieldID     = id.NewFieldID
	MustFieldID    = id.MustFieldID
	FieldIDFrom    = id.FieldIDFrom
	FieldIDFromRef = id.FieldIDFromRef
)

var ErrInvalidID = id.ErrInvalidID
