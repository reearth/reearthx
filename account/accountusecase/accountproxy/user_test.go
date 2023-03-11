package accountproxy

import "github.com/reearth/reearthx/account/accountusecase/accountinterfaces"

var _ accountinterfaces.User = (*User)(nil)
