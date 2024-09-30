package appx

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	v, err := validator.New(
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
		validator: v,
		iss:       issuerURL,
		aud:       slices.Clone(audience),
	}, nil
}

func (v *JWTValidatorWithError) ValidateToken(ctx context.Context, token string) (interface{}, error) {
	res, err := v.validator.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT: iss=%s aud=%v err=%w", v.iss, v.aud, err)
	}
	return res, nil
}

type JWTMultipleValidator []JWTValidator

func NewJWTMultipleValidator(providers []JWTProvider) (JWTMultipleValidator, error) {
	return util.TryMap(providers, func(p JWTProvider) (JWTValidator, error) {
		return p.validator()
	})
}

// ValidateToken tries to validate the token with each validator concurrently
// NOTE: the last validation error only is returned
func (mv JWTMultipleValidator) ValidateToken(ctx context.Context, tokenString string) (interface{}, error) {
	type result struct {
		res interface{}
		err error
	}

	resultChan := make(chan result, len(mv))
	var wg sync.WaitGroup

	for _, v := range mv {
		wg.Add(1)
		go func(validator JWTValidator) {
			defer wg.Done()
			res, err := validator.ValidateToken(ctx, tokenString)
			resultChan <- result{res, err}
		}(v)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var lastErr error
	for r := range resultChan {
		if r.err == nil {
			return r.res, nil
		}
		lastErr = errors.Join(lastErr, r.err)
	}

	log.Debugfc(ctx, "auth: invalid JWT token: %s", tokenString)
	log.Errorfc(ctx, "auth: invalid JWT token: %v", lastErr)
	return nil, lastErr
}
