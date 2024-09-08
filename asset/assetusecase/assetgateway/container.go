package assetgateway

import (
	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
	"github.com/reearth/reearthx/mailer"
)

type Container struct {
	Authenticator accountgateway.Authenticator
	File          File
	Mailer        mailer.Mailer
	TaskRunner    TaskRunner
}
