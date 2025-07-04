package interfaces

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

type CreateFieldParam struct {
	ModelID      *id.ModelID
	Description  *string
	TypeProperty *schema.TypeProperty
	DefaultValue *value.Multiple
	Type         value.Type
	Name         string
	Key          string
	SchemaID     id.SchemaID
	Multiple     bool
	Unique       bool
	Required     bool
	IsTitle      bool
}

type UpdateFieldParam struct {
	ModelID      *id.ModelID
	Name         *string
	Description  *string
	Order        *int
	Key          *string
	Multiple     *bool
	Unique       *bool
	Required     *bool
	IsTitle      *bool
	TypeProperty *schema.TypeProperty
	DefaultValue *value.Multiple
	SchemaID     id.SchemaID
	FieldID      id.FieldID
}

type ModelData struct {
	ModelID   *id.ModelID
	SchemaID  id.SchemaID
	ProjectID id.ProjectID
}

var (
	ErrInvalidTypeProperty       = rerror.NewE(i18n.T("invalid type property"))
	ErrReferencedFiledKeyExists  = rerror.NewE(i18n.T("referenced field key exists"))
	ErrReferenceDirectionChanged = rerror.NewE(
		i18n.T("reference field direction can not be changed"),
	)
	ErrReferenceModelChanged = rerror.NewE(i18n.T("reference field model can not be changed"))
	ErrFieldNotFound         = rerror.NewE(i18n.T("field not found"))
	ErrInvalidValue          = rerror.NewE(i18n.T("invalid value"))
	ErrEitherModelOrGroup    = rerror.NewE(i18n.T("either model or group should be provided"))
)

type Schema interface {
	FindByID(context.Context, id.SchemaID, *usecase.Operator) (*schema.Schema, error)
	FindByIDs(context.Context, []id.SchemaID, *usecase.Operator) (schema.List, error)
	FindByModel(context.Context, id.ModelID, *usecase.Operator) (*schema.Package, error)
	FindByGroup(context.Context, id.GroupID, *usecase.Operator) (*schema.Schema, error)
	FindByGroups(context.Context, id.GroupIDList, *usecase.Operator) (schema.List, error)
	CreateField(context.Context, CreateFieldParam, *usecase.Operator) (*schema.Field, error)
	CreateFields(
		context.Context,
		id.SchemaID,
		[]CreateFieldParam,
		*usecase.Operator,
	) (schema.FieldList, error)
	UpdateField(context.Context, UpdateFieldParam, *usecase.Operator) (*schema.Field, error)
	UpdateFields(
		context.Context,
		id.SchemaID,
		[]UpdateFieldParam,
		*usecase.Operator,
	) (schema.FieldList, error)
	DeleteField(context.Context, id.SchemaID, id.FieldID, *usecase.Operator) error
	GetSchemasAndGroupSchemasByIDs(
		context.Context,
		id.SchemaIDList,
		*usecase.Operator,
	) (schema.List, schema.List, error)
}
