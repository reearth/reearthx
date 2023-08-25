package appx

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt"
	"github.com/jarcoal/httpmock"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
	jwt2 "gopkg.in/square/go-jose.v2/jwt"
)

func TestMultiValidator(t *testing.T) {
	key := lo.Must(rsa.GenerateKey(rand.Reader, 2048))

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://example.com/.well-known/openid-configuration",
		util.DR(httpmock.NewJsonResponder(http.StatusOK, map[string]string{"jwks_uri": "https://example.com/jwks"})),
	)
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://example2.com/.well-known/openid-configuration",
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

	v, err := NewJWTMultipleValidator([]JWTProvider{
		{ISS: "https://example.com/", AUD: []string{"a", "b"}, ALG: &jwt.SigningMethodRS256.Name},
		{ISS: "https://example2.com/", AUD: []string{"c"}, ALG: &jwt.SigningMethodRS256.Name},
	})
	assert.NoError(t, err)

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

	claims2 := jwt.MapClaims{
		"exp":      expiry,
		"iss":      "https://example2.com/",
		"sub":      "subsub2",
		"aud":      "c",
		"name":     "aaa",
		"nickname": "bbb",
	}
	token2 := jwt.NewWithClaims(jwt.SigningMethodRS256, claims2)
	token2.Header["kid"] = "0"
	tokenString2 := lo.Must(token2.SignedString(key))

	claims3 := jwt.MapClaims{
		"exp": expiry,
		"iss": "https://example3.com/",
		"aud": "c",
	}
	token3 := jwt.NewWithClaims(jwt.SigningMethodRS256, claims3)
	token3.Header["kid"] = "0"
	tokenString3 := lo.Must(token3.SignedString(key))

	res, err := v.ValidateToken(context.Background(), tokenString)
	assert.NoError(t, err)
	assert.Equal(t, &validator.ValidatedClaims{
		CustomClaims: &customClaims{
			Name:          "aaa",
			Nickname:      "bbb",
			Email:         "ccc",
			EmailVerified: lo.ToPtr(true),
		},
		RegisteredClaims: validator.RegisteredClaims{
			Issuer:   "https://example.com/",
			Subject:  "subsub",
			Audience: []string{"a", "b"},
			Expiry:   expiry,
		},
	}, res)

	res2, err := v.ValidateToken(context.Background(), tokenString2)
	assert.NoError(t, err)
	assert.Equal(t, &validator.ValidatedClaims{
		CustomClaims: &customClaims{
			Name:     "aaa",
			Nickname: "bbb",
		},
		RegisteredClaims: validator.RegisteredClaims{
			Issuer:   "https://example2.com/",
			Subject:  "subsub2",
			Audience: []string{"c"},
			Expiry:   expiry,
		},
	}, res2)

	res3, err := v.ValidateToken(context.Background(), tokenString3)
	assert.ErrorIs(t, err, jwt2.ErrInvalidIssuer)
	assert.Nil(t, res3)
}
