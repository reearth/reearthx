package authserver

import (
	"time"

	"github.com/zitadel/oidc/pkg/oidc"
	"gopkg.in/square/go-jose.v2"
)

type refreshTokenClaims struct {
	JWTID     string                   `json:"jti"`
	AuthID    string                   `json:"auth_id"`
	AMR       []string                 `json:"amr"`
	Issuer    string                   `json:"iss"`
	Subject   string                   `json:"sub"`
	Scope     oidc.SpaceDelimitedArray `json:"scope"`
	Audience  oidc.Audience            `json:"aud"`
	IssuedAt  oidc.Time                `json:"iat"`
	ExpiresAt oidc.Time                `json:"exp"`
	ClientID  string                   `json:"client_id"`
	AuthTime  oidc.Time                `json:"auth_time"`
}

func (c *refreshTokenClaims) GetClientID() string {
	return c.ClientID
}

func (c *refreshTokenClaims) GetScopes() []string {
	return c.Scope
}

func (c *refreshTokenClaims) SetCurrentScopes(scopes []string) {
	c.Scope = scopes
}

func (c *refreshTokenClaims) GetIssuer() string {
	return c.Issuer
}

func (c *refreshTokenClaims) GetSubject() string {
	return c.Subject
}

func (c *refreshTokenClaims) GetAudience() []string {
	return c.Audience
}

func (c *refreshTokenClaims) GetExpiration() time.Time {
	return time.Time(c.ExpiresAt)
}

func (c *refreshTokenClaims) GetIssuedAt() time.Time {
	return time.Time(c.IssuedAt)
}

func (c *refreshTokenClaims) GetAuthTime() time.Time {
	return time.Time(c.AuthTime)
}

func (c *refreshTokenClaims) GetAMR() []string {
	return c.AMR
}

func (c *refreshTokenClaims) GetNonce() string {
	panic("unsupported")
}

func (c *refreshTokenClaims) GetAuthenticationContextClassReference() string {
	panic("unsupported")
}

func (c *refreshTokenClaims) GetAuthorizedParty() string {
	panic("unsupported")
}

func (c *refreshTokenClaims) SetSignatureAlgorithm(algorithm jose.SignatureAlgorithm) {}
