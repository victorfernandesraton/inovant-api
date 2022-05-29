package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// SpecialtyHandler service to create handler
type SpecialtyHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.Specialty) (*m.Specialty, error)
	update       func(*m.Specialty) (*m.Specialty, error)
	delete       func(specID int64) (*m.Specialty, error)
	list         func(m.FilterSpecialty) ([]m.Specialty, error)
	get          func(specID int64) (*m.Specialty, error)
}

type specialtyResponse struct {
	Item *m.Specialty `json:"item"`
	Kind string       `json:"kind"`
}

type specialtyGetResponse struct {
	dataResponse
	Data specialtyResponse `json:"data"`
}

type specialtyDelResponse struct {
	Item *m.Specialty `json:"item"`
	Kind string       `json:"kind"`
}

type specialtyDeleteResponse struct {
	dataResponse
	Data specialtyDelResponse `json:"data"`
}

type specialtysResponse struct {
	collectionItemData
	Items []m.Specialty `json:"items"`
	Kind  string        `json:"kind"`
}

type specialtysListResponse struct {
	dataResponse
	Data specialtysResponse `json:"data"`
}

// Create Specialty returns an echo handler
// @Summary Specialty.Create
// @Description Create Specialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Specialty body models.Specialty true "Create new Specialty"
// @Success 200 {object} handler.specialtyGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/specialty [post]
func (handler *SpecialtyHandler) Create(c echo.Context) error {
	req := m.Specialty{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	spe, err := handler.create(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Specialty")
	}
	return c.JSON(http.StatusOK, specialtyGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: specialtyResponse{
			Kind: "Specialty",
			Item: spe,
		},
	})
}

// Update returns an echo handler
// @Summary Specialty.Update
// @Description Update Specialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param specID path string true "Specialty ID" Format(string)
// @Param Specialty body models.Specialty true "Specialty Update Body"
// @Success 200 {object} handler.specialtyGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/specialty/{specID} [put]
func (handler *SpecialtyHandler) Update(c echo.Context) error {
	req := m.Specialty{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	req.SpecID, err = strconv.ParseInt(c.Param("specID"), 10, 64)
	if err != nil {
		return errors.Wrap(err, "Error int64 format")
	}

	spe, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Specialty")
	}
	return c.JSON(http.StatusOK, specialtyGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: specialtyResponse{
			Kind: "Specialty update",
			Item: spe,
		},
	})
}

// Delete Specialty returns an echo handler
// @Summary Specialty.Delete
// @Description   delete Specialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param specID path string true "Delete Specialty" Format(string)
// @Success 200 {object} handler.specialtyDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/specialty/{specID} [del]
func (handler *SpecialtyHandler) Delete(c echo.Context) error {
	specID, err := strconv.ParseInt(c.Param("specID"), 10, 64)
	if err != nil {
		return errors.Wrap(err, "Error int64 format")
	}
	spe, err := handler.delete(specID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Specialty")
	}
	return c.JSON(http.StatusOK, specialtyDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: specialtyDelResponse{
			Kind: "Specialty deleted",
			Item: spe,
		},
	})
}

// Get returns an echo handler
// @Summary Specialty.Get
// @Description Get a Specialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param specID query string true "Filter specialty by type [specID]"
// @Success 200 {object} handler.specialtyGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/specialty/{specID} [get]
func (handler *SpecialtyHandler) Get(c echo.Context) error {
	specID, err := strconv.ParseInt(c.Param("specID"), 10, 64)
	if err != nil {
		return errors.Wrap(err, "Error int64 format")
	}

	spe, err := handler.get(specID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Specialty")
	}
	return c.JSON(http.StatusOK, specialtyGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: specialtyResponse{
			Kind: "Specialty get",
			Item: spe,
		},
	})
}

// List returns an echo handler
// @Summary Specialty.List
// @Description Get Specialty list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param specID query string false "Filter Specialty by type [specID]"
// @Param name query string false "Filter Specialty by type [name]"
// @Success 200 {object} handler.specialtysListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/specialty [get]
func (handler *SpecialtyHandler) List(c echo.Context) error {
	f, err := buildFilterSpecialty(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	spe, err := handler.list(f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Specialtys")
	}
	return c.JSON(http.StatusOK, specialtysListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: specialtysResponse{
			Kind:  "Specialty list",
			Items: spe,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(spe)),
				TotalItems:       int64(len(spe)),
			},
		},
	})
}

/* buildFilterSpecialty - Verifying params to method List */
func buildFilterSpecialty(QueryParam func(string) string) (m.FilterSpecialty, error) {
	f := m.FilterSpecialty{}
	name := QueryParam("name")
	if len(name) > 0 {
		f.Name = &name
	}
	description := QueryParam("description")
	if len(description) > 0 {
		f.Description = &description
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
