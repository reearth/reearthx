package authserver

import (
	"fmt"
	"time"

	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/oidc/pkg/op"
)

type Client struct {
	id                 string
	applicationType    op.ApplicationType
	authMethod         oidc.AuthMethod
	accessTokenType    op.AccessTokenType
	responseTypes      []oidc.ResponseType
	grantTypes         []oidc.GrantType
	allowedScopes      []string
	redirectURIs       []string
	logoutRedirectURIs []string
	loginURI           string
	idTokenLifetime    time.Duration
	clockSkew          time.Duration
	devMode            bool
}

func NewLocalClient(dev bool, id string, domain string) op.Client {
	return &Client{
		id:              id,
		applicationType: op.ApplicationTypeWeb,
		authMethod:      oidc.AuthMethodNone,
		accessTokenType: op.AccessTokenTypeJWT,
		responseTypes:   []oidc.ResponseType{oidc.ResponseTypeCode},
		grantTypes:      []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		redirectURIs:    []string{domain},
		allowedScopes:   []string{"openid", "profile", "email", "offline_access"},
		loginURI:        domain + "/login?id=%s",
		idTokenLifetime: 5 * time.Minute,
		clockSkew:       0,
		devMode:         dev,
	}
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) RedirectURIs() []string {
	return c.redirectURIs
}

func (c *Client) PostLogoutRedirectURIs() []string {
	return c.logoutRedirectURIs
}

func (c *Client) LoginURL(id string) string {
	return fmt.Sprintf(c.loginURI, id)
}

func (c *Client) ApplicationType() op.ApplicationType {
	return c.applicationType
}

func (c *Client) AuthMethod() oidc.AuthMethod {
	return c.authMethod
}

func (c *Client) IDTokenLifetime() time.Duration {
	return c.idTokenLifetime
}

func (c *Client) AccessTokenType() op.AccessTokenType {
	return c.accessTokenType
}

func (c *Client) ResponseTypes() []oidc.ResponseType {
	return c.responseTypes
}

func (c *Client) GrantTypes() []oidc.GrantType {
	return c.grantTypes
}

func (c *Client) DevMode() bool {
	return c.devMode
}

func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *Client) IsScopeAllowed(scope string) bool {
	for _, clientScope := range c.allowedScopes {
		if clientScope == scope {
			return true
		}
	}
	return false
}

func (c *Client) IDTokenUserinfoClaimsAssertion() bool {
	return false
}

func (c *Client) ClockSkew() time.Duration {
	return c.clockSkew
}
