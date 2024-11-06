package client

import (
	"context"

	"github.com/reearth/reearthx/appx"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

var (
	errOperationDenied error = rerror.NewE(i18n.T("operation denied"))
)

type PermissionService interface {
	CheckPermission(ctx context.Context, authInfo *appx.AuthInfo, resource string, action string) (bool, error)
}

func checkPermissionClient(client any) (PermissionService, bool) {
	if client == nil {
		return nil, false
	}

	adapter, ok := client.(PermissionService)
	if !ok || adapter == nil {
		return nil, false
	}
	return adapter, true
}

func CheckPermission(ctx context.Context, authInfo *appx.AuthInfo, client any, resource string, action string) (bool, error) {
	checkPermissionAdapter, ok := checkPermissionClient(client)
	if !ok {
		return false, errOperationDenied
	}

	return checkPermissionAdapter.CheckPermission(ctx, authInfo, resource, action)
}
