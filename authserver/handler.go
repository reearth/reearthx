package authserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/util"
	"github.com/zitadel/oidc/pkg/op"
)

type SubLoader func(ctx context.Context, email, password, authRequestID string) (string, error)

type LoginHandlerConfig struct {
	SubLoader SubLoader
	URL       *url.URL
	WebURL    *url.URL
	Storage   op.Storage
}

type loginForm struct {
	Email         string `json:"username" form:"username"`
	Password      string `json:"password" form:"password"`
	AuthRequestID string `json:"id" form:"id"`
}

func LoginHandler(ctx context.Context, cfg LoginHandlerConfig) func(ctx echo.Context) error {
	return func(ec echo.Context) error {
		request := new(loginForm)
		if err := ec.Bind(request); err != nil {
			log.Errorln("auth: filed to parse login request")
			return ec.Redirect(
				http.StatusFound,
				redirectURL(cfg.WebURL, "/login", "", "Bad request!"),
			)
		}

		if _, err := cfg.Storage.AuthRequestByID(ctx, request.AuthRequestID); err != nil {
			log.Errorf("auth: filed to parse login request: %s\n", err)
			return ec.Redirect(
				http.StatusFound,
				redirectURL(cfg.WebURL, "/login", "", "Bad request!"),
			)
		}

		if len(request.Email) == 0 || len(request.Password) == 0 {
			log.Errorln("auth: one of credentials are not provided")
			return ec.Redirect(
				http.StatusFound,
				redirectURL(cfg.WebURL, "/login", request.AuthRequestID, "Bad request!"),
			)
		}

		// check user credentials from db
		sub, err := cfg.SubLoader(ctx, request.Email, request.Password, request.AuthRequestID)
		if err != nil || sub == "" {
			if err == nil && sub == "" {
				err = errors.New("empty sub")
			}
			log.Errorf("auth: wrong credentials: %s\n", err)
			return ec.Redirect(
				http.StatusFound,
				redirectURL(cfg.WebURL, "/login", request.AuthRequestID, "Login failed; Invalid s ID or password."),
			)
		}

		// Complete the auth request && set the subject
		if err := cfg.Storage.(*Storage).CompleteAuthRequest(ctx, request.AuthRequestID, sub); err != nil {
			log.Errorf("auth: failed to complete the auth request: %s\n", err)
			return ec.Redirect(
				http.StatusFound,
				redirectURL(cfg.WebURL, "/login", request.AuthRequestID, "Bad request!"),
			)
		}

		return ec.Redirect(
			http.StatusFound,
			redirectURL(cfg.URL, "/authorize/callback", request.AuthRequestID, ""),
		)
	}
}

func LogoutHandler() echo.HandlerFunc {
	return func(ec echo.Context) error {
		u := ec.QueryParam("returnTo")
		return ec.Redirect(http.StatusTemporaryRedirect, u)
	}
}

func redirectURL(u *url.URL, p string, requestID, err string) string {
	v := util.CopyURL(u)
	if p != "" {
		v.Path = p
	}
	queryValues := u.Query()
	queryValues.Set("id", requestID)
	if err != "" {
		queryValues.Set("error", err)
	}
	v.RawQuery = queryValues.Encode()
	return v.String()
}
