package appx

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt"
	"github.com/jarcoal/httpmock"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"gopkg.in/square/go-jose.v2"
)

func TestAuthInfoMiddleware(t *testing.T) {
	key := struct{}{}
	m := AuthInfoMiddleware(key)
	h := m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res any
		v := r.Context().Value(key)
		if a, ok := v.(AuthInfo); ok {
			res = a
		} else {
			res = "error"
		}
		payload, _ := json.Marshal(res)
		_, _ = w.Write(payload)
	}))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.Header.Get("Authorization") != "" {
			claims := &validator.ValidatedClaims{
				CustomClaims: &customClaims{
					Name:          "aaa",
					Nickname:      "bbb",
					Email:         "ccc",
					EmailVerified: lo.ToPtr(true),
				},
				RegisteredClaims: validator.RegisteredClaims{
					Subject: "subsub",
					Issuer:  "ississ",
				},
			}
			ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, claims)
		}
		h.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer ts.Close()

	// normal
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Authorization", "Bearer aaaaa")
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	body, _ := io.ReadAll(res.Body)
	want, _ := json.Marshal(AuthInfo{
		Token:         "aaaaa",
		Sub:           "subsub",
		Iss:           "ississ",
		Name:          "bbb",
		Email:         "ccc",
		EmailVerified: lo.ToPtr(true),
	})
	assert.Equal(t, string(want), string(body))

	// abnormal
	req, _ = http.NewRequest(http.MethodGet, ts.URL, nil)
	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	body, _ = io.ReadAll(res.Body)
	assert.Equal(t, `"error"`, string(body))
}

func TestAuthMiddleware(t *testing.T) {
	ctxkey := struct{}{}
	m, err := AuthMiddleware([]JWTProvider{
		{ISS: "https://example.com/", AUD: []string{"a"}, ALG: &jwt.SigningMethodRS256.Name},
	}, ctxkey, false)
	assert.NoError(t, err)
	ts := httptest.NewServer(m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res any
		v := r.Context().Value(ctxkey)
		if a, ok := v.(AuthInfo); ok {
			res = a
		} else {
			res = "error"
		}
		payload, _ := json.Marshal(res)
		_, _ = w.Write(payload)
	})))
	defer ts.Close()

	key := lo.Must(rsa.GenerateKey(rand.Reader, 2048))

	tr := NewTestTransport([]string{ts.URL})
	defer tr.Activate()()

	httpmock.RegisterResponder(
		http.MethodGet,
		"https://example.com/.well-known/openid-configuration",
		util.DR(httpmock.NewJsonResponder(http.StatusOK, map[string]string{"jwks_uri": "https://example.com/jwks"})),
	)
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://example.com/jwks",
		httpmock.NewBytesResponder(http.StatusOK, lo.Must(json.Marshal(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{
				{KeyID: "0", Key: &key.PublicKey, Algorithm: jwt.SigningMethodRS256.Name},
			},
		}))),
	)

	expiry := time.Now().Add(time.Hour * 24).Unix()
	claims := jwt.MapClaims{
		"exp":            expiry,
		"iss":            "https://example.com/",
		"sub":            "subsub",
		"aud":            []string{"a", "b"},
		"name":           "aaa",
		"nickname":       "bbb",
		"email":          "ccc",
		"email_verified": true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "0"
	tokenString := lo.Must(token.SignedString(key))

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	body, _ := io.ReadAll(res.Body)
	want, _ := json.Marshal(AuthInfo{
		Token:         tokenString,
		Sub:           "subsub",
		Iss:           "https://example.com/",
		Name:          "bbb",
		Email:         "ccc",
		EmailVerified: lo.ToPtr(true),
	})
	assert.Equal(t, string(want), string(body))
}

type TestTransport struct {
	whitelist []string
}

func NewTestTransport(whitelist []string) *TestTransport {
	return &TestTransport{
		whitelist: whitelist,
	}
}

func (t *TestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if slices.Contains(t.whitelist, req.URL.String()) {
		return httpmock.InitialTransport.RoundTrip(req)
	}
	return httpmock.DefaultTransport.RoundTrip(req)
}

func (t *TestTransport) Activate() func() {
	if !httpmock.Disabled() {
		httpmock.Activate()
		http.DefaultTransport = t
	}
	return t.Deactivate
}

func (t *TestTransport) Deactivate() {
	httpmock.Deactivate()
}
