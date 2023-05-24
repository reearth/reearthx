package accountproxy

import (
	"errors"
	"reflect"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
)

func TestUserByIDsResponseTo(t *testing.T) {
	uid := accountdomain.NewUserID()
	u := &UserByIDsNodesUser{
		Id:       uid.String(),
		Name:     "name",
		Email:    "email",
		Typename: "User",
	}

	type args struct {
		r   *UserByIDsResponse
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    []*user.Simple
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				&UserByIDsResponse{
					[]UserByIDsNodesNode{
						u,
					},
				},
				nil,
			},
			want: []*user.Simple{
				{
					ID:    uid,
					Name:  "name",
					Email: "email",
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				nil,
				errors.New("test"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserByIDsResponseTo(tt.args.r, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserByIDsResponseTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserByIDsResponseTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeToUser(t *testing.T) {
	uid := accountdomain.NewUserID()
	wid := accountdomain.NewWorkspaceID()
	type args struct {
		me FragmentMe
	}
	tests := []struct {
		name    string
		args    args
		want    *user.User
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				FragmentMe{
					Id:            uid.String(),
					Name:          "name",
					Email:         "test@exmple.com",
					Lang:          "ja",
					Theme:         "dark",
					MyWorkspaceId: wid.String(),
					Auths:         []string{"foo|bar"},
				},
			},
			want: user.New().ID(uid).Name("name").
				Email("test@exmple.com").LangFrom("ja").Theme(user.ThemeDark).
				Workspace(wid).Auths([]user.Auth{{Provider: "foo", Sub: "foo|bar"}}).MustBuild(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MeToUser(tt.args.me)
			if (err != nil) != tt.wantErr {
				t.Errorf("MeToUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MeToUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserByIDsNodesNodeTo(t *testing.T) {
	uid := accountdomain.NewUserID()
	u := &UserByIDsNodesUser{
		Id:       uid.String(),
		Name:     "name",
		Email:    "email",
		Typename: "User",
	}
	type args struct {
		r UserByIDsNodesNode
	}
	tests := []struct {
		name    string
		args    args
		want    *user.Simple
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				r: u,
			},
			want: &user.Simple{
				ID:    uid,
				Name:  "name",
				Email: "email",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserByIDsNodesNodeTo(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserByIDsNodesNodeTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserByIDsNodesNodeTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserByIDsNodesUserTo(t *testing.T) {
	uid := accountdomain.NewUserID()
	u := &UserByIDsNodesUser{
		Id:       uid.String(),
		Name:     "name",
		Email:    "email",
		Typename: "User",
	}
	type args struct {
		r *UserByIDsNodesUser
	}
	tests := []struct {
		name    string
		args    args
		want    *user.Simple
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				r: u,
			},
			want: &user.Simple{
				ID:    uid,
				Name:  "name",
				Email: "email",
			},
			wantErr: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserByIDsNodesUserTo(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserByIDsNodesUserTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserByIDsNodesUserTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
