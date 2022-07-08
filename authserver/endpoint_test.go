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
		SubLoader: func(ctx context.Context, email, password, _requestID string) (string, error) {
			if email == "aaa@example.com" && password == "aaa" {
				return "subsub", nil
			}
			return "", errors.New("not found")
		},
		UserInfoProvider: func(ctx context.Context, sub string, scope []string, ui oidc.UserInfoSetter) error {
			if sub == "subsub" {
				ui.SetEmail("aaa@example.com", true)
				ui.SetName("aaa")
				return nil
			}
			return errors.New("not found")
		},
		Issuer:           "https://example.com/",
		AuthProviderName: "test",
		URL:              util.Unwrap(url.Parse("https://example.com")),
		WebURL:           util.Unwrap(url.Parse("https://web.example.com")),
		Key:              "",
		Dev:              false,
		DefaultClientID:  "default-client",
		ConfigRepo:       cr,
		RequestRepo:      rr,
	}, e.Group(""))

	ts := httptest.NewServer(e)
	defer ts.Close()

	// step 1
	verifier, challenge := randomCodeChallenge()
	res := send(http.MethodGet, ts.URL+"/authorize", false, map[string]string{
		"response_type":         "code",
		"client_id":             "default-client",
		"redirect_uri":          "https://web.example.com",
		"scope":                 "openid email profile",
		"state":                 "hogestate",
		"code_challenge":        challenge,
		"code_challenge_method": "S256",
	}, nil)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	loc := res.Header.Get("Location")
	assert.Contains(t, loc, "https://web.example.com/login?id=")
	reqID := util.Unwrap(url.Parse(loc)).Query().Get("id")

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
	locu := util.Unwrap(url.Parse(loc))
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
	util.Must(json.Unmarshal(util.Unwrap(io.ReadAll(res2.Body)), &r))
	assert.Equal(t, map[string]any{
		"id_token":     r["id_token"],
		"access_token": r["access_token"],
		"expires_in":   r["expires_in"],
		"token_type":   "Bearer",
		"state":        "hogestate",
	}, r)
	accessToken := r["access_token"].(string)
	idToken := r["id_token"].(string)

	// userinfo
	res3 := send(http.MethodGet, ts.URL+"/userinfo", false, nil, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	var r2 map[string]any
	util.Must(json.Unmarshal(util.Unwrap(io.ReadAll(res3.Body)), &r2))
	assert.Equal(t, map[string]any{
		"sub":            "subsub",
		"email":          "aaa@example.com",
		"name":           "aaa",
		"email_verified": true,
	}, r2)

	// openid-configuration
	res4 := send(http.MethodGet, ts.URL+"/.well-known/openid-configuration", false, nil, nil)
	var r3 map[string]any
	util.Must(json.Unmarshal(util.Unwrap(io.ReadAll(res4.Body)), &r3))
	assert.Equal(t, "https://example.com/jwks.json", r3["jwks_uri"])

	// jwks
	res5 := send(http.MethodGet, ts.URL+"/jwks.json", false, nil, nil)
	var jwks jose.JSONWebKeySet
	util.Must(json.Unmarshal(util.Unwrap(io.ReadAll(res5.Body)), &jwks))

	// validate access_token
	token := util.Unwrap(jwt.ParseSigned(accessToken))
	header, _ := lo.Find(token.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key := jwks.Key(header.KeyID)[0]
	claims := map[string]any{}
	util.Must(token.Claims(key.Key, &claims))
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
	token2 := util.Unwrap(jwt.ParseSigned(idToken))
	header2, _ := lo.Find(token2.Headers, func(h jose.Header) bool {
		return h.Algorithm == string(jose.RS256)
	})
	key2 := jwks.Key(header2.KeyID)[0]
	claims2 := map[string]any{}
	util.Must(token.Claims(key2.Key, &claims2))
	assert.Equal(t, map[string]any{
		"iss": "https://example.com/",
		"sub": "subsub",
		"aud": []any{"https://example.com"},
		"jti": claims["jti"],
		"exp": claims["exp"],
		"nbf": claims["nbf"],
		"iat": claims["iat"],
	}, claims2)
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
				j := util.Unwrap(json.Marshal(body))
				b = bytes.NewReader(j)
			}
		} else if b, ok := body.(map[string]string); ok {
			u2 := util.Unwrap(url.Parse(u))
			q := u2.Query()
			for k, v := range b {
				q.Set(k, v)
			}
			u2.RawQuery = q.Encode()
			u = u2.String()
		}
	}

	req := util.Unwrap(http.NewRequest(method, u, b))
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
	return util.Unwrap(httpClient.Do(req))
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
