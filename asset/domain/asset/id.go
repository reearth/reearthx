package asset

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID            = id.AssetID
	IDList        = id.AssetIDList
	ProjectID     = id.ProjectID
	UserID        = accountdomain.UserID
	ThreadID      = id.ThreadID
	IntegrationID = id.IntegrationID
)

var (
	NewID            = id.NewAssetID
	NewProjectID     = id.NewProjectID
	NewUserID        = accountdomain.NewUserID
	NewThreadID      = id.NewThreadID
	NewIntegrationID = id.NewIntegrationID
)

var (
	MustID        = id.MustAssetID
	MustProjectID = id.MustProjectID
	MustUserID    = id.MustUserID
	MustThreadID  = id.MustThreadID
)

var (
	IDFrom        = id.AssetIDFrom
	ProjectIDFrom = id.ProjectIDFrom
	UserIDFrom    = accountdomain.UserIDFrom
	ThreadIDFrom  = id.ThreadIDFrom
)

var (
	IDFromRef        = id.AssetIDFromRef
	ProjectIDFromRef = id.ProjectIDFromRef
	UserIDFromRef    = accountdomain.UserIDFromRef
	ThreadIDFromRef  = id.ThreadIDFromRef
)

var ErrInvalidID = id.ErrInvalidID
