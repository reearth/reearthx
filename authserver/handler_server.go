package authserver

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/zitadel/oidc/pkg/op"
)

const jwksEndpoint = "jwks.json"

type ServerConfig struct {
	Issuer  string
	Key     string
	Storage op.Storage
}

func Server(ctx context.Context, cfg ServerConfig) (*mux.Router, error) {
	handler, err := op.NewOpenIDProvider(
		ctx,
		&op.Config{
			Issuer:                cfg.Issuer,
			CryptoKey:             sha256.Sum256([]byte(cfg.Key)),
			GrantTypeRefreshToken: true,
		},
		cfg.Storage,
		op.WithHttpInterceptors(jsonToFormHandler()),
		op.WithHttpInterceptors(setURLVarsHandler()),
		op.WithCustomEndSessionEndpoint(op.NewEndpoint(logoutEndpoint)),
		op.WithCustomKeysEndpoint(op.NewEndpoint(jwksEndpoint)),
	)
	if err != nil {
		return nil, err
	}

	return handler.HttpHandler().(*mux.Router), nil
}

func setURLVarsHandler() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/authorize/callback" {
				handler.ServeHTTP(w, r)
				return
			}

			r2 := mux.SetURLVars(r, map[string]string{"id": r.URL.Query().Get("id")})
			handler.ServeHTTP(w, r2)
		})
	}
}

func jsonToFormHandler() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/oauth/token" {
				handler.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Content-Type") != "" {
				value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
				if value != "application/json" {
					// Content-Type header is not application/json
					handler.ServeHTTP(w, r)
					return
				}
			}

			if err := r.ParseForm(); err != nil {
				return
			}

			var result map[string]string

			if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			for key, value := range result {
				r.Form.Set(key, value)
			}

			handler.ServeHTTP(w, r)
		})
	}
}
