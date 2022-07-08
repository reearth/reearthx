package authserver

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
	"github.com/reearth/reearthx/log"
	"github.com/zitadel/oidc/pkg/op"
)

const (
	loginEndpoint  = "api/login"
	logoutEndpoint = "api/logout"
	jwksEndpoint   = ".well-known/jwks.json"
)

type EndpointConfig struct {
	SubLoader        SubLoader
	Issuer           string
	AuthProviderName string
	URL              *url.URL
	WebURL           *url.URL
	Key              string
	UserInfoSetter   UserInfoSetter
	DefaultClientID  string
	Dev              bool
	DN               *DNConfig
	ConfigRepo       ConfigRepo
	RequestRepo      RequestRepo
}

func (c EndpointConfig) storageConfig() StorageConfig {
	return StorageConfig{
		UserInfoSetter: c.UserInfoSetter,
		Domain:         c.URL.String(),
		ClientDomain:   c.WebURL.String(),
		ClientID:       c.DefaultClientID,
		Dev:            c.Dev,
		DN:             c.DN,
		ConfigRepo:     c.ConfigRepo,
		RequestRepo:    c.RequestRepo,
	}
}

func Endpoint(ctx context.Context, cfg EndpointConfig, r *echo.Group) {
	if cfg.Issuer != "" && !strings.HasSuffix(cfg.Issuer, "/") {
		cfg.Issuer = cfg.Issuer + "/"
	}

	storage, err := NewStorage(ctx, cfg.storageConfig())
	if err != nil {
		log.Fatalf("auth: failed to init: %s\n", err)
	}

	handler, err := op.NewOpenIDProvider(
		ctx,
		&op.Config{
			Issuer:                cfg.Issuer,
			CryptoKey:             sha256.Sum256([]byte(cfg.Key)),
			GrantTypeRefreshToken: true,
		},
		storage,
		op.WithHttpInterceptors(jsonToFormHandler()),
		op.WithHttpInterceptors(setURLVarsHandler()),
		op.WithCustomEndSessionEndpoint(op.NewEndpoint(logoutEndpoint)),
		op.WithCustomKeysEndpoint(op.NewEndpoint(jwksEndpoint)),
	)
	if err != nil {
		log.Fatalf("auth: init failed: %s\n", err)
	}

	router := handler.HttpHandler().(*mux.Router)

	if err := router.Walk(muxToEchoMapper(r)); err != nil {
		log.Fatalf("auth: walk failed: %s\n", err)
	}

	// Actual login endpoint
	r.POST(loginEndpoint, LoginHandler(ctx, LoginHandlerConfig{
		SubLoader: cfg.SubLoader,
		URL:       cfg.URL,
		WebURL:    cfg.WebURL,
		Storage:   storage,
	}))

	r.GET(logoutEndpoint, LogoutHandler())

	// used for auth0/auth0-react; the logout endpoint URL is hard-coded
	// can be removed when the mentioned issue is solved
	// https://github.com/auth0/auth0-spa-js/issues/845
	r.GET("v2/logout", LogoutHandler())

	debugMsg := ""
	if dev, ok := os.LookupEnv(op.OidcDevMode); ok {
		if isDev, _ := strconv.ParseBool(dev); isDev {
			debugMsg = " with debug mode"
		}
	}

	log.Infof("auth: oidc server started%s", debugMsg)
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

func muxToEchoMapper(r *echo.Group) func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		methods, err := route.GetMethods()
		if err != nil {
			r.Any(path, echo.WrapHandler(route.GetHandler()))
			return nil
		}

		for _, method := range methods {
			r.Add(method, path, echo.WrapHandler(route.GetHandler()))
		}

		return nil
	}
}
