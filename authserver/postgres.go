package authserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/rerror"
	"github.com/zitadel/oidc/pkg/oidc"
)

// Postgres is a RequestRepo backed by Postgres via pgxx. It mirrors Mongo: a
// single auth_requests table holds the OAuth2 authorization-request state, with
// the variable-length scopes/audiences/code_challenge persisted as jsonb.
type Postgres struct {
	client *pgxx.Client
}

var _ RequestRepo = (*Postgres)(nil)

func NewPostgres(client *pgxx.Client) *Postgres {
	return &Postgres{client: client}
}

// Init creates the auth_requests table and its lookup indexes if they do not
// exist. Like Mongo.Init (which ensures indexes), it lets the repo manage its
// own storage so a consumer need not wire the schema separately. Consumers that
// own their schema declaratively (e.g. via Atlas) may instead fold the
// equivalent DDL into their schema and skip this call.
func (r *Postgres) Init(ctx context.Context) error {
	db := r.client.DB(ctx)
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS auth_requests (
			id             text PRIMARY KEY,
			client_id      text NOT NULL DEFAULT '',
			subject        text NOT NULL DEFAULT '',
			code           text NOT NULL DEFAULT '',
			state          text NOT NULL DEFAULT '',
			response_type  text NOT NULL DEFAULT '',
			scopes         jsonb,
			audiences      jsonb,
			redirect_uri   text NOT NULL DEFAULT '',
			nonce          text NOT NULL DEFAULT '',
			code_challenge jsonb,
			authorized_at  timestamptz
		)`,
		`CREATE INDEX IF NOT EXISTS auth_requests_code_idx ON auth_requests (code)`,
		`CREATE INDEX IF NOT EXISTS auth_requests_subject_idx ON auth_requests (subject)`,
	} {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return rerror.ErrInternalByWithContext(ctx, pgxx.WrapError(err))
		}
	}
	return nil
}

func (r *Postgres) FindByID(ctx context.Context, id RequestID) (*Request, error) {
	return r.findOne(ctx, "id = $1", id.String())
}

func (r *Postgres) FindByCode(ctx context.Context, s string) (*Request, error) {
	return r.findOne(ctx, "code = $1", s)
}

func (r *Postgres) FindBySubject(ctx context.Context, s string) (*Request, error) {
	return r.findOne(ctx, "subject = $1", s)
}

func (r *Postgres) Save(ctx context.Context, request *Request) error {
	d, err := newPostgresDocument(request)
	if err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	const q = `INSERT INTO auth_requests (
		id, client_id, subject, code, state, response_type,
		scopes, audiences, redirect_uri, nonce, code_challenge, authorized_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	ON CONFLICT (id) DO UPDATE SET
		client_id      = EXCLUDED.client_id,
		subject        = EXCLUDED.subject,
		code           = EXCLUDED.code,
		state          = EXCLUDED.state,
		response_type  = EXCLUDED.response_type,
		scopes         = EXCLUDED.scopes,
		audiences      = EXCLUDED.audiences,
		redirect_uri   = EXCLUDED.redirect_uri,
		nonce          = EXCLUDED.nonce,
		code_challenge = EXCLUDED.code_challenge,
		authorized_at  = EXCLUDED.authorized_at`
	if _, err := r.client.DB(ctx).Exec(ctx, q,
		d.ID, d.ClientID, d.Subject, d.Code, d.State, d.ResponseType,
		d.Scopes, d.Audiences, d.RedirectURI, d.Nonce, d.CodeChallenge, d.AuthorizedAt,
	); err != nil {
		return rerror.ErrInternalByWithContext(ctx, pgxx.WrapError(err))
	}
	return nil
}

// Remove deletes the request, returning rerror.ErrNotFound when no row matched —
// mirroring Mongo.RemoveOne so the two backends behave identically.
func (r *Postgres) Remove(ctx context.Context, id RequestID) error {
	tag, err := r.client.DB(ctx).Exec(ctx, `DELETE FROM auth_requests WHERE id = $1`, id.String())
	if err != nil {
		return rerror.ErrInternalByWithContext(ctx, pgxx.WrapError(err))
	}
	if tag.RowsAffected() == 0 {
		return rerror.ErrNotFound
	}
	return nil
}

func (r *Postgres) findOne(ctx context.Context, where string, arg any) (*Request, error) {
	const cols = `id, client_id, subject, code, state, response_type,
		scopes, audiences, redirect_uri, nonce, code_challenge, authorized_at`
	row := r.client.DB(ctx).QueryRow(ctx, `SELECT `+cols+` FROM auth_requests WHERE `+where, arg)

	var d postgresDocument
	if err := row.Scan(
		&d.ID, &d.ClientID, &d.Subject, &d.Code, &d.State, &d.ResponseType,
		&d.Scopes, &d.Audiences, &d.RedirectURI, &d.Nonce, &d.CodeChallenge, &d.AuthorizedAt,
	); err != nil {
		return nil, pgxx.MapError(err)
	}
	return d.Model()
}

// postgresDocument is the row shape. scopes/audiences/code_challenge are jsonb,
// scanned as raw bytes and decoded in Model() to keep the SQL types simple.
type postgresDocument struct {
	ID            string
	ClientID      string
	Subject       string
	Code          string
	State         string
	ResponseType  string
	Scopes        []byte
	Audiences     []byte
	RedirectURI   string
	Nonce         string
	CodeChallenge []byte
	AuthorizedAt  *time.Time
}

type postgresCodeChallenge struct {
	Challenge string `json:"challenge"`
	Method    string `json:"method"`
}

func newPostgresDocument(req *Request) (*postgresDocument, error) {
	if req == nil {
		return nil, nil
	}
	scopes, err := pgxx.MarshalJSONBSlice(req.GetScopes())
	if err != nil {
		return nil, err
	}
	audiences, err := pgxx.MarshalJSONBSlice(req.GetAudience())
	if err != nil {
		return nil, err
	}
	var cc []byte
	if c := req.GetCodeChallenge(); c != nil {
		if cc, err = json.Marshal(postgresCodeChallenge{
			Challenge: c.Challenge,
			Method:    string(c.Method),
		}); err != nil {
			return nil, err
		}
	}
	return &postgresDocument{
		ID:            req.GetID(),
		ClientID:      req.GetClientID(),
		Subject:       req.GetSubject(),
		Code:          req.GetCode(),
		State:         req.GetState(),
		ResponseType:  string(req.GetResponseType()),
		Scopes:        scopes,
		Audiences:     audiences,
		RedirectURI:   req.GetRedirectURI(),
		Nonce:         req.GetNonce(),
		CodeChallenge: cc,
		AuthorizedAt:  req.AuthorizedAt(),
	}, nil
}

func (d *postgresDocument) Model() (*Request, error) {
	if d == nil {
		return nil, nil
	}
	id, err := RequestIDFrom(d.ID)
	if err != nil {
		return nil, err
	}
	scopes, err := pgxx.UnmarshalJSONBSlice[string](d.Scopes)
	if err != nil {
		return nil, err
	}
	audiences, err := pgxx.UnmarshalJSONBSlice[string](d.Audiences)
	if err != nil {
		return nil, err
	}
	var cc *oidc.CodeChallenge
	if len(d.CodeChallenge) > 0 {
		var pc postgresCodeChallenge
		if err := json.Unmarshal(d.CodeChallenge, &pc); err != nil {
			return nil, err
		}
		cc = &oidc.CodeChallenge{
			Challenge: pc.Challenge,
			Method:    oidc.CodeChallengeMethod(pc.Method),
		}
	}
	return NewRequest().
		ID(id).
		ClientID(d.ClientID).
		Subject(d.Subject).
		Code(d.Code).
		State(d.State).
		ResponseType(oidc.ResponseType(d.ResponseType)).
		Scopes(scopes).
		Audiences(audiences).
		RedirectURI(d.RedirectURI).
		Nonce(d.Nonce).
		CodeChallenge(cc).
		AuthorizedAt(d.AuthorizedAt).
		Build()
}
