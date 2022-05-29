package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/perm"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// ConfigHandler service to create handler
type ConfigHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.Config) (*m.Config, error)
	update       func(*m.Config) (*m.Config, error)
	delete       func(key string) (*m.Config, error)
	list         func(m.FilterConfig) ([]m.Config, error)
	get          func(key string) (*m.Config, error)
}

type configResponse struct {
	Item *m.Config `json:"item"`
	Kind string    `json:"kind"`
}

type configGetResponse struct {
	dataResponse
	Data configResponse `json:"data"`
}

type configDelResponse struct {
	Item *m.Config `json:"item"`
	Kind string    `json:"kind"`
}

type configDeleteResponse struct {
	dataResponse
	Data configDelResponse `json:"data"`
}

type configsResponse struct {
	collectionItemData
	Items []m.Config `json:"items"`
	Kind  string     `json:"kind"`
}

type configsListResponse struct {
	dataResponse
	Data configsResponse `json:"data"`
}

// Create Config returns an echo handler
// @Summary Config.Create
// @Description Create Config
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Config body models.Config true "Create new Config"
// @Success 200 {object} handler.configGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/configs [post]
func (handler *ConfigHandler) Create(c echo.Context) error {
	req := m.Config{}

	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	if !p.Can(perm.Admin) && !p.Can(perm.Secretary) {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	err = c.Bind(&req)
	if err != nil {
		return err
	}
	con, err := handler.create(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Config")
	}
	return c.JSON(http.StatusOK, configGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: configResponse{
			Kind: "Config",
			Item: con,
		},
	})
}

// Update returns an echo handler
// @Summary Config.Update
// @Description Update Config
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param key path string true "Config key" Format(string)
// @Param Config body models.Config true "Config Update Body"
// @Success 200 {object} handler.configGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/configs/{key} [put]
func (handler *ConfigHandler) Update(c echo.Context) error {
	req := m.Config{}

	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	if !p.Can(perm.Admin) && !p.Can(perm.Secretary) {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	err = c.Bind(&req)
	if err != nil {
		return err
	}
	req.Key = c.Param("key")
	con, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Config")
	}
	return c.JSON(http.StatusOK, configGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: configResponse{
			Kind: "Config update",
			Item: con,
		},
	})
}

// Delete Config returns an echo handler
// @Summary Config.Delete
// @Description   delete Config
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param key path string true "Delete Config" Format(string)
// @Success 200 {object} handler.configDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/configs/{key} [del]
func (handler *ConfigHandler) Delete(c echo.Context) error {
	key := c.Param("key")
	con, err := handler.delete(key)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Config")
	}
	return c.JSON(http.StatusOK, configDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: configDelResponse{
			Kind: "Config deleted",
			Item: con,
		},
	})
}

// Get returns an echo handler
// @Summary Config.Get
// @Description Get a Config
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param key query string true "Filter Configs by type [key]"
// @Success 200 {object} handler.configGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/configs/{key} [get]
func (handler *ConfigHandler) Get(c echo.Context) error {
	key := c.Param("key")
	con, err := handler.get(key)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Configs")
	}
	return c.JSON(http.StatusOK, configGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: configResponse{
			Kind: "Config get",
			Item: con,
		},
	})
}

// List returns an echo handler
// @Summary Config.List
// @Description Get Config list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param key query string false "Filter Configs by type [key]"
// @Success 200 {object} handler.configsListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/configs [get]
func (handler *ConfigHandler) List(c echo.Context) error {
	f, err := buildFilterConfig(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	con, err := handler.list(f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Configs")
	}
	return c.JSON(http.StatusOK, configsListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: configsResponse{
			Kind:  "Config list",
			Items: con,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(con)),
				TotalItems:       int64(len(con)),
			},
		},
	})
}

/* buildFilterConfig - Verifying params to method List */
func buildFilterConfig(QueryParam func(string) string) (m.FilterConfig, error) {
	f := m.FilterConfig{}
	key := QueryParam("key")
	if len(key) > 0 {
		f.Key = &key
	}
	l := QueryParam("limit")
	if len(l) > 0 {
		limit, err := strconv.ParseInt(l, 10, 64)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse limit: "+l)
		}
		o := (int64)(limit)
		f.Limit = &o
	}
	s := QueryParam("offset")
	if len(s) > 0 {
		offset, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse offset: "+s)
		}
		o := (int64)(offset)
		f.Offset = &o
	}
	return f, nil
}
