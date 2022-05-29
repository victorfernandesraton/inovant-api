package tokenauth

import (
	"net/http"

	"github.com/labstack/echo"
)

//TokenAuth checks if request has a token
func TokenAuth(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqToken := c.Request().Header.Get(echo.HeaderAuthorization)
			he := echo.NewHTTPError(http.StatusUnauthorized)
			if reqToken == token {
				return next(c)
			}
			return he
		}
	}
}
