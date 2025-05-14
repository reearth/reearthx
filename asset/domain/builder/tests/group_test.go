package builder_test

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/builder"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/stretchr/testify/assert"
)

func TestGroupBuilder_Build(t *testing.T) {
	now := time.Now()
	groupID := id.NewGroupID()

	tests := []struct {
		name    string
		builder func() *builder.GroupBuilder
		want    *entity.Group
		wantErr error
	}{
		{
			name: "success",
			builder: func() *builder.GroupBuilder {
				return builder.NewGroupBuilder().
					CreatedAt(now).
					ID(groupID).
					Name("test-group").
					Policy("test-policy").
					Description("test description")
			},
			want: func() *entity.Group {
				group := entity.NewGroup(groupID, "test-group")
				err := group.UpdatePolicy("test-policy")
				if err != nil {
					panic(err)
				}
				err = group.UpdateDescription("test description")
				if err != nil {
					panic(err)
				}
				return group
			}(),
			wantErr: nil,
		},
		{
			name: "missing ID",
			builder: func() *builder.GroupBuilder {
				return builder.NewGroupBuilder().
					Name("test-group").
					Policy("test-policy")
			},
			wantErr: id.ErrInvalidID,
		},
		{
			name: "missing name",
			builder: func() *builder.GroupBuilder {
				return builder.NewGroupBuilder().
					ID(groupID).
					Policy("test-policy")
			},
			wantErr: domain.ErrEmptyGroupName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.builder().Build()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.want != nil {
				assert.Equal(t, tt.want.ID(), got.ID())
				assert.Equal(t, tt.want.Name(), got.Name())
				assert.Equal(t, tt.want.Policy(), got.Policy())
				assert.Equal(t, tt.want.Description(), got.Description())
			}
		})
	}
}

func TestGroupBuilder_MustBuild(t *testing.T) {
	groupID := id.NewGroupID()

	// Test successful build
	assert.NotPanics(t, func() {
		group := builder.NewGroupBuilder().
			CreatedAt(time.Now()).
			ID(groupID).
			Name("test-group").
			Policy("test-policy").
			Description("test description").
			MustBuild()
		assert.NotNil(t, group)
	})

	// Test panic on invalid build
	assert.Panics(t, func() {
		builder.NewGroupBuilder().MustBuild()
	})
}

func TestGroupBuilder_Setters(t *testing.T) {
	groupID := id.NewGroupID()
	now := time.Now()

	b := builder.NewGroupBuilder().
		CreatedAt(now).
		ID(groupID).
		Name("test-group").
		Policy("test-policy").
		Description("test description")

	group, err := b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, group)

	assert.Equal(t, groupID, group.ID())
	assert.Equal(t, "test-group", group.Name())
	assert.Equal(t, "test-policy", group.Policy())
	assert.Equal(t, "test description", group.Description())
	assert.Equal(t, now.Unix(), group.CreatedAt().Unix())
}

func TestGroupBuilder_NewID(t *testing.T) {
	b := builder.NewGroupBuilder().NewID()
	// Add required fields to make the build succeed
	b = b.Name("test-group")

	group, err := b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.NotEqual(t, id.GroupID{}, group.ID()) // ID should be set
}

func TestGroupBuilder_InvalidSetters(t *testing.T) {
	groupID := id.NewGroupID()

	// Test setting empty name
	b := builder.NewGroupBuilder().
		ID(groupID).
		Name("")
	group, err := b.Build()
	assert.Equal(t, domain.ErrEmptyGroupName, err)
	assert.Nil(t, group)

	// Test setting empty policy
	b = builder.NewGroupBuilder().
		ID(groupID).
		Name("test-group").
		Policy("")
	group, err = b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, group)
}
