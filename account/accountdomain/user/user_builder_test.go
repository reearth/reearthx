package user

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/idx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestBuilder_ID(t *testing.T) {
	uid := NewID()
	b := New().ID(uid).Name("aaa").Email("aaa@bbb.com").MustBuild()
	assert.Equal(t, uid, b.ID())
	assert.Nil(t, b.passwordReset)
}

func TestBuilder_NewID(t *testing.T) {
	b := New().NewID().Name("aaa").Email("aaa@bbb.com").MustBuild()
	assert.NotNil(t, b.ID())
}

func TestBuilder_ParseID(t *testing.T) {
	id := NewID()
	b := New().ParseID(id.String()).Name("aaa").Email("aaa@bbb.com").MustBuild()
	assert.Equal(t, id, b.ID())

	_, err := New().ParseID("invalid").Name("aaa").Email("aaa@bbb.com").Build()
	assert.Equal(t, idx.ErrInvalidID, err)
}

func TestBuilder_Name(t *testing.T) {
	b := New().NewID().Name("xxx").Email("aaa@bbb.com").MustBuild()
	assert.Equal(t, "xxx", b.Name())
}

func TestBuilder_DisplayName(t *testing.T) {
	b := New().NewID().Name("aaa").DisplayName("xxx").Email("aaa@bbb.com").MustBuild()
	assert.Equal(t, "xxx", b.DisplayName())
}

func TestBuilder_Workspace(t *testing.T) {
	tid := NewWorkspaceID()
	b := New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(tid).MustBuild()
	assert.Equal(t, tid, b.Workspace())
}

func TestBuilder_Auths(t *testing.T) {
	b := New().NewID().Name("aaa").Email("aaa@bbb.com").Auths([]Auth{
		{
			Provider: "xxx",
			Sub:      "aaa",
		},
	}).MustBuild()
	assert.Equal(t, Auths([]Auth{
		{
			Provider: "xxx",
			Sub:      "aaa",
		},
	}), b.Auths())
}

func TestBuilder_Email(t *testing.T) {
	b := New().NewID().Name("aaa").Email("xx@yy.zz").MustBuild()
	assert.Equal(t, "xx@yy.zz", b.Email())
}

func TestBuilder_Lang(t *testing.T) {
	l := language.Make("en")
	b := New().NewID().Name("aaa").Email("aaa@bbb.com").Lang(l).MustBuild()
	assert.Equal(t, l, b.Lang())
}

func TestBuilder_LangFrom(t *testing.T) {
	tests := []struct {
		Name, Lang string
		Expected   language.Tag
	}{
		{
			Name:     "success creating language",
			Lang:     "en",
			Expected: language.Make("en"),
		},
		{
			Name:     "empty language and empty tag",
			Lang:     "",
			Expected: language.Tag{},
		},
		{
			Name:     "empty tag of parse err",
			Lang:     "xxxxxxxxxxx",
			Expected: language.Tag{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			b := New().NewID().Name("aaa").Email("aaa@bbb.com").LangFrom(tc.Lang).MustBuild()
			assert.Equal(t, tc.Expected, b.Lang())
		})
	}
}

func TestNew(t *testing.T) {
	b := New()
	assert.NotNil(t, b)
	assert.IsType(t, &Builder{}, b)
}

func TestBuilder_Build(t *testing.T) {
	// bcrypt is not suitable for unit tests as it requires heavy computation
	DefaultPasswordEncoder = &NoopPasswordEncoder{}

	uid := NewID()
	wid := NewWorkspaceID()
	pass := MustEncodedPassword("abcDEF0!")

	type args struct {
		Name, Lang, Email string
		ID                ID
		Workspace         WorkspaceID
		Auths             []Auth
		PasswordBin       []byte
	}

	tests := []struct {
		Name     string
		Args     args
		Expected *User
		Err      error
	}{
		{
			Name: "Success build user",
			Args: args{
				Name:        "xxx",
				Email:       "xx@yy.zz",
				Lang:        "en",
				ID:          uid,
				Workspace:   wid,
				PasswordBin: pass,
				Auths: []Auth{
					{
						Provider: "ppp",
						Sub:      "sss",
					},
				},
			},
			Expected: &User{
				id:        uid,
				workspace: wid,
				email:     "xx@yy.zz",
				name:      "xxx",
				password:  pass,
				auths:     []Auth{{Provider: "ppp", Sub: "sss"}},
				lang:      language.English,
				theme:     ThemeDefault,
			},
		}, {
			Name:     "failed invalid id",
			Expected: nil,
			Err:      ErrInvalidID,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			res, err := New().
				ID(tt.Args.ID).
				EncodedPassword(pass).
				Name(tt.Args.Name).
				Auths(tt.Args.Auths).
				LangFrom(tt.Args.Lang).
				Email(tt.Args.Email).
				Workspace(tt.Args.Workspace).
				Build()
			if tt.Err == nil {
				assert.Equal(t, tt.Expected, res)
			} else {
				assert.Equal(t, tt.Err, err)
			}
		})
	}
}

func TestBuilder_MustBuild(t *testing.T) {
	// bcrypt is not suitable for unit tests as it requires heavy computation
	DefaultPasswordEncoder = &NoopPasswordEncoder{}

	uid := NewID()
	wid := NewWorkspaceID()
	pass := MustEncodedPassword("abcDEF0!")

	type args struct {
		Name, Lang, Email string
		ID                ID
		Workspace         WorkspaceID
		PasswordBin       []byte
		Auths             []Auth
	}

	tests := []struct {
		Name     string
		Args     args
		Expected *User
		Err      error
	}{
		{
			Name: "Success build user",
			Args: args{
				Name:        "xxx",
				Email:       "xx@yy.zz",
				Lang:        "en",
				ID:          uid,
				Workspace:   wid,
				PasswordBin: pass,
				Auths: []Auth{
					{
						Provider: "ppp",
						Sub:      "sss",
					},
				},
			},
			Expected: &User{
				id:        uid,
				workspace: wid,
				email:     "xx@yy.zz",
				name:      "xxx",
				password:  pass,
				auths:     []Auth{{Provider: "ppp", Sub: "sss"}},
				lang:      language.English,
				theme:     ThemeDefault,
			},
		}, {
			Name: "failed invalid id",
			Err:  ErrInvalidID,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			build := func() *User {
				t.Helper()
				return New().
					ID(tt.Args.ID).
					EncodedPassword(pass).
					Name(tt.Args.Name).
					Auths(tt.Args.Auths).
					LangFrom(tt.Args.Lang).
					Email(tt.Args.Email).
					Workspace(tt.Args.Workspace).
					MustBuild()
			}

			if tt.Err != nil {
				assert.PanicsWithValue(t, tt.Err, func() { _ = build() })
			} else {
				assert.Equal(t, tt.Expected, build())
			}
		})
	}
}

func TestBuilder_Verification(t *testing.T) {
	tests := []struct {
		name  string
		input *Verification
		want  *Builder
	}{
		{
			name: "should return verification",
			input: &Verification{
				verified:   true,
				code:       "xxx",
				expiration: time.Time{},
			},

			want: &Builder{
				u: &User{
					verification: &Verification{
						verified:   true,
						code:       "xxx",
						expiration: time.Time{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New()
			b.Verification(tt.input)
			assert.Equal(t, tt.want, b)
		})
	}
}
