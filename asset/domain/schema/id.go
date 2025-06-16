package schema

import (
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	FieldID     = id.FieldID
	WorkspaceID = id.WorkspaceID
	TagID       = id.TagID
	TagIDList   = id.TagIDList
	GroupID     = id.GroupID
	GroupIDList = id.GroupIDList
)

var (
	NewTagID          = id.NewTagID
	MustTagID         = id.MustTagID
	TagIDFrom         = id.TagIDFrom
	TagIDFromRef      = id.TagIDFromRef
	ErrInvalidTagID   = id.ErrInvalidID
	NewFieldID        = id.NewFieldID
	MustFieldID       = id.MustFieldID
	FieldIDFrom       = id.FieldIDFrom
	FieldIDFromRef    = id.FieldIDFromRef
	ErrInvalidFieldID = id.ErrInvalidID
)

type (
	ID        = id.SchemaID
	IDList    = id.SchemaIDList
	ProjectID = id.ProjectID
)

var (
	NewID        = id.NewSchemaID
	MustID       = id.MustSchemaID
	IDFrom       = id.SchemaIDFrom
	IDListFrom   = id.SchemaIDListFrom
	IDFromRef    = id.SchemaIDFromRef
	ErrInvalidID = id.ErrInvalidID
)

type FieldIDOrKey string
