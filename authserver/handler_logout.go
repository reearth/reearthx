package authserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func LogoutHandler() echo.HandlerFunc {
	return func(ec echo.Context) error {
		u := ec.QueryParam("returnTo")
		return ec.Redirect(http.StatusTemporaryRedirect, u)
	}
}
