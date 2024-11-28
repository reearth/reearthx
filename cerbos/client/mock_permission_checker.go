package client

import (
	"context"

	"github.com/reearth/reearthx/appx"
)

type MockPermissionChecker struct {
	Allow bool
	Error error
}

func NewMockPermissionChecker() *MockPermissionChecker {
	return &MockPermissionChecker{
		Allow: true,
		Error: nil,
	}
}

func (m *MockPermissionChecker) CheckPermission(ctx context.Context, authInfo *appx.AuthInfo, resource string, action string) (bool, error) {
	if m.Error != nil {
		return false, m.Error
	}
	return m.Allow, nil
}
