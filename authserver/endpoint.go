package authserver

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
	"github.com/reearth/reearthx/log"
	"github.com/zitadel/oidc/pkg/op"
)

const (
	loginEndpoint  = "api/login"
	logoutEndpoint = "api/logout"
)

type EndpointConfig struct {
	Issuer          string
	URL             *url.URL
	WebURL          *url.URL
	Key             string
	DefaultClientID string
	Dev             bool
	DN              *DNConfig
	UserRepo        UserRepo
	ConfigRepo      ConfigRepo
	RequestRepo     RequestRepo
	DenyHTTP        bool
}

func (c *EndpointConfig) normalize() {
	if c.Issuer == "" {
		c.Issuer = c.URL.String()
	}

	if strings.HasPrefix(c.Issuer, "http://") && !c.DenyHTTP {
		c.Dev = true
	}

	if !strings.HasSuffix(c.Issuer, "/") {
		c.Issuer = c.Issuer + "/"
	}

	if c.Dev {
		os.Setenv(op.OidcDevMode, "true")
	} else {
		os.Unsetenv(op.OidcDevMode)
	}
}

func (c EndpointConfig) storageConfig() StorageConfig {
	return StorageConfig{
		UserInfoSetter: c.UserRepo.Info,
		Domain:         c.URL.String(),
		ClientDomain:   c.WebURL.String(),
		ClientID:       c.DefaultClientID,
		Dev:            c.Dev,
		DN:             c.DN,
		ConfigRepo:     c.ConfigRepo,
		RequestRepo:    c.RequestRepo,
	}
}

func Endpoint(ctx context.Context, cfg EndpointConfig, g *echo.Group) {
	cfg.normalize()

	storage, err := NewStorage(ctx, cfg.storageConfig(), cfg.Issuer)
	if err != nil {
		log.Fatalf("auth: storage init failed: %s\n", err)
	}

	router, err := Server(ctx, ServerConfig{
		Issuer:  cfg.Issuer,
		Key:     cfg.Key,
		Storage: storage,
	})
	if err != nil {
		log.Fatalf("auth: server init failed: %s\n", err)
	}

	if err := router.Walk(muxToEchoMapper(g)); err != nil {
		log.Fatalf("auth: walk failed: %s\n", err)
	}

	g.POST(loginEndpoint, LoginHandler(ctx, LoginHandlerConfig{
		SubLoader: cfg.UserRepo.Sub,
		URL:       cfg.URL,
		WebURL:    cfg.WebURL,
		Storage:   storage,
	}))

	g.GET(logoutEndpoint, LogoutHandler())

	// compatibility with auth0/auth0-spa-js; the logout endpoint URL is hard-coded
	// https://github.com/auth0/auth0-spa-js/issues/845
	g.GET("v2/logout", LogoutHandler())

	debugMsg := ""
	if cfg.Dev {
		debugMsg = " with dev mode"
	}

	log.Infof("auth: oidc server started%s", debugMsg)
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
