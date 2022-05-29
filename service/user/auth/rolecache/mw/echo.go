package middleware

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/rolecache"
)

func EchoMiddleware(rc *rolecache.RoleCache, cfg JWTConfig) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := auth.Extract(c.Get(cfg.TokenCtxKey))
			if err != nil {
				return err
			}
			r, err := rc.GetRoles(claims.UserID)
			if err != nil {
				return errors.Wrap(err, "Error retrieving roles for user "+claims.UserID)
			}
			c.Set(cfg.RolesCtxKey, r)
			return next(c)
		}
	}
}
