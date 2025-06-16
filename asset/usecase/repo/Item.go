package repo

import (
	"context"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/reearth/reearthx/asset/domain/version"

	"github.com/reearth/reearthx/usecasex"
)

type FieldAndValue struct {
	Value *value.Multiple
	Field schema.FieldID
}

type CopyParams struct {
	Timestamp   time.Time
	User        *string
	Integration *string
	OldSchema   id.SchemaID
	NewSchema   id.SchemaID
	NewModel    id.ModelID
}

type Item interface {
	Filtered(ProjectFilter) Item
	FindByID(context.Context, id.ItemID, *version.Ref) (item.Versioned, error)
	FindByIDs(context.Context, id.ItemIDList, *version.Ref) (item.VersionedList, error)
	FindBySchema(
		context.Context,
		id.SchemaID,
		*version.Ref,
		*usecasex.Sort,
		*usecasex.Pagination,
	) (item.VersionedList, *usecasex.PageInfo, error)
	FindByModel(
		context.Context,
		id.ModelID,
		*version.Ref,
		*usecasex.Sort,
		*usecasex.Pagination,
	) (item.VersionedList, *usecasex.PageInfo, error)
	FindByAssets(context.Context, id.AssetIDList, *version.Ref) (item.VersionedList, error)
	LastModifiedByModel(context.Context, id.ModelID) (time.Time, error)
	Search(
		context.Context,
		schema.Package,
		*item.Query,
		*usecasex.Pagination,
	) (item.VersionedList, *usecasex.PageInfo, error)
	FindVersionByID(context.Context, id.ItemID, version.VersionOrRef) (item.Versioned, error)
	FindAllVersionsByID(context.Context, id.ItemID) (item.VersionedList, error)
	FindAllVersionsByIDs(context.Context, id.ItemIDList) (item.VersionedList, error)
	FindByModelAndValue(
		context.Context,
		id.ModelID,
		[]FieldAndValue,
		*version.Ref,
	) (item.VersionedList, error)
	IsArchived(context.Context, id.ItemID) (bool, error)
	Save(context.Context, *item.Item) error
	SaveAll(context.Context, item.List) error
	UpdateRef(context.Context, id.ItemID, version.Ref, *version.VersionOrRef) error
	Remove(context.Context, id.ItemID) error
	Archive(context.Context, id.ItemID, id.ProjectID, bool) error
	Copy(context.Context, CopyParams) (*string, *string, error)
}
