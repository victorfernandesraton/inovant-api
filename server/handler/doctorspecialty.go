package handler

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// DoctorSpecialtyHandler service to create handler
type DoctorSpecialtyHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.DoctorSpecialty) (*m.DoctorSpecialty, error)
	update       func(*m.DoctorSpecialty) (*m.DoctorSpecialty, error)
	delete       func(doctID uuid.UUID, specID int64) (*m.DoctorSpecialty, error)
	list         func(m.FilterDoctorSpecialty) ([]m.DoctorSpecialty, error)
	get          func(doctID uuid.UUID, specID int64) (*m.DoctorSpecialty, error)
}

type doctorSpecialtyResponse struct {
	Item *m.DoctorSpecialty `json:"item"`
	Kind string             `json:"kind"`
}

type doctorSpecialtyGetResponse struct {
	dataResponse
	Data doctorSpecialtyResponse `json:"data"`
}

type doctorSpecialtyDelResponse struct {
	Item *m.DoctorSpecialty `json:"item"`
	Kind string             `json:"kind"`
}

type doctorSpecialtyDeleteResponse struct {
	dataResponse
	Data doctorSpecialtyDelResponse `json:"data"`
}

type doctorSpecialtysResponse struct {
	collectionItemData
	Items []m.DoctorSpecialty `json:"items"`
	Kind  string              `json:"kind"`
}

type doctorSpecialtysListResponse struct {
	dataResponse
	Data doctorSpecialtysResponse `json:"data"`
}

// Create DoctorSpecialty returns an echo handler
// @Summary DoctorSpecialty.Create
// @Description Create DoctorSpecialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param DoctorSpecialty body models.DoctorSpecialty true "Create new DoctorSpecialty"
// @Success 200 {object} handler.doctorSpecialtyGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors-specialty [post]
func (handler *DoctorSpecialtyHandler) Create(c echo.Context) error {
	req := m.DoctorSpecialty{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	spe, err := handler.create(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to create new DoctorSpecialty")
	}
	return c.JSON(http.StatusOK, doctorSpecialtyGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorSpecialtyResponse{
			Kind: "DoctorSpecialty",
			Item: spe,
		},
	})
}

// Update returns an echo handler
// @Summary DoctorSpecialty.Update
// @Description Update DoctorSpecialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID path string true "DoctorSpecialty ID" Format(string)
// @Param DoctorSpecialty body models.DoctorSpecialty true "DoctorSpecialty Update Body"
// @Success 200 {object} handler.doctorSpecialtyGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors-specialty/{doctID} [put]
func (handler *DoctorSpecialtyHandler) Update(c echo.Context) error {
	req := m.DoctorSpecialty{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	req.DoctID, err = uuid.FromString(c.Param("doctID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	spe, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update DoctorSpecialty")
	}
	return c.JSON(http.StatusOK, doctorSpecialtyGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorSpecialtyResponse{
			Kind: "DoctorSpecialty update",
			Item: spe,
		},
	})
}

// Delete DoctorSpecialty returns an echo handler
// @Summary DoctorSpecialty.Delete
// @Description   delete DoctorSpecialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID and specID path string true "Delete DoctorSpecialty" Format(string)
// @Success 200 {object} handler.doctorSpecialtyDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors-specialty/{doctID}/{specID} [del]
func (handler *DoctorSpecialtyHandler) Delete(c echo.Context) error {
	doctID, err := uuid.FromString(c.Param("doctID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}
	specID, err := strconv.ParseInt(c.Param("specID"), 10, 64)
	if err != nil {
		return errors.Wrap(err, "Error int64 format")
	}
	spe, err := handler.delete(doctID, specID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete DoctorSpecialty")
	}
	return c.JSON(http.StatusOK, doctorSpecialtyDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorSpecialtyDelResponse{
			Kind: "DoctorSpecialty deleted",
			Item: spe,
		},
	})
}

// Get returns an echo handler
// @Summary DoctorSpecialty.Get
// @Description Get a DoctorSpecialty
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID query string true "Filter DoctorSpecialty by type [doctID]"
// @Success 200 {object} handler.doctorSpecialtyGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors-specialty/{doctID}/{specID} [get]
func (handler *DoctorSpecialtyHandler) Get(c echo.Context) error {
	doctID, err := uuid.FromString(c.Param("doctID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}
	specID, err := strconv.ParseInt(c.Param("specID"), 10, 64)
	if err != nil {
		return errors.Wrap(err, "Error int64 format")
	}

	spe, err := handler.get(doctID, specID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Specialty")
	}
	return c.JSON(http.StatusOK, doctorSpecialtyGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorSpecialtyResponse{
			Kind: "DoctorSpecialty get",
			Item: spe,
		},
	})
}

// List returns an echo handler
// @Summary DoctorSpecialty.List
// @Description Get DoctorSpecialty list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID query string false "Filter DoctorSpecialty by type [doctID]"
// @Param name query string false "Filter DoctorSpecialty by type [name]"
// @Success 200 {object} handler.doctorSpecialtysListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors-specialty [get]
func (handler *DoctorSpecialtyHandler) List(c echo.Context) error {
	f, err := buildFilterDoctorSpecialty(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	spe, err := handler.list(f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of DoctorSpecialty")
	}
	return c.JSON(http.StatusOK, doctorSpecialtysListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorSpecialtysResponse{
			Kind:  "DoctorSpecialty list",
			Items: spe,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(spe)),
				TotalItems:       int64(len(spe)),
			},
		},
	})
}

/* buildFilterDoctorSpecialty - Verifying params to method List */
func buildFilterDoctorSpecialty(QueryParam func(string) string) (m.FilterDoctorSpecialty, error) {
	f := m.FilterDoctorSpecialty{}
	doctID := QueryParam("doctID")
	if len(doctID) > 0 {
		f.DoctID = &doctID
	}
	specID := QueryParam("specID")
	if len(specID) > 0 {
		pSpecID, err := strconv.ParseInt(specID, 10, 64)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse spec_id: "+specID)
		}
		o := (int64)(pSpecID)
		f.SpecID = &o
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
