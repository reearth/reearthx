package authserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/zitadel/oidc/pkg/oidc"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestEndpoint(t *testing.T) {
	e := echo.New()
	cr := &configRepo{}
	rr := &requestRepo{}

	Endpoint(context.Background(), EndpointConfig{
		Issuer:          "https://example.com/",
		URL:             lo.Must(url.Parse("https://example.com")),
		WebURL:          lo.Must(url.Parse("https://web.example.com")),
		Key:             "",
		Dev:             false,
		DefaultClientID: "default-client",
		UserRepo:        &userRepo{},
		ConfigRepo:      cr,
		RequestRepo:     rr,
	}, e.Group(""))

	ts := httptest.NewServer(e)
	defer ts.Close()

	// step 1
	verifier, challenge := randomCodeChallenge()
	res := send(http.MethodGet, ts.URL+"/authorize", false, map[string]string{
		"response_type":         "code",
		"client_id":             "default-client",
		"redirect_uri":          "https://web.example.com",
		"scope":                 "openid email profile offline_access",
		"state":                 "hogestate",
		"code_challenge":        challenge,
		"code_challenge_method": "S256",
	}, nil)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	loc := res.Header.Get("Location")
	assert.Contains(t, loc, "https://web.example.com/login?id=")
	reqID := lo.Must(url.Parse(loc)).Query().Get("id")

	// step 2
	res = send(http.MethodPost, ts.URL+"/api/login", true, map[string]string{
		"username": "aaa@example.com",
		"password": "aaa_",
		"id":       reqID,
	}, nil)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.Equal(t, "https://web.example.com/login?error=Login+failed%3B+Invalid+s+ID+or+password.&id="+reqID, res.Header.Get("Location"))

	res = send(http.MethodPost, ts.URL+"/api/login", true, map[string]string{
		"username": "aaa@example.com",
		"password": "aaa",
		"id":       reqID,
	}, nil)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.Equal(t, "https://example.com/authorize/callback?id="+reqID, res.Header.Get("Location"))

	// step 3
	res = send(http.MethodGet, ts.URL+"/authorize/callback?id="+reqID, false, nil, nil)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	loc = res.Header.Get("Location")
	assert.Contains(t, loc, "https://web.example.com?code=")
	locu := lo.Must(url.Parse(loc))
	assert.Equal(t, "hogestate", locu.Query().Get("state"))
	code := locu.Query().Get("code")

	// step 4
	res2 := send(http.MethodPost, ts.URL+"/oauth/token", true, map[string]string{
		"grant_type":    "authorization_code",
		"redirect_uri":  "https://web.example.com",
		"client_id":     "default-client",
		"code":          code,
		"code_verifier": verifier,
	}, nil)
	var r map[string]any
	lo.Must0(json.Unmarshal(lo.Must(io.ReadAll(res2.Body)), &r))
	assert.Equal(t, map[string]any{
		"id_token":      r["id_token"],
		"access_token":  r["access_token"],
		"expires_in":    r["expires_in"],
		"refresh_token": r["refresh_token"],
		"token_type":    "Bearer",
		"state":         "hogestate",
	}, r)
	accessToken := r["access_token"].(string)
	idToken := r["id_token"].(string)
	refreshToken := r["refresh_token"].(string)

	// userinfo
	res3 := send(http.MethodGet, ts.URL+"/userinfo", false, nil, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	var r2 map[string]any
	lo.Must0(json.Unmarshal(lo.Must(io.ReadAll(res3.Body)), &r2))
	assert.Equal(t, map[string]any{
		"sub":            "subsub",
		"email":          "aaa@example.com",
		"name":           "aaa",
		"email_verified": true,
	}, r2)

	// openid-configuration
	res4 := send(http.MethodGet, ts.URL+"/.well-known/openid-configuration", false, nil, nil)
	var r3 map[string]any
	lo.Must0(json.Unmarshal(lo.Must(io.ReadAll(res4.Body)), &r3))
	assert.Equal(t, "https://example.com/jwks.json", r3["jwks_uri"])

	// jwks
	res5 := send(http.MethodGet, ts.URL+"/jwks.json", false, nil, nil)
	var jwks jose.JSONWebKeySet
	lo.Must0(json.Unmarshal(lo.Must(io.ReadAll(res5.Body)), &jwks))

	res6 := send(http.MethodPost, ts.URL+"/oauth/token", true, map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     "default-client",
		"refresh_token": refreshToken,
	}, nil)
	var r4 map[string]any
	lo.Must0(json.Unmarshal(lo.Must(io.ReadAll(res6.Body)), &r4))
	assert.Equal(t, map[string]any{
		"access_token":  r4["access_token"],
		"refresh_token": r4["refresh_token"],
		"id_token":      r4["id_token"],
		"token_type":    "Bearer",
		"expires_in":    r4["expires_in"],
	}, r4)
	accessToken2 := r4["access_token"].(string)
	refreshToken2 := r4["refresh_token"].(string)

	// confirm access_token and refresh_token are rotated
	assert.NotEqual(t, accessToken, accessToken2)
	assert.NotEqual(t, refreshToken, refreshToken2)

	// validate access_token
	token := lo.Must(jwt.ParseSigned(accessToken))
	header, _ := lo.Find(token.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key := jwks.Key(header.KeyID)[0]
	claims := map[string]any{}
	lo.Must0(token.Claims(key.Key, &claims))
	assert.Equal(t, map[string]any{
		"iss": "https://example.com/",
		"sub": "subsub",
		"aud": []any{"https://example.com"},
		"jti": claims["jti"],
		"exp": claims["exp"],
		"nbf": claims["nbf"],
		"iat": claims["iat"],
	}, claims)

	// validate id_token
	token2 := lo.Must(jwt.ParseSigned(idToken))
	header2, _ := lo.Find(token2.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key2 := jwks.Key(header2.KeyID)[0]
	claims2 := map[string]any{}
	lo.Must0(token2.Claims(key2.Key, &claims2))
	assert.Equal(t, map[string]any{
		"sub":            "subsub",
		"iss":            "https://example.com/",
		"aud":            []any{"https://example.com", "default-client"},
		"exp":            claims2["exp"],
		"iat":            claims2["iat"],
		"amr":            []any{"password"},
		"azp":            "default-client",
		"auth_time":      claims2["auth_time"],
		"at_hash":        claims2["at_hash"],
		"c_hash":         claims2["c_hash"],
		"email":          "aaa@example.com",
		"email_verified": true,
		"name":           "aaa",
	}, claims2)

	// validate refresh_token
	token3 := lo.Must(jwt.ParseSigned(refreshToken))
	header3, _ := lo.Find(token3.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key3 := jwks.Key(header3.KeyID)[0]
	claims3 := map[string]any{}
	lo.Must0(token3.Claims(key3.Key, &claims3))
	assert.Equal(t, map[string]any{
		"iss":       "https://example.com/",
		"sub":       "subsub",
		"aud":       []any{"https://example.com"},
		"jti":       claims3["jti"],
		"exp":       claims3["exp"],
		"iat":       claims3["iat"],
		"client_id": "default-client",
		"auth_id":   claims3["auth_id"],
		"auth_time": claims3["auth_time"],
		"scope":     "openid email profile offline_access",
		"amr":       []any{"password"},
	}, claims3)

	// validate access_token 2
	token4 := lo.Must(jwt.ParseSigned(accessToken2))
	header4, _ := lo.Find(token4.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key4 := jwks.Key(header4.KeyID)[0]
	claims4 := map[string]any{}
	lo.Must0(token4.Claims(key4.Key, &claims4))
	assert.Equal(t, map[string]any{
		"iss": "https://example.com/",
		"sub": "subsub",
		"aud": []any{"https://example.com"},
		"jti": claims4["jti"],
		"exp": claims4["exp"],
		"nbf": claims4["nbf"],
		"iat": claims4["iat"],
	}, claims4)

	// validate refresh_token 2
	token5 := lo.Must(jwt.ParseSigned(refreshToken2))
	header5, _ := lo.Find(token5.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key5 := jwks.Key(header5.KeyID)[0]
	claims5 := map[string]any{}
	lo.Must0(token5.Claims(key5.Key, &claims5))
	assert.Equal(t, map[string]any{
		"iss":       "https://example.com/",
		"sub":       "subsub",
		"aud":       []any{"https://example.com"},
		"jti":       claims5["jti"],
		"exp":       claims5["exp"],
		"iat":       claims5["iat"],
		"client_id": "default-client",
		"auth_id":   claims5["auth_id"],
		"auth_time": claims5["auth_time"],
		"scope":     "openid email profile offline_access",
		"amr":       []any{"password"},
	}, claims5)

	assert.Equal(t, claims3["auth_id"], claims5["auth_id"])
	assert.Equal(t, claims3["auth_time"], claims5["auth_time"])
}

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func send(method, u string, form bool, body any, headers map[string]string) *http.Response {
	var b io.Reader
	if body != nil {
		if method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut {
			if form {
				values := url.Values{}
				for k, v := range body.(map[string]string) {
					values.Set(k, v)
				}
				b = strings.NewReader(values.Encode())
			} else {
				j := lo.Must(json.Marshal(body))
				b = bytes.NewReader(j)
			}
		} else if b, ok := body.(map[string]string); ok {
			u2 := lo.Must(url.Parse(u))
			q := u2.Query()
			for k, v := range b {
				q.Set(k, v)
			}
			u2.RawQuery = q.Encode()
			u = u2.String()
		}
	}

	req := lo.Must(http.NewRequest(method, u, b))
	if b != nil {
		if !form {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return lo.Must(httpClient.Do(req))
}

func TestEndpointHTTPDomain(t *testing.T) {
	cr := &configRepo{}
	rr := &requestRepo{}

	// should not fail
	e := echo.New()
	Endpoint(context.Background(), EndpointConfig{
		// Issuer should be same as the URL
		URL:             lo.Must(url.Parse("http://example.com")),
		WebURL:          lo.Must(url.Parse("http://web.example.com")),
		Key:             "",
		DefaultClientID: "default-client",
		UserRepo:        &userRepo{},
		ConfigRepo:      cr,
		RequestRepo:     rr,
	}, e.Group(""))
}

type userRepo struct{}

func (*userRepo) Sub(ctx context.Context, email, password, _requestID string) (string, error) {
	if email == "aaa@example.com" && password == "aaa" {
		return "subsub", nil
	}
	return "", errors.New("not found")
}

func (*userRepo) Info(ctx context.Context, sub string, _ []string, ui oidc.UserInfoSetter) error {
	if sub == "subsub" {
		ui.SetEmail("aaa@example.com", true)
		ui.SetName("aaa")
		return nil
	}
	return errors.New("not found")
}

type configRepo struct {
	m sync.Mutex
	c Config
}

func (r *configRepo) Load(_ context.Context) (*Config, error) {
	if r.c.Key == "" {
		return nil, nil
	}
	c := r.c
	return &c, nil
}

func (r *configRepo) Save(_ context.Context, config *Config) error {
	if config == nil {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.c = *config
	return nil
}

func (r *configRepo) Unlock(_ context.Context) error {
	return nil
}

type requestRepo struct {
	m util.SyncMap[RequestID, *Request]
}

func (r *requestRepo) FindByID(ctx context.Context, id RequestID) (*Request, error) {
	return util.DR(r.m.Load(id)), nil
}

func (r *requestRepo) FindByCode(ctx context.Context, code string) (*Request, error) {
	return r.m.Find(func(k RequestID, r *Request) bool {
		return r.GetCode() == code
	}), nil
}

func (r *requestRepo) FindBySubject(ctx context.Context, sub string) (*Request, error) {
	return r.m.Find(func(k RequestID, r *Request) bool {
		return r.GetSubject() == sub
	}), nil
}

func (r *requestRepo) Save(ctx context.Context, req *Request) error {
	r.m.Store(req.ID(), req)
	return nil
}

func (r *requestRepo) Remove(ctx context.Context, id RequestID) error {
	r.m.Delete(id)
	return nil
}

func codeChallenge(seed []byte) (string, string) {
	verifier := base64.RawURLEncoding.EncodeToString(seed)
	challengeSum := sha256.Sum256([]byte(verifier))
	challenge := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(base64.StdEncoding.EncodeToString(challengeSum[:]), "+", "-"), "/", "_"), "=", "")
	return verifier, challenge
}

func randomCodeChallenge() (string, string) {
	seed := make([]byte, 32)
	_, _ = rand.Read(seed)
	return codeChallenge(seed)
}
