package handler

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"

	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/perm"
)

// DoctorHandler service to create handler
type DoctorHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.Doctor) (*m.Doctor, error)
	update       func(*m.Doctor) (*m.Doctor, error)
	delete       func(doctID uuid.UUID) (*m.Doctor, error)
	list         func(doctID *uuid.UUID, f m.FilterDoctor) ([]m.Doctor, error)
	get          func(doctID uuid.UUID) (*m.Doctor, error)
}

type doctorResponse struct {
	Item *m.Doctor `json:"item"`
	Kind string    `json:"kind"`
}

type doctorGetResponse struct {
	dataResponse
	Data doctorResponse `json:"data"`
}

type doctorDelResponse struct {
	Item *m.Doctor `json:"item"`
	Kind string    `json:"kind"`
}

type doctorDeleteResponse struct {
	dataResponse
	Data doctorDelResponse `json:"data"`
}

type doctorsResponse struct {
	collectionItemData
	Items []m.Doctor `json:"items"`
	Kind  string     `json:"kind"`
}

type doctorsListResponse struct {
	dataResponse
	Data doctorsResponse `json:"data"`
}

// Create Doctor returns an echo handler
// @Summary Doctor.Create
// @Description Create Doctor
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Doctor body models.Doctor true "Create new Doctor"
// @Success 200 {object} handler.doctorGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors [post]
func (handler *DoctorHandler) Create(c echo.Context) error {
	req := struct {
		m.Doctor
		Password string `json:"password"`
	}{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	// User can create new user
	if !p.Can(perm.Admin) && !p.Can(perm.Secretary) {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	req.Doctor.Password = []byte(req.Password)
	doc, err := handler.create(&req.Doctor)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Doctor")
	}
	return c.JSON(http.StatusOK, doctorGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorResponse{
			Kind: "Doctor",
			Item: doc,
		},
	})
}

// Update returns an echo handler
// @Summary Doctor.Update
// @Description Update Doctor
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID path string true "Doctor ID" Format(string)
// @Param Doctor body models.Doctor true "Doctor Update Body"
// @Success 200 {object} handler.doctorGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors/{doctID} [put]
func (handler *DoctorHandler) Update(c echo.Context) error {
	req := m.Doctor{}
	err := c.Bind(&req)
	if err != nil {
		return errors.Wrap(err, "parse body")
	}

	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}
	claimsDoctID := uuid.FromStringOrNil(claims.DoctID)

	if !p.Can(perm.Admin) && !p.Can(perm.Secretary) {
		if claimsDoctID != req.DoctID {
			return c.JSON(http.StatusUnauthorized, errorResponse{
				Error: generalError{
					Code:    http.StatusUnauthorized,
					Message: "Unauthorized",
				},
			})
		}
	}

	req.DoctID, err = uuid.FromString(c.Param("doctID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	if doctID != nil {
		req.DoctID = *doctID
	}

	doc, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Doctor")
	}
	return c.JSON(http.StatusOK, doctorGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorResponse{
			Kind: "Doctor update",
			Item: doc,
		},
	})
}

// Delete Doctor returns an echo handler
// @Summary Doctor.Delete
// @Description   delete Doctor
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID path string true "Delete Doctor" Format(string)
// @Success 200 {object} handler.doctorDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors/{doctID} [del]
func (handler *DoctorHandler) Delete(c echo.Context) error {
	doctID, err := uuid.FromString(c.Param("doctID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	// User can create new user
	if !p.Can(perm.Admin) && !p.Can(perm.Secretary) {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	doc, err := handler.delete(doctID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Doctor")
	}
	return c.JSON(http.StatusOK, doctorDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorDelResponse{
			Kind: "Doctor deleted",
			Item: doc,
		},
	})
}

// Get returns an echo handler
// @Summary Doctor.Get
// @Description Get a Doctor
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID query string false "Filter Doctors by type [doctID]"
// @Success 200 {object} handler.doctorGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors/{doctID} [get]
func (handler *DoctorHandler) Get(c echo.Context) error {
	doctID, err := uuid.FromString(c.Param("doctID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	claimsdoctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	if claimsdoctID != nil {
		doctID = *claimsdoctID
	}

	doc, err := handler.get(doctID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Doctors")
	}
	return c.JSON(http.StatusOK, doctorGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorResponse{
			Kind: "Doctor get",
			Item: doc,
		},
	})
}

// List returns an echo handler
// @Summary Doctor.List
// @Description Get Doctor list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param doctID query string false "Filter Doctors by type [doctID]"
// @Param userID query string false "Filter Doctors by type [userID]"
// @Param name query string false "Filter Doctors by type [name]"
// @Success 200 {object} handler.doctorsListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/doctors [get]
func (handler *DoctorHandler) List(c echo.Context) error {
	f, err := buildFilterDoctor(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	doc, err := handler.list(doctID, f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Doctor")
	}
	return c.JSON(http.StatusOK, doctorsListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: doctorsResponse{
			Kind:  "Doctor list",
			Items: doc,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(doc)),
				TotalItems:       int64(len(doc)),
			},
		},
	})
}

/* buildFilterDoctor - Verifying params to method List */
func buildFilterDoctor(QueryParam func(string) string) (m.FilterDoctor, error) {
	f := m.FilterDoctor{}
	doctID := QueryParam("doctID")
	if len(doctID) > 0 {
		f.DoctID = &doctID
	}
	userID := QueryParam("userID")
	if len(userID) > 0 {
		f.UserID = &userID
	}
	name := QueryParam("name")
	if len(name) > 0 {
		f.Name = &name
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
