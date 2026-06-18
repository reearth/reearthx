package item

import (
	"github.com/reearth/mongogit/version"
	"github.com/reearth/reearthx/util"
)

func init() {
	// Route the version library's time source through reearthx's clock so
	// util.MockNow controls version timestamps, including in tests.
	version.Now = util.Now
}
