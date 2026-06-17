package authserver

import (
	"context"
	"testing"
	"time"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/oidc/pkg/oidc"
)

func newPostgresRepo(t *testing.T) *Postgres {
	t.Helper()
	pool := pgxtest.Connect(t)(t)
	r := NewPostgres(pgxx.NewClient(pool))
	require.NoError(t, r.Init(context.Background()))
	return r
}

func TestNewPostgres(t *testing.T) {
	c := pgxx.NewClient(nil)
	assert.Equal(t, &Postgres{client: c}, NewPostgres(c))
}

func TestPostgres_FindByID(t *testing.T) {
	r := newPostgresRepo(t)
	ctx := context.Background()
	id := NewRequestID()
	want := NewRequest().ID(id).MustBuild()

	got, err := r.FindByID(ctx, id)
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	_, err = r.client.DB(ctx).Exec(ctx, `INSERT INTO auth_requests (id) VALUES ($1)`, id.String())
	require.NoError(t, err)

	got, err = r.FindByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPostgres_FindByCode(t *testing.T) {
	r := newPostgresRepo(t)
	ctx := context.Background()
	want := NewRequest().NewID().Code("aaa").MustBuild()

	got, err := r.FindByCode(ctx, "aaa")
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	_, err = r.client.DB(ctx).Exec(ctx, `INSERT INTO auth_requests (id, code) VALUES ($1, $2)`, want.GetID(), "aaa")
	require.NoError(t, err)

	got, err = r.FindByCode(ctx, "aaa")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPostgres_FindBySubject(t *testing.T) {
	r := newPostgresRepo(t)
	ctx := context.Background()
	want := NewRequest().NewID().Subject("sss").MustBuild()

	got, err := r.FindBySubject(ctx, "sss")
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	_, err = r.client.DB(ctx).Exec(ctx, `INSERT INTO auth_requests (id, subject) VALUES ($1, $2)`, want.GetID(), "sss")
	require.NoError(t, err)

	got, err = r.FindBySubject(ctx, "sss")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPostgres_Save(t *testing.T) {
	r := newPostgresRepo(t)
	ctx := context.Background()
	aa := time.Now()
	req := NewRequest().NewID().
		ClientID("client").
		Subject("sub").
		Code("code").
		State("state").
		ResponseType("rt").
		Scopes([]string{"a", "openid"}).
		Audiences([]string{"aud"}).
		RedirectURI("ru").
		Nonce("nonce").
		CodeChallenge(&oidc.CodeChallenge{
			Challenge: "xxx",
			Method:    oidc.CodeChallengeMethodPlain,
		}).
		AuthorizedAt(&aa).
		MustBuild()

	require.NoError(t, r.Save(ctx, req))

	// Inspect the persisted row directly (mirrors the Mongo test's bson check).
	var (
		clientID, subject, code, state, responseType, redirectURI, nonce string
		scopesB, audiencesB, ccB                                         []byte
		authorizedAt                                                     *time.Time
	)
	err := r.client.DB(ctx).QueryRow(ctx,
		`SELECT client_id, subject, code, state, response_type, scopes, audiences, redirect_uri, nonce, code_challenge, authorized_at
		 FROM auth_requests WHERE id = $1`, req.GetID()).
		Scan(&clientID, &subject, &code, &state, &responseType, &scopesB, &audiencesB, &redirectURI, &nonce, &ccB, &authorizedAt)
	require.NoError(t, err)

	assert.Equal(t, "client", clientID)
	assert.Equal(t, "sub", subject)
	assert.Equal(t, "code", code)
	assert.Equal(t, "state", state)
	assert.Equal(t, "rt", responseType)
	assert.Equal(t, "ru", redirectURI)
	assert.Equal(t, "nonce", nonce)
	scopes, err := pgxx.UnmarshalJSONBSlice[string](scopesB)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "openid", "profile", "email"}, scopes) // essential scopes appended
	audiences, err := pgxx.UnmarshalJSONBSlice[string](audiencesB)
	require.NoError(t, err)
	assert.Equal(t, []string{"aud"}, audiences)
	assert.JSONEq(t, `{"challenge":"xxx","method":"plain"}`, string(ccB))
	require.NotNil(t, authorizedAt)
	assert.WithinDuration(t, aa, *authorizedAt, time.Millisecond)

	// Read path round-trips every field (compared via accessors so scope
	// augmentation matches on both sides).
	got, err := r.FindByID(ctx, req.ID())
	require.NoError(t, err)
	assert.Equal(t, req.GetClientID(), got.GetClientID())
	assert.Equal(t, req.GetSubject(), got.GetSubject())
	assert.Equal(t, req.GetCode(), got.GetCode())
	assert.Equal(t, req.GetState(), got.GetState())
	assert.Equal(t, req.GetResponseType(), got.GetResponseType())
	assert.Equal(t, req.GetScopes(), got.GetScopes())
	assert.Equal(t, req.GetAudience(), got.GetAudience())
	assert.Equal(t, req.GetRedirectURI(), got.GetRedirectURI())
	assert.Equal(t, req.GetNonce(), got.GetNonce())
	assert.Equal(t, req.GetCodeChallenge(), got.GetCodeChallenge())
	require.NotNil(t, got.AuthorizedAt())
	assert.WithinDuration(t, *req.AuthorizedAt(), *got.AuthorizedAt(), time.Millisecond)
}

func TestPostgres_Save_Upsert(t *testing.T) {
	r := newPostgresRepo(t)
	ctx := context.Background()
	id := NewRequestID()

	require.NoError(t, r.Save(ctx, NewRequest().ID(id).Code("first").MustBuild()))
	require.NoError(t, r.Save(ctx, NewRequest().ID(id).Code("second").MustBuild()))

	got, err := r.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "second", got.GetCode())

	// FindByCode reflects the updated value, and the stale code is gone.
	_, err = r.FindByCode(ctx, "first")
	assert.Same(t, rerror.ErrNotFound, err)
	got, err = r.FindByCode(ctx, "second")
	require.NoError(t, err)
	assert.Equal(t, id, got.ID())
}

func TestPostgres_Remove(t *testing.T) {
	r := newPostgresRepo(t)
	ctx := context.Background()
	id := NewRequestID()

	// Removing a missing row reports not-found, mirroring Mongo.
	assert.Same(t, rerror.ErrNotFound, r.Remove(ctx, id))

	require.NoError(t, r.Save(ctx, NewRequest().ID(id).MustBuild()))
	require.NoError(t, r.Remove(ctx, id))

	_, err := r.FindByID(ctx, id)
	assert.Same(t, rerror.ErrNotFound, err)
}
