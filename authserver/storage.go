package authserver

import (
	"context"
	"crypto/rsa"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/log"
	"github.com/zitadel/oidc/pkg/crypto"
	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/oidc/pkg/op"
	"gopkg.in/square/go-jose.v2"
)

type Storage struct {
	config         StorageConfig
	userInfoSetter UserInfoProvider
	clients        map[string]op.Client
	requestRepo    RequestRepo
	keySet         jose.JSONWebKeySet
	key            *rsa.PrivateKey
	sigKey         jose.SigningKey
	issuer         string
}

var _ op.Storage = (*Storage)(nil)

type UserInfoProvider func(context.Context, string, []string, oidc.UserInfoSetter) error

type StorageConfig struct {
	ClientID        string
	ClientDomain    string
	Domain          string
	Dev             bool
	DN              *DNConfig
	ConfigRepo      ConfigRepo
	RequestRepo     RequestRepo
	UserInfoSetter  UserInfoProvider
	AudienceForTest string
}

type DNConfig struct {
	CommonName         string
	Organization       []string
	OrganizationalUnit []string
	Country            []string
	Province           []string
	Locality           []string
	StreetAddress      []string
	PostalCode         []string
}

var dummyName = pkix.Name{
	CommonName:         "Dummy company, INC.",
	Organization:       []string{"Dummy company, INC."},
	OrganizationalUnit: []string{"Dummy OU"},
	Country:            []string{"US"},
	Province:           []string{"Dummy"},
	Locality:           []string{"Dummy locality"},
	StreetAddress:      []string{"Dummy street"},
	PostalCode:         []string{"1"},
}

func NewStorage(ctx context.Context, cfg StorageConfig, issuer string) (op.Storage, error) {
	client := NewLocalClient(cfg.Dev, cfg.ClientID, cfg.ClientDomain)

	name := dummyName
	if cfg.DN != nil {
		name = pkix.Name{
			CommonName:         cfg.DN.CommonName,
			Organization:       cfg.DN.Organization,
			OrganizationalUnit: cfg.DN.OrganizationalUnit,
			Country:            cfg.DN.Country,
			Province:           cfg.DN.Province,
			Locality:           cfg.DN.Locality,
			StreetAddress:      cfg.DN.StreetAddress,
			PostalCode:         cfg.DN.PostalCode,
		}
	}
	c, err := cfg.ConfigRepo.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not load auth config: %w\n", err)
	}
	defer func() {
		if err := cfg.ConfigRepo.Unlock(ctx); err != nil {
			log.Errorf("auth: could not release config lock: %s\n", err)
		}
	}()

	var keyBytes, certBytes []byte
	if c != nil {
		keyBytes = []byte(c.Key)
		certBytes = []byte(c.Cert)
	} else {
		keyBytes, certBytes, err = generateCert(name)
		if err != nil {
			return nil, fmt.Errorf("could not generate raw cert: %w\n", err)
		}
		c = &Config{
			Key:  string(keyBytes),
			Cert: string(certBytes),
		}

		if err := cfg.ConfigRepo.Save(ctx, c); err != nil {
			return nil, fmt.Errorf("could not save raw cert: %w\n", err)
		}
		log.Info("auth: init a new private key and certificate")
	}

	key, sigKey, keySet, err := initKeys(keyBytes, certBytes)
	if err != nil {
		return nil, fmt.Errorf("could not init keys: %w\n", err)
	}

	return &Storage{
		config:         cfg,
		userInfoSetter: cfg.UserInfoSetter,
		requestRepo:    cfg.RequestRepo,
		key:            key,
		sigKey:         *sigKey,
		keySet:         *keySet,
		clients: map[string]op.Client{
			client.GetID(): client,
		},
		issuer: issuer,
	}, nil
}

func (s *Storage) Health(_ context.Context) error {
	return nil
}

func (s *Storage) CreateAuthRequest(ctx context.Context, authReq *oidc.AuthRequest, _ string) (op.AuthRequest, error) {
	audiences := []string{
		s.config.Domain,
	}
	if s.config.Dev && s.config.AudienceForTest != "" {
		audiences = append(audiences, s.config.AudienceForTest)
	}

	var cc *oidc.CodeChallenge
	if authReq.CodeChallenge != "" {
		cc = &oidc.CodeChallenge{
			Challenge: authReq.CodeChallenge,
			Method:    authReq.CodeChallengeMethod,
		}
	}
	var request = NewRequest().
		NewID().
		ClientID(authReq.ClientID).
		State(authReq.State).
		ResponseType(authReq.ResponseType).
		Scopes(authReq.Scopes).
		Audiences(audiences).
		RedirectURI(authReq.RedirectURI).
		Nonce(authReq.Nonce).
		CodeChallenge(cc).
		AuthorizedAt(nil).
		MustBuild()

	if err := s.requestRepo.Save(ctx, request); err != nil {
		return nil, err
	}
	return request, nil
}

func (s *Storage) AuthRequestByID(ctx context.Context, requestID string) (op.AuthRequest, error) {
	if requestID == "" {
		return nil, errors.New("invalid id")
	}
	reqId, err := RequestIDFrom(requestID)
	if err != nil {
		return nil, err
	}
	request, err := s.requestRepo.FindByID(ctx, reqId)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (s *Storage) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	if code == "" {
		return nil, errors.New("invalid code")
	}
	return s.requestRepo.FindByCode(ctx, code)
}

func (s *Storage) AuthRequestBySubject(ctx context.Context, subject string) (op.AuthRequest, error) {
	if subject == "" {
		return nil, errors.New("invalid subject")
	}

	return s.requestRepo.FindBySubject(ctx, subject)
}

func (s *Storage) SaveAuthCode(ctx context.Context, requestID, code string) error {
	request, err := s.AuthRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	request2 := request.(*Request)
	request2.SetCode(code)
	return s.updateRequest(ctx, requestID, *request2)
}

func (s *Storage) DeleteAuthRequest(ctx context.Context, requestID string) error {
	reqId, err := RequestIDFrom(requestID)
	if err != nil {
		return err
	}
	return s.requestRepo.Remove(ctx, reqId)
}

func (s *Storage) CreateAccessToken(_ context.Context, _ op.TokenRequest) (string, time.Time, error) {
	return uuid.NewString(), time.Now().UTC().Add(5 * time.Hour), nil
}

func (s *Storage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, refreshToken string) (accessTokenID string, newRefreshToken string, expiration time.Time, err error) {
	keyCh := make(chan jose.SigningKey, 1)
	s.GetSigningKey(ctx, keyCh)
	signer := op.NewSigner(ctx, s, keyCh)

	now := time.Now().UTC()
	refreshTokenExpiresAt := now.Add(24 * time.Hour)

	var client op.Client
	var authTime time.Time
	var authID string
	var amr []string
	switch authReq := request.(type) {
	case *Request:
		client, _ = s.GetClientByClientID(ctx, authReq.GetClientID())
		authID = uuid.NewString()
		authTime = now
		amr = authReq.GetAMR()
	case *refreshTokenClaims:
		client, _ = s.GetClientByClientID(ctx, authReq.GetClientID())
		authID = authReq.AuthID
		authTime = authReq.GetAuthTime()
		amr = authReq.GetAMR()
	}

	audience := request.GetAudience()
	skew := client.ClockSkew()
	if len(audience) == 0 {
		audience = append(audience, client.GetID())
	}
	claims := &refreshTokenClaims{
		JWTID:     uuid.NewString(),
		AuthID:    authID,
		AMR:       amr,
		Issuer:    s.issuer,
		Subject:   request.GetSubject(),
		Scope:     request.GetScopes(),
		Audience:  audience,
		IssuedAt:  oidc.Time(now.Add(-skew)),
		ExpiresAt: oidc.Time(refreshTokenExpiresAt),
		ClientID:  client.GetID(),
		AuthTime:  oidc.Time(authTime),
	}

	newRefreshToken, err = crypto.Sign(claims, signer.Signer())
	if err != nil {
		return
	}
	return uuid.NewString(), newRefreshToken, now.Add(5 * time.Minute), nil
}

func (s *Storage) TokenRequestByRefreshToken(ctx context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
	v := op.NewAccessTokenVerifier(s.issuer, keySet{&s.keySet})
	claims, err := s.verifyRefreshToken(ctx, refreshToken, v)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	return claims, nil
}

func (s *Storage) TerminateSession(_ context.Context, _, _ string) error {
	return nil
}

func (s *Storage) GetSigningKey(_ context.Context, keyCh chan<- jose.SigningKey) {
	keyCh <- s.sigKey
}

func (s *Storage) GetKeySet(_ context.Context) (*jose.JSONWebKeySet, error) {
	return &s.keySet, nil
}

func (s *Storage) GetKeyByIDAndUserID(_ context.Context, kid, _ string) (*jose.JSONWebKey, error) {
	return &s.keySet.Key(kid)[0], nil
}

func (s *Storage) GetClientByClientID(_ context.Context, clientID string) (op.Client, error) {
	if clientID == "" {
		return nil, errors.New("invalid client id")
	}

	client, exists := s.clients[clientID]
	if !exists {
		return nil, errors.New("not found")
	}

	return client, nil
}

func (s *Storage) AuthorizeClientIDSecret(_ context.Context, _ string, _ string) error {
	return nil
}

func (s *Storage) SetUserinfoFromToken(ctx context.Context, userinfo oidc.UserInfoSetter, _tokenID, subject, _origin string) error {
	userinfo.SetSubject(subject)
	return s.userInfoSetter(ctx, subject, nil, userinfo)
}

func (s *Storage) SetUserinfoFromScopes(ctx context.Context, userinfo oidc.UserInfoSetter, subject, _clientID string, scope []string) error {
	if err := s.userInfoSetter(ctx, subject, scope, userinfo); err != nil {
		return err
	}
	userinfo.SetSubject(subject)
	return nil
}

func (s *Storage) GetPrivateClaimsFromScopes(_ context.Context, _, _ string, _ []string) (map[string]interface{}, error) {
	return nil, nil
}

func (s *Storage) SetIntrospectionFromToken(ctx context.Context, introspect oidc.IntrospectionResponse, _, subject, clientID string) error {
	if err := s.SetUserinfoFromScopes(ctx, introspect, subject, clientID, []string{}); err != nil {
		return err
	}
	request, err := s.AuthRequestBySubject(ctx, subject)
	if err != nil {
		return err
	}
	introspect.SetClientID(request.GetClientID())
	return nil
}

func (s *Storage) ValidateJWTProfileScopes(_ context.Context, _ string, scope []string) ([]string, error) {
	return scope, nil
}

func (s *Storage) RevokeToken(_ context.Context, _ string, _ string, _ string) *oidc.Error {
	return nil
}

func (s *Storage) CompleteAuthRequest(ctx context.Context, requestId, sub string) error {
	request, err := s.AuthRequestByID(ctx, requestId)
	if err != nil {
		return err
	}
	req := request.(*Request)
	req.Complete(sub)
	err = s.updateRequest(ctx, requestId, *req)
	return err
}

func (s *Storage) updateRequest(ctx context.Context, requestID string, req Request) error {
	if requestID == "" {
		return errors.New("invalid id")
	}
	reqId, err := RequestIDFrom(requestID)
	if err != nil {
		return err
	}

	if _, err := s.requestRepo.FindByID(ctx, reqId); err != nil {
		return err
	}

	if err := s.requestRepo.Save(ctx, &req); err != nil {
		return err
	}

	return nil
}

func (s *Storage) verifyRefreshToken(ctx context.Context, token string, v op.AccessTokenVerifier) (*refreshTokenClaims, error) {
	claims := new(refreshTokenClaims)
	decrypted, err := oidc.DecryptToken(token)
	if err != nil {
		return nil, err
	}
	payload, err := oidc.ParseToken(decrypted, claims)
	if err != nil {
		return nil, err
	}
	if err = oidc.CheckIssuer(claims, v.Issuer()); err != nil {
		return nil, err
	}
	if err = oidc.CheckSignature(ctx, decrypted, payload, claims, v.SupportedSignAlgs(), v.KeySet()); err != nil {
		return nil, err
	}
	if err = oidc.CheckExpiration(claims, v.Offset()); err != nil {
		return nil, err
	}
	return claims, nil
}

type keySet struct {
	keySet *jose.JSONWebKeySet
}

func (k keySet) VerifySignature(ctx context.Context, jws *jose.JSONWebSignature) (payload []byte, err error) {
	keyID, alg := oidc.GetKeyIDAndAlg(jws)
	key, err := oidc.FindMatchingKey(keyID, oidc.KeyUseSignature, alg, k.keySet.Keys...)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}
	return jws.Verify(&key)
}
