package handler

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	um "gitlab.com/falqon/inovantapp/backend/models"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/perm"
)

// UserHandler service to create handler
type UserHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	list         func() ([]um.User, error)
	get          func(uuid.UUID) (*um.User, error)
	update       func(*um.User) (*um.User, error)
	create       func(*um.User, string) (*um.User, error)
	inactive     func(userID uuid.UUID) (*um.User, error)
	active       func(userID uuid.UUID) (*um.User, error)
	setPushToken func(userID uuid.UUID, token string) error
}

type UserCreateModel struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

// List returns an echo handler
// @Summary users.List
// @Description Get user list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Success 200 {object} handler.userListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users [get]
func (handler *UserHandler) List(c echo.Context) error {
	u, err := handler.list()
	if err != nil {
		return errors.Wrap(err, "Fail to list Users")
	}
	return c.JSON(http.StatusOK, userListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: usersResponse{
			Kind:  "User list",
			Items: u,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(u)),
				TotalItems:       int64(len(u)),
			},
		},
	})
}

// Get returns an echo handler
// @Summary users.Get
// @Description Get user
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param userID path string true "get user" Format(uuid)
// @Success 200 {object} handler.userGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users/{userID} [get]
func (handler *UserHandler) Get(c echo.Context) error {
	userID := uuid.FromStringOrNil(c.Param("userID"))
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}
	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	// User can get his own data
	isOwnData := claims.UserID == userID.String()
	if !p.Can(perm.Admin) && !isOwnData {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	i, err := handler.get(userID)
	if err != nil {
		return errors.Wrap(err, "Failed to get user")
	}
	return c.JSON(http.StatusOK, userGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: userResponse{
			Kind: "User",
			Item: i,
		},
	})
}

// Update returns an echo handler
// @Summary users.Update
// @Description Update User
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param userID path string true "get user" Format(uuid)
// @Param credentials body models.User true "Update instution data"
// @Success 200 {object} handler.userGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users/{userID} [put]
func (handler *UserHandler) Update(c echo.Context) error {
	req := um.User{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	userID := uuid.FromStringOrNil(c.Param("userID"))
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}
	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}

	// User can get his own data
	isOwnData := claims.UserID == userID.String()
	if !p.Can(perm.Admin) && !isOwnData {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	i, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update user")
	}
	return c.JSON(http.StatusOK, userGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: userResponse{
			Kind: "User",
			Item: i,
		},
	})
}

// Create user returns an echo handler
// @Summary users.Create
// @Description Create user
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param credentials body models.User true "create new user"
// @Success 200 {object} handler.userGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users [post]
func (handler *UserHandler) Create(c echo.Context) error {
	_, err := auth.Extract(c.Get("user"))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}

	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	if !p.Can(perm.Admin) {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	req := UserCreateModel{}
	err = c.Bind(&req)
	user := um.User{
		Email:    req.Email,
		Password: []byte(req.Password),
		Roles:    req.Roles,
	}
	if err != nil {
		return err
	}

	u, err := handler.create(&user, string(req.Password))
	if err != nil {
		return errors.Wrap(err, "Fail to create new user")
	}
	return c.JSON(http.StatusOK, userGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: userResponse{
			Kind: "User",
			Item: u,
		},
	})
}

// Inactive user returns an echo handler
// @Summary users.Inactive
// @Description inactive user
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param userID path string true "create new user" Format(uuid)
// @Success 200 {object} handler.userGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users/{userID} [del]
func (handler *UserHandler) Inactive(c echo.Context) error {
	userID := uuid.FromStringOrNil(c.Param("userID"))
	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}
	// User can create new user
	isOwnData := claims.UserID == userID.String()
	if !p.Can(perm.Admin) && !isOwnData {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	u, err := handler.inactive(userID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete user")
	}
	return c.JSON(http.StatusOK, userGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: userResponse{
			Kind: "User inactive",
			Item: u,
		},
	})
}

// Active user returns an echo handler
// @Summary users.Active
// @Description active user
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param userID path string true "create new user" Format(uuid)
// @Success 200 {object} handler.userGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users/active/{userID} [put]
func (handler *UserHandler) Active(c echo.Context) error {
	userID := uuid.FromStringOrNil(c.Param("userID"))
	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}
	// User can create new user
	isOwnData := claims.UserID == userID.String()
	if !p.Can(perm.Admin) && !isOwnData {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	u, err := handler.active(userID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete user")
	}
	return c.JSON(http.StatusOK, userGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: userResponse{
			Kind: "User active",
			Item: u,
		},
	})
}

// SetPushToken returns an echo handler
// @Summary users.SetPushToken
// @Description SetPushToken adds/updates the Expo push token
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param requestData body handler.pushTokenBody true "push token"
// @Success 200 {object} handler.pushTokenResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/users/{userID}/push-token [post]
func (handler *UserHandler) SetPushToken(c echo.Context) error {
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}

	req := pushTokenBody{}
	err = c.Bind(&req)
	if err != nil {
		return err
	}
	err = handler.setPushToken(uuid.FromStringOrNil(claims.UserID), req.Token)
	if err != nil {
		return errors.Wrap(err, "Failed to set user push token")
	}
	return c.JSON(http.StatusOK, pushTokenResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: pushTokenData{
			Kind: "PushToken",
			Item: req.Token,
		},
	})
}

type pushTokenBody struct {
	Token string `json:"token"`
}

type pushTokenData struct {
	singleItemData
	Item string `json:"item"`
	Kind string `json:"kind"`
}

type pushTokenResponse struct {
	dataResponse
	Data pushTokenData `json:"data"`
}

type userResponse struct {
	singleItemData
	Item *um.User `json:"item"`
	Kind string   `json:"kind"`
}

type userGetResponse struct {
	dataResponse
	Data userResponse `json:"data"`
}

type usersResponse struct {
	collectionItemData
	Items []um.User `json:"items"`
	Kind  string    `json:"kind"`
}

type userListResponse struct {
	dataResponse
	Data usersResponse `json:"data"`
}
