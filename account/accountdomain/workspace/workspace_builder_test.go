package workspace

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceBuilder_ID(t *testing.T) {
	wid := NewID()
	b := NewWorkspace().ID(wid)
	assert.Equal(t, wid, b.w.id)
}

func TestWorkspaceBuilder_Members(t *testing.T) {
	m := map[UserID]Member{NewUserID(): {Role: RoleOwner}}
	b := NewWorkspace().Members(m)
	assert.Equal(t, m, b.members)
}

func TestWorkspaceBuilder_Name(t *testing.T) {
	w := NewWorkspace().Name("xxx")
	assert.Equal(t, "xxx", w.w.name)
}

func TestWorkspaceBuilder_NewID(t *testing.T) {
	b := NewWorkspace().NewID()
	assert.False(t, b.w.id.IsEmpty())
}

func TestWorkspaceBuilder_Build(t *testing.T) {
	m := map[UserID]Member{NewUserID(): {Role: RoleOwner}}
	i := map[IntegrationID]Member{NewIntegrationID(): {Role: RoleOwner}}
	id := NewID()
	w, err := NewWorkspace().ID(id).Name("a").Integrations(i).Members(m).Build()
	assert.NoError(t, err)

	assert.Equal(t, &Workspace{
		id:      id,
		name:    "a",
		members: NewMembersWith(m, i),
	}, w)

	w, err = NewWorkspace().Build()
	assert.Equal(t, ErrInvalidID, err)
	assert.Nil(t, w)
}

func TestWorkspaceBuilder_MustBuild(t *testing.T) {
	m := map[UserID]Member{NewUserID(): {Role: RoleOwner}}
	i := map[IntegrationID]Member{NewIntegrationID(): {Role: RoleOwner}}
	id := NewID()
	w := NewWorkspace().ID(id).Name("a").Integrations(i).Members(m).MustBuild()

	assert.Equal(t, &Workspace{
		id:      id,
		name:    "a",
		members: NewMembersWith(m, i),
	}, w)

	//expect panic
	defer func() { recover() }() //nolint:errcheck //test code

	w = NewWorkspace().MustBuild()
	assert.Nil(t, w)
	t.Errorf("expect: panic but not happened")
}

func TestWorkspaceBuilder_Integrations(t *testing.T) {
	type fields struct {
		w            *Workspace
		members      map[UserID]Member
		integrations map[IntegrationID]Member
		personal     bool
	}
	type args struct {
		integrations map[IntegrationID]Member
	}
	integrations := map[IntegrationID]Member{NewIntegrationID(): {Role: RoleOwner}}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *WorkspaceBuilder
	}{
		{
			name: "ok",
			fields: fields{
				w:            &Workspace{},
				members:      nil,
				integrations: nil,
				personal:     false,
			},
			args: args{
				integrations: integrations,
			},
			want: &WorkspaceBuilder{
				w:            &Workspace{},
				members:      nil,
				integrations: integrations,
				personal:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &WorkspaceBuilder{
				w:            tt.fields.w,
				members:      tt.fields.members,
				integrations: tt.fields.integrations,
				personal:     tt.fields.personal,
			}
			if got := b.Integrations(tt.args.integrations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WorkspaceBuilder.Integrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkspaceBuilder_Personal(t *testing.T) {
	type fields struct {
		w            *Workspace
		members      map[UserID]Member
		integrations map[IntegrationID]Member
		personal     bool
	}
	type args struct {
		p bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *WorkspaceBuilder
	}{
		{
			name: "ok",
			fields: fields{
				w:            &Workspace{},
				members:      nil,
				integrations: nil,
				personal:     false,
			},
			args: args{
				p: true,
			},
			want: &WorkspaceBuilder{
				w:            &Workspace{},
				members:      nil,
				integrations: nil,
				personal:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &WorkspaceBuilder{
				w:            tt.fields.w,
				members:      tt.fields.members,
				integrations: tt.fields.integrations,
				personal:     tt.fields.personal,
			}
			if got := b.Personal(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WorkspaceBuilder.Personal() = %v, want %v", got, tt.want)
			}
		})
	}
}
