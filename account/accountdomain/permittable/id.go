// TODO: Delete this file once the permission check migration is complete.

package permittable

import "github.com/reearth/reearthx/account/accountdomain"

type ID = accountdomain.PermittableID

var NewID = accountdomain.NewPermittableID

var MustID = accountdomain.MustPermittableID

var IDFrom = accountdomain.PermittableIDFrom

var IDFromRef = accountdomain.PermittableIDFromRef

var ErrInvalidID = accountdomain.ErrInvalidID
