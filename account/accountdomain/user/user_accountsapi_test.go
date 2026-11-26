package user

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/stretchr/testify/assert"
)

func TestUser_AccountsAPIFields(t *testing.T) {
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	deleted := created.Add(time.Hour)
	logout := created.Add(2 * time.Hour)

	u := New().NewID().Email("a@example.com").Workspace(accountdomain.NewWorkspaceID()).
		CreatedAt(&created).DeletedAt(&deleted).LatestLogoutAt(logout).MustBuild()

	// WithHost (used on every read) goes through Clone, which must keep the fields.
	for _, got := range []*User{u, u.Clone(), u.WithHost("h")} {
		assert.Equal(t, &created, got.CreatedAt())
		assert.Equal(t, &deleted, got.DeletedAt())
		assert.Equal(t, logout, got.LatestLogoutAt())
	}

	assert.Nil(t, New().NewID().Email("a@example.com").MustBuild().DeletedAt())
}
