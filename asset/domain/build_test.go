package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	b := New()
	assert.NotNil(t, b)
	assert.NotNil(t, b.a)
}

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	id := NewID()
	wid := NewWorkspaceID()
	gid := NewGroupID()
	pid := NewProjectID()

	tests := []struct {
		name    string
		build   func() *Builder
		want    *Asset
		wantErr error
	}{
		{
			name: "success",
			build: func() *Builder {
				return New().
					ID(id).
					WorkspaceID(wid).
					GroupID(gid).
					ProjectID(pid).
					Name("test.txt").
					Size(100).
					URL("https://example.com/test.txt").
					ContentType("text/plain").
					Status(StatusActive).
					Error("").
					CreatedAt(now).
					UpdatedAt(now)
			},
			want: &Asset{
				id:          id,
				workspaceID: wid,
				groupID:     gid,
				projectID:   pid,
				name:        "test.txt",
				size:        100,
				url:         "https://example.com/test.txt",
				contentType: "text/plain",
				status:      StatusActive,
				error:       "",
				createdAt:   now,
				updatedAt:   now,
			},
		},
		{
			name: "success with defaults",
			build: func() *Builder {
				return New().
					ID(id).
					WorkspaceID(wid).
					URL("https://example.com/test.txt").
					Size(100)
			},
			want: &Asset{
				id:          id,
				workspaceID: wid,
				url:         "https://example.com/test.txt",
				size:        100,
				status:      StatusPending,
			},
		},
		{
			name: "error invalid id",
			build: func() *Builder {
				return New().
					WorkspaceID(wid).
					URL("https://example.com/test.txt").
					Size(100)
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "error empty workspace id",
			build: func() *Builder {
				return New().
					ID(id).
					URL("https://example.com/test.txt").
					Size(100)
			},
			wantErr: ErrEmptyWorkspaceID,
		},
		{
			name: "error empty url",
			build: func() *Builder {
				return New().
					ID(id).
					WorkspaceID(wid).
					Size(100)
			},
			wantErr: ErrEmptyURL,
		},
		{
			name: "error invalid size",
			build: func() *Builder {
				return New().
					ID(id).
					WorkspaceID(wid).
					URL("https://example.com/test.txt").
					Size(0)
			},
			wantErr: ErrEmptySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.build().Build()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				// For tests with default timestamps, we need to check if they're set
				if tt.want.createdAt.IsZero() {
					assert.False(t, got.createdAt.IsZero())
					assert.False(t, got.updatedAt.IsZero())
					// Copy the generated timestamps to the expected struct for full comparison
					tt.want.createdAt = got.createdAt
					tt.want.updatedAt = got.updatedAt
				}
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuilder_MustBuild(t *testing.T) {
	id := NewID()
	wid := NewWorkspaceID()

	tests := []struct {
		name      string
		build     func() *Builder
		want      *Asset
		wantPanic error
	}{
		{
			name: "success",
			build: func() *Builder {
				return New().
					ID(id).
					WorkspaceID(wid).
					URL("https://example.com/test.txt").
					Size(100)
			},
			want: &Asset{
				id:          id,
				workspaceID: wid,
				url:         "https://example.com/test.txt",
				size:        100,
				status:      StatusPending,
			},
		},
		{
			name: "panic on invalid id",
			build: func() *Builder {
				return New().
					WorkspaceID(wid).
					URL("https://example.com/test.txt").
					Size(100)
			},
			wantPanic: ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic != nil {
				assert.PanicsWithValue(t, tt.wantPanic, func() {
					//nolint:errcheck // MustBuild panics on error, return value is intentionally not checked
					tt.build().MustBuild()
				})
			} else {
				got := tt.build().MustBuild()
				if tt.want.createdAt.IsZero() {
					assert.False(t, got.createdAt.IsZero())
					assert.False(t, got.updatedAt.IsZero())
					tt.want.createdAt = got.createdAt
					tt.want.updatedAt = got.updatedAt
				}
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuilder_NewID(t *testing.T) {
	b := New().NewID()
	assert.NotNil(t, b.a.id)
	assert.False(t, b.a.id.IsNil())
}

func TestBuilder_Setters(t *testing.T) {
	now := time.Now()
	id := NewID()
	wid := NewWorkspaceID()
	gid := NewGroupID()
	pid := NewProjectID()

	tests := []struct {
		name  string
		build func() *Builder
		check func(*testing.T, *Builder)
	}{
		{
			name: "ID",
			build: func() *Builder {
				return New().ID(id)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, id, b.a.id)
			},
		},
		{
			name: "WorkspaceID",
			build: func() *Builder {
				return New().WorkspaceID(wid)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, wid, b.a.workspaceID)
			},
		},
		{
			name: "GroupID",
			build: func() *Builder {
				return New().GroupID(gid)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, gid, b.a.groupID)
			},
		},
		{
			name: "ProjectID",
			build: func() *Builder {
				return New().ProjectID(pid)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, pid, b.a.projectID)
			},
		},
		{
			name: "Name",
			build: func() *Builder {
				return New().Name("test.txt")
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, "test.txt", b.a.name)
			},
		},
		{
			name: "Size",
			build: func() *Builder {
				return New().Size(100)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, int64(100), b.a.size)
			},
		},
		{
			name: "URL",
			build: func() *Builder {
				return New().URL("https://example.com/test.txt")
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, "https://example.com/test.txt", b.a.url)
			},
		},
		{
			name: "ContentType",
			build: func() *Builder {
				return New().ContentType("text/plain")
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, "text/plain", b.a.contentType)
			},
		},
		{
			name: "Status",
			build: func() *Builder {
				return New().Status(StatusActive)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, StatusActive, b.a.status)
			},
		},
		{
			name: "Error",
			build: func() *Builder {
				return New().Error("test error")
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, "test error", b.a.error)
			},
		},
		{
			name: "CreatedAt",
			build: func() *Builder {
				return New().CreatedAt(now)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, now, b.a.createdAt)
			},
		},
		{
			name: "UpdatedAt",
			build: func() *Builder {
				return New().UpdatedAt(now)
			},
			check: func(t *testing.T, b *Builder) {
				assert.Equal(t, now, b.a.updatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.build()
			tt.check(t, b)
		})
	}
}
