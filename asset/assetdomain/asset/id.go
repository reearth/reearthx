package asset

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/idx"
)

type ID = assetdomain.AssetID
type IDList = assetdomain.AssetIDList
type ProjectID = assetdomain.ProjectID
type UserID = accountdomain.UserID
type ThreadID = assetdomain.ThreadID
type IntegrationID = assetdomain.IntegrationID

var NewID = assetdomain.NewAssetID
var NewProjectID = assetdomain.NewProjectID
var NewUserID = accountdomain.NewUserID
var NewThreadID = assetdomain.NewThreadID
var NewIntegrationID = assetdomain.NewIntegrationID

var MustID = assetdomain.MustAssetID
var MustProjectID = assetdomain.MustProjectID
var MustUserID = accountdomain.MustUserID
var MustThreadID = assetdomain.MustThreadID

var IDFrom = assetdomain.AssetIDFrom
var ProjectIDFrom = assetdomain.ProjectIDFrom
var UserIDFrom = accountdomain.UserIDFrom
var ThreadIDFrom = assetdomain.ThreadIDFrom

var IDFromRef = assetdomain.AssetIDFromRef
var ProjectIDFromRef = assetdomain.ProjectIDFromRef
var UserIDFromRef = accountdomain.UserIDFromRef
var ThreadIDFromRef = assetdomain.ThreadIDFromRef

var ErrInvalidID = idx.ErrInvalidID
