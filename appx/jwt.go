package appx

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)

const defaultJWTTTL = 5

type AuthInfo struct {
	Token         string
	Sub           string
	Iss           string
	Name          string
	Email         string
	EmailVerified *bool
}

type JWTProvider struct {
	ISS string
	AUD []string
	ALG *string
	TTL *int
}

func (p JWTProvider) validator() (*validator.Validator, error) {
	issuerURL, err := url.Parse(p.ISS)
	issuerURL.Path = "/"
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issuer url: %w", err)
	}
	if issuerURL.Scheme != "http" && issuerURL.Scheme != "https" {
		return nil, fmt.Errorf("failed to parse the issuer url")
	}

	ttl := time.Duration(lo.FromPtrOr(p.TTL, defaultJWTTTL)) * time.Minute
	provider := jwks.NewCachingProvider(issuerURL, ttl)
	algorithm := validator.SignatureAlgorithm(lo.FromPtrOr(p.ALG, jwt.SigningMethodRS256.Name))

	var aud []string
	if p.AUD != nil {
		aud = p.AUD
	} else {
		aud = []string{}
	}

	return validator.New(
		provider.KeyFunc,
		algorithm,
		issuerURL.String(),
		aud,
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &customClaims{}
		}),
	)
}

type customClaims struct {
	Name          string `json:"name"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	EmailVerified *bool  `json:"email_verified"`
}

func (c *customClaims) Validate(ctx context.Context) error {
	return nil
}

type MultiValidator []*validator.Validator

func NewMultiValidator(providers []JWTProvider) (MultiValidator, error) {
	return util.TryMap(providers, func(p JWTProvider) (*validator.Validator, error) {
		return p.validator()
	})
}

// ValidateToken Trys to validate the token with each validator
// NOTE: the last validation error only is returned
func (mv MultiValidator) ValidateToken(ctx context.Context, tokenString string) (res interface{}, err error) {
	for _, v := range mv {
		res, err = v.ValidateToken(ctx, tokenString)
		if err == nil {
			return
		}
	}
	return
}

// AuthInfoMiddleware loads claim from context and attach the user info.
func AuthInfoMiddleware(key any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			rawClaims := ctx.Value(jwtmiddleware.ContextKey{})
			if claims, ok := rawClaims.(*validator.ValidatedClaims); ok {
				// attach auth info to context
				customClaims := claims.CustomClaims.(*customClaims)
				name := customClaims.Nickname
				if name == "" {
					name = customClaims.Name
				}
				ctx = context.WithValue(ctx, key, AuthInfo{
					Token:         strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "),
					Sub:           claims.RegisteredClaims.Subject,
					Iss:           claims.RegisteredClaims.Issuer,
					Name:          name,
					Email:         customClaims.Email,
					EmailVerified: customClaims.EmailVerified,
				})
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthMiddleware(providers []JWTProvider, key any, optional bool) (func(http.Handler) http.Handler, error) {
	v, err := NewMultiValidator(providers)
	if err != nil {
		return nil, err
	}

	jwtm := jwtmiddleware.New(v.ValidateToken, jwtmiddleware.WithCredentialsOptional(optional)).CheckJWT
	aim := AuthInfoMiddleware(key)

	return func(next http.Handler) http.Handler {
		return jwtm(aim(next))
	}, nil
}
