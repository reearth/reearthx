package user

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthFromAuth0Sub(t *testing.T) {
	tests := []struct {
		Name, Sub string
		Expected  Auth
	}{
		{
			Name: "with provider",
			Sub:  "xx|yy",
			Expected: Auth{
				Provider: "xx",
				Sub:      "xx|yy",
			},
		},
		{
			Name: "without provider",
			Sub:  "yy",
			Expected: Auth{
				Provider: "",
				Sub:      "yy",
			},
		},
		{
			Name:     "empty",
			Sub:      "",
			Expected: Auth{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expected, AuthFrom(tc.Sub))
		})
	}
}

func TestAuth_IsAuth0(t *testing.T) {
	tests := []struct {
		Name     string
		Auth     Auth
		Expected bool
	}{
		{
			Name: "is Auth",
			Auth: Auth{
				Provider: "auth0",
				Sub:      "xxx",
			},
			Expected: true,
		},
		{
			Name: "is not Auth",
			Auth: Auth{
				Provider: "foo",
				Sub:      "hoge",
			},
			Expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expected, tc.Auth.IsAuth0())
		})
	}
}

func TestReearthSub(t *testing.T) {
	uid := NewID()

	tests := []struct {
		name  string
		input string
		want  *Auth
	}{
		{
			name:  "should return reearth sub",
			input: uid.String(),
			want: &Auth{
				Provider: "reearth",
				Sub:      "reearth|" + uid.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReearthSub(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuth_Ref(t *testing.T) {
	type fields struct {
		Provider string
		Sub      string
	}
	tests := []struct {
		name   string
		fields fields
		want   *Auth
	}{
		{
			name: "ok",
			fields: fields{
				Provider: "auth0",
				Sub:      "xxx",
			},
			want: &Auth{
				Provider: "auth0",
				Sub:      "xxx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Auth{
				Provider: tt.fields.Provider,
				Sub:      tt.fields.Sub,
			}
			if got := a.Ref(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auth.Ref() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_String(t *testing.T) {
	type fields struct {
		Provider string
		Sub      string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ok",
			fields: fields{
				Provider: "auth0",
				Sub:      "xxx",
			},
			want: "xxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Auth{
				Provider: tt.fields.Provider,
				Sub:      tt.fields.Sub,
			}
			if got := a.String(); got != tt.want {
				t.Errorf("Auth.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_Has(t *testing.T) {
	type args struct {
		sub string
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want bool
	}{
		{
			name: "ok: true",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				sub: "xxx",
			},
			want: true,
		},
		{
			name: "ok: false",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				sub: "yyy",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Has(tt.args.sub); got != tt.want {
				t.Errorf("Auths.Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_HasProvider(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want bool
	}{
		{
			name: "ok: true",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				p: "auth0",
			},
			want: true,
		},
		{
			name: "ok: false",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				p: "yyy",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.HasProvider(tt.args.p); got != tt.want {
				t.Errorf("Auths.HasProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_GetByProvider(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want *Auth
	}{
		{
			name: "ok",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				p: "auth0",
			},
			want: &Auth{
				Provider: "auth0",
				Sub:      "xxx",
			},
		},
		{
			name: "ok: return nil",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				p: "yyy",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.GetByProvider(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auths.GetByProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_Get(t *testing.T) {
	type args struct {
		sub string
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want *Auth
	}{
		{
			name: "ok",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				sub: "xxx",
			},
			want: &Auth{
				Provider: "auth0",
				Sub:      "xxx",
			},
		},
		{
			name: "ok: return nil",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				sub: "yyy",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Get(tt.args.sub); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auths.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_Add(t *testing.T) {
	type args struct {
		u Auth
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want Auths
	}{
		{
			name: "ok",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				u: Auth{
					Provider: "foo",
					Sub:      "bar",
				},
			},
			want: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
				{
					Provider: "foo",
					Sub:      "bar",
				},
			},
		},
		{
			name: "ok: already exist",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				u: Auth{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			want: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Add(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auths.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_Remove(t *testing.T) {
	type args struct {
		sub string
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want Auths
	}{
		{
			name: "ok",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				sub: "xxx",
			},
			want: []Auth{},
		},
		{
			name: "ok: not exist",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				sub: "yyy",
			},
			want: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Remove(tt.args.sub); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auths.Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuths_RemoveByProvider(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		a    Auths
		args args
		want Auths
	}{
		{
			name: "ok",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				p: "auth0",
			},
			want: []Auth{},
		},
		{
			name: "ok: not exist",
			a: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
			args: args{
				p: "yyy",
			},
			want: []Auth{
				{
					Provider: "auth0",
					Sub:      "xxx",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.RemoveByProvider(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auths.RemoveByProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
