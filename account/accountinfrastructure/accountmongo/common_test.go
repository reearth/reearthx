package accountmongo

import (
	"github.com/reearth/reearthx/mongox/mongotest"
)

func init() {
	mongotest.Env = "REEARTH_CMS_DB"
}
