package appx

import (
	"context"
	"errors"
	"fmt"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/util"
	"golang.org/x/exp/slices"
)

type JWTValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (interface{}, error)
}

// JWTValidatorWithError wraps "validator.Validator and attach iss and aud to the error message to make it easy to track errors.
type JWTValidatorWithError struct {
	validator *validator.Validator
	iss       string
	aud       []string
}

func NewJWTValidatorWithError(
	keyFunc func(context.Context) (interface{}, error),
	signatureAlgorithm validator.SignatureAlgorithm,
	issuerURL string,
	audience []string,
	opts ...validator.Option,
) (*JWTValidatorWithError, error) {
	validator, err := validator.New(
		keyFunc,
		signatureAlgorithm,
		issuerURL,
		audience,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &JWTValidatorWithError{
		validator: validator,
		iss:       issuerURL,
		aud:       slices.Clone(audience),
	}, nil
}

func (v *JWTValidatorWithError) ValidateToken(ctx context.Context, token string) (interface{}, error) {
	res, err := v.validator.ValidateToken(ctx, token)
	if err != nil {
		err = fmt.Errorf("invalid JWT: iss=%s aud=%v err=%w", v.iss, v.aud, err)
	}
	return res, err
}

type JWTMultipleValidator []JWTValidator

func NewJWTMultipleValidator(providers []JWTProvider) (JWTMultipleValidator, error) {
	return util.TryMap(providers, func(p JWTProvider) (JWTValidator, error) {
		return p.validator()
	})
}

// ValidateToken Trys to validate the token with each validator
// NOTE: the last validation error only is returned
func (mv JWTMultipleValidator) ValidateToken(ctx context.Context, tokenString string) (res interface{}, err error) {
	for _, v := range mv {
		var err2 error
		res, err2 = v.ValidateToken(ctx, tokenString)
		if err2 == nil {
			err = nil
			return
		}
		err = errors.Join(err, err2)
	}

	log.Debugfc(ctx, "auth: invalid JWT token: %s", tokenString)
	log.Errorfc(ctx, "auth: invalid JWT token: %v", err)
	return
}
