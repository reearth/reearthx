package accountgateway

import "github.com/reearth/reearthx/mailer"

type Container struct {
	Authenticator Authenticator
	Mailer        mailer.Mailer
}
