package view

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID        = id.ViewID
	IDList    = id.ViewIDList
	UserID    = accountdomain.UserID
	ProjectID = id.ProjectID
	ModelID   = id.ModelID
	SchemaID  = id.SchemaID
)

var (
	NewID        = id.NewViewID
	NewProjectID = id.NewProjectID
	NewModelID   = id.NewModelID
	NewSchemaID  = id.NewSchemaID
	NewUserID    = accountdomain.NewUserID
)
