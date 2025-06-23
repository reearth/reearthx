package asset

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/idx"
)

type ID = domain.AssetID
type IDList = domain.AssetIDList
type ProjectID = domain.ProjectID
type UserID = accountdomain.UserID
type ThreadID = domain.ThreadID
type IntegrationID = domain.IntegrationID

var NewID = domain.NewAssetID
var NewProjectID = domain.NewProjectID
var NewUserID = accountdomain.NewUserID
var NewThreadID = domain.NewThreadID
var NewIntegrationID = domain.NewIntegrationID

var MustID = domain.MustAssetID
var MustProjectID = domain.MustProjectID
var MustUserID = accountdomain.MustUserID
var MustThreadID = domain.MustThreadID

var IDFrom = domain.AssetIDFrom
var ProjectIDFrom = domain.ProjectIDFrom
var UserIDFrom = accountdomain.UserIDFrom
var ThreadIDFrom = domain.ThreadIDFrom

var IDFromRef = domain.AssetIDFromRef
var ProjectIDFromRef = domain.ProjectIDFromRef
var UserIDFromRef = accountdomain.UserIDFromRef
var ThreadIDFromRef = domain.ThreadIDFromRef

var ErrInvalidID = idx.ErrInvalidID
