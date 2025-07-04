package group

import (
	"fmt"
	"testing"

	"github.com/reearth/reearthx/asset/domain/id"

	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
)

func TestGroup_Clone(t *testing.T) {
	mId := NewID()
	pId := id.NewProjectID()
	sId := id.NewSchemaID()
	tests := []struct {
		name  string
		group *Group
	}{
		{
			name: "test",
			group: &Group{
				id:          mId,
				project:     pId,
				schema:      sId,
				name:        "n1",
				description: "d1",
				key:         id.NewKey("123456"),
				order:       1,
			},
		},
		{
			name:  "test",
			group: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := tt.group.Clone()
			if tt.group == nil {
				assert.Nil(t, c)
				return
			}
			assert.Equal(t, tt.group, c)
			assert.NotSame(t, tt.group, c)
		})
	}
}

func TestGroup_Description(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		want  string
	}{
		{
			name: "test",
			group: Group{
				description: "d1",
			},
			want: "d1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.group.Description())
		})
	}
}

func TestGroup_ID(t *testing.T) {
	mId := NewID()
	tests := []struct {
		name  string
		group Group
		want  ID
	}{
		{
			name: "test",
			group: Group{
				id: mId,
			},
			want: mId,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.group.ID())
		})
	}
}

func TestGroup_Key(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		want  id.Key
	}{
		{
			name: "test",
			group: Group{
				key: id.NewKey("123456"),
			},
			want: id.NewKey("123456"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.group.Key())
		})
	}
}

func TestGroup_Name(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		want  string
	}{
		{
			name: "test",
			group: Group{
				name: "n1",
			},
			want: "n1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.group.Name())
		})
	}
}

func TestGroup_Project(t *testing.T) {
	pId := id.NewProjectID()
	tests := []struct {
		name  string
		group Group
		want  ProjectID
	}{
		{
			name: "test",
			group: Group{
				project: pId,
			},
			want: pId,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.group.Project(), "Project()")
		})
	}
}

func TestGroup_Schema(t *testing.T) {
	sId := id.NewSchemaID()
	tests := []struct {
		name  string
		group Group
		want  SchemaID
	}{
		{
			name: "test",
			group: Group{
				schema: sId,
			},
			want: sId,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.group.Schema())
		})
	}
}

func TestGroup_SetDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		want Group
		args args
	}{
		{
			name: "test",
			args: args{
				description: "d1",
			},
			want: Group{
				description: "d1",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := Group{}
			m.SetDescription(tt.args.description)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestGroup_SetKey(t *testing.T) {
	type args struct {
		key id.Key
	}
	tests := []struct {
		name    string
		args    args
		want    Group
		wantErr error
	}{
		{
			name:    "pass",
			args:    args{key: id.NewKey("123456")},
			want:    Group{key: id.NewKey("123456")},
			wantErr: nil,
		},
		{
			name: "fail",
			args: args{key: id.NewKey("a")},
			want: Group{},
			wantErr: &rerror.Error{
				Label: id.ErrInvalidKey,
				Err:   fmt.Errorf("%s", "a"),
			},
		},
		// {
		// 	name: "fail 2",
		// 	args: args{key: id.NewKey("_aaaaaaaa")},
		// 	want: Group{},
		// 	wantErr: &rerror.Error{
		// 		Label: id.ErrInvalidKey,
		// 		Err:   fmt.Errorf("%s", "_aaaaaaaa"),
		// 	},
		// },
		// {
		// 	name: "fail 3",
		// 	args: args{key: id.NewKey("-aaaaaaaa")},
		// 	want: Group{},
		// 	wantErr: &rerror.Error{
		// 		Label: id.ErrInvalidKey,
		// 		Err:   fmt.Errorf("%s", "-aaaaaaaa"),
		// 	},
		// },
		{
			name: "empty",
			args: args{key: id.NewKey("")},
			want: Group{},
			wantErr: &rerror.Error{
				Label: id.ErrInvalidKey,
				Err:   fmt.Errorf("%s", ""),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := Group{}
			err := m.SetKey(tt.args.key)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestGroup_SetName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		want Group
		args args
	}{
		{
			name: "test",
			args: args{
				name: "n1",
			},
			want: Group{
				name: "n1",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := Group{}
			m.SetName(tt.args.name)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestGroup_SetOrder(t *testing.T) {
	g := &Group{order: 3}
	g.SetOrder(1)
	assert.Equal(t, 1, g.Order())
}
