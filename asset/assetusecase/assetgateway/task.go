package assetgateway

import (
	"context"

	"github.com/reearth/reearthx/asset/assetdomain/task"
)

type TaskRunner interface {
	Run(context.Context, task.Payload) error
	Retry(context.Context, string) error
}
