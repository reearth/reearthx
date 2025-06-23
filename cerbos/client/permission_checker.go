package client

import (
	"context"
	"fmt"

	"github.com/reearth/reearthx/appx"
)

type PermissionChecker struct {
	Service      string
	DashboardURL string
}

type PermissionCheckerParams struct {
	Service      string
	DashboardURL string
}

func NewPermissionChecker(params PermissionCheckerParams) *PermissionChecker {
	return &PermissionChecker{
		Service:      params.Service,
		DashboardURL: params.DashboardURL,
	}
}

func (p *PermissionChecker) CheckPermission(ctx context.Context, authInfo *appx.AuthInfo, userId string, resource string, action string) (bool, error) {
	if p == nil {
		return false, fmt.Errorf("permission checker not found")
	}

	input := CheckPermissionInput{
		UserId:   userId,
		Service:  p.Service,
		Resource: resource,
		Action:   action,
	}

	client := NewClient(p.DashboardURL)
	return client.CheckPermission(ctx, authInfo, input)
}
