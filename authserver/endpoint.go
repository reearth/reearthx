package authserver

import (
	"context"
	"net/url"
	"os"
	"strconv"
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
	SubLoader        SubLoader
	Issuer           string
	AuthProviderName string
	URL              *url.URL
	WebURL           *url.URL
	Key              string
	UserInfoProvider UserInfoProvider
	DefaultClientID  string
	Dev              bool
	DN               *DNConfig
	ConfigRepo       ConfigRepo
	RequestRepo      RequestRepo
}

func (c EndpointConfig) storageConfig() StorageConfig {
	return StorageConfig{
		UserInfoSetter: c.UserInfoProvider,
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
	if cfg.Issuer != "" && !strings.HasSuffix(cfg.Issuer, "/") {
		cfg.Issuer = cfg.Issuer + "/"
	}

	storage, err := NewStorage(ctx, cfg.storageConfig())
	if err != nil {
		log.Fatalf("auth: failed to init: %s\n", err)
	}

	router := Server(ctx, ServerConfig{
		Issuer:  cfg.Issuer,
		Key:     cfg.Key,
		Storage: storage,
	}).(*mux.Router)

	if err := router.Walk(muxToEchoMapper(g)); err != nil {
		log.Fatalf("auth: walk failed: %s\n", err)
	}

	g.POST(loginEndpoint, LoginHandler(ctx, LoginHandlerConfig{
		SubLoader: cfg.SubLoader,
		URL:       cfg.URL,
		WebURL:    cfg.WebURL,
		Storage:   storage,
	}))

	g.GET(logoutEndpoint, LogoutHandler())

	// compability with auth0/auth0-spa-js; the logout endpoint URL is hard-coded
	// https://github.com/auth0/auth0-spa-js/issues/845
	g.GET("v2/logout", LogoutHandler())

	debugMsg := ""
	if dev, ok := os.LookupEnv(op.OidcDevMode); ok {
		if isDev, _ := strconv.ParseBool(dev); isDev {
			debugMsg = " with debug mode"
		}
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
