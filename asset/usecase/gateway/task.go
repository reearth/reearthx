package gateway

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/task"
)

type TaskRunner interface {
	Run(context.Context, task.Payload) error
	Retry(context.Context, string) error
}
