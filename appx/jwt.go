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
	ISS     string
	JWKSURI *string
	AUD     []string
	ALG     *string
	TTL     *int
}

func (p JWTProvider) validator() (JWTValidator, error) {
	issuerURL, err := url.Parse(p.ISS)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issuer url: %w", err)
	}
	if issuerURL.Scheme != "http" && issuerURL.Scheme != "https" {
		return nil, fmt.Errorf("failed to parse the issuer url")
	}

	opts := []jwks.ProviderOption{}
	if p.JWKSURI != nil && *p.JWKSURI != "" {
		u, err := url.Parse(*p.JWKSURI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the jwks uri: %w", err)
		}
		opts = append(opts, jwks.WithCustomJWKSURI(u))
	}
	ttl := time.Duration(lo.FromPtrOr(p.TTL, defaultJWTTTL)) * time.Minute
	interfaceOpts := make([]interface{}, len(opts))

	for i, opt := range opts {
		interfaceOpts[i] = opt
	}

	provider := jwks.NewCachingProvider(issuerURL, ttl, interfaceOpts...)
	algorithm := validator.SignatureAlgorithm(lo.FromPtrOr(p.ALG, jwt.SigningMethodRS256.Name))

	var aud []string
	if p.AUD != nil {
		aud = p.AUD
	} else {
		aud = []string{}
	}

	return NewJWTValidatorWithError(
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

func (c *customClaims) Validate(_ context.Context) error {
	return nil
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
	v, err := NewJWTMultipleValidator(providers)
	if err != nil {
		return nil, err
	}

	jwtm := jwtmiddleware.New(v.ValidateToken, jwtmiddleware.WithCredentialsOptional(optional)).CheckJWT
	aim := AuthInfoMiddleware(key)

	return func(next http.Handler) http.Handler {
		return jwtm(aim(next))
	}, nil
}
