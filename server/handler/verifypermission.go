package handler

import (
	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/perm"
)

func doctIDOrNil(c echo.Context, claimsCtxKey, rolesCtxKey string) (*uuid.UUID, error) {
	p, err := auth.ExtractPermissions(c.Get(rolesCtxKey))
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't parse permissions")
	}
	isSuperAdmin := false
	if p.Can(perm.Admin) || p.Can(perm.Secretary) || p.Can(perm.Config) {
		isSuperAdmin = true
	}
	claims, err := auth.Extract(c.Get(claimsCtxKey))
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't parse token")
	}
	claimsDoctID, err := uuid.FromString(claims.DoctID)
	if err != nil {
		if isSuperAdmin {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Fail to find doctID id")
	}
	return &claimsDoctID, nil
}
