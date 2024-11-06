package client

import (
	"context"
	"fmt"

	"github.com/reearth/reearthx/appx"
)

type ContextKey string

const (
	contextAuthInfo ContextKey = "authinfo"
)

type PermissionChecker struct {
	Service      string
	DashboardURL string
}

func NewPermissionChecker(service string, dashboardURL string) *PermissionChecker {
	return &PermissionChecker{
		Service:      service,
		DashboardURL: dashboardURL,
	}
}

func (p *PermissionChecker) CheckPermission(ctx context.Context, resource string, action string) (bool, error) {
	authInfo := getAuthInfo(ctx)
	if authInfo == nil {
		return false, fmt.Errorf("auth info not found")
	}

	input := CheckPermissionInput{
		Service:  p.Service,
		Resource: resource,
		Action:   action,
	}

	client := NewClient(p.DashboardURL)
	return client.CheckPermission(ctx, authInfo, input)
}

func getAuthInfo(ctx context.Context) *appx.AuthInfo {
	if v := ctx.Value(contextAuthInfo); v != nil {
		if v2, ok := v.(appx.AuthInfo); ok {
			return &v2
		}
	}
	return nil
}
