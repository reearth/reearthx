package authserver

import (
	"time"

	"github.com/reearth/reearthx/idx"
	"github.com/zitadel/oidc/pkg/oidc"
)

var essentialScopes = []string{"openid", "profile", "email"}

type RequestIDType struct{}

func (a RequestIDType) Type() string {
	return "authRequest"
}

type RequestID = idx.ID[RequestIDType]

func NewRequestID() RequestID {
	return idx.New[RequestIDType]()
}

func RequestIDFrom(id string) (RequestID, error) {
	return idx.From[RequestIDType](id)
}

type Request struct {
	id            RequestID
	clientID      string
	subject       string
	code          string
	state         string
	responseType  oidc.ResponseType
	scopes        []string
	audiences     []string
	redirectURI   string
	nonce         string
	codeChallenge *oidc.CodeChallenge
	authorizedAt  *time.Time
}

func (a *Request) ID() RequestID {
	return a.id
}

func (a *Request) GetID() string {
	return a.id.String()
}

func (a *Request) GetACR() string {
	return ""
}

func (a *Request) GetAMR() []string {
	return []string{
		"password",
	}
}

func (a *Request) GetAudience() []string {
	if a.audiences == nil {
		return make([]string, 0)
	}

	return a.audiences
}

func (a *Request) GetAuthTime() time.Time {
	return a.CreatedAt()
}

func (a *Request) GetClientID() string {
	return a.clientID
}

func (a *Request) GetResponseMode() oidc.ResponseMode {
	// TODO make sure about this
	return oidc.ResponseModeQuery
}

func (a *Request) GetCode() string {
	return a.code
}

func (a *Request) GetState() string {
	return a.state
}

func (a *Request) GetCodeChallenge() *oidc.CodeChallenge {
	return a.codeChallenge
}

func (a *Request) GetNonce() string {
	return a.nonce
}

func (a *Request) GetRedirectURI() string {
	return a.redirectURI
}

func (a *Request) GetResponseType() oidc.ResponseType {
	return a.responseType
}

func (a *Request) GetScopes() []string {
	return unique(append(a.scopes, essentialScopes...))
}

func (a *Request) SetCurrentScopes(scopes []string) {
	a.scopes = unique(append(scopes, essentialScopes...))
}

func (a *Request) GetSubject() string {
	return a.subject
}

func (a *Request) CreatedAt() time.Time {
	return a.id.Timestamp()
}

func (a *Request) AuthorizedAt() *time.Time {
	return a.authorizedAt
}

func (a *Request) SetAuthorizedAt(authorizedAt *time.Time) {
	a.authorizedAt = authorizedAt
}

func (a *Request) Done() bool {
	return a.authorizedAt != nil
}

func (a *Request) Complete(sub string) {
	a.subject = sub
	now := time.Now()
	a.authorizedAt = &now
}

func (a *Request) SetCode(code string) {
	a.code = code
}

func unique(list []string) []string {
	allKeys := make(map[string]struct{})
	var uniqueList []string
	for _, item := range list {
		if _, ok := allKeys[item]; !ok {
			allKeys[item] = struct{}{}
			uniqueList = append(uniqueList, item)
		}
	}
	return uniqueList
}
