package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// PatientHandler service to create handler
type PatientHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.Patient) (*m.Patient, error)
	update       func(*m.Patient) (*m.Patient, error)
	delete       func(doctID *uuid.UUID, patiID uuid.UUID) (*m.Patient, error)
	list         func(doctID *uuid.UUID, f m.FilterPatient) ([]m.Patient, error)
	get          func(doctID *uuid.UUID, patiID uuid.UUID) (*m.Patient, error)
}

type patientResponse struct {
	Item *m.Patient `json:"item"`
	Kind string     `json:"kind"`
}

type patientGetResponse struct {
	dataResponse
	Data patientResponse `json:"data"`
}

type patientDelResponse struct {
	Item *m.Patient `json:"item"`
	Kind string     `json:"kind"`
}

type patientDeleteResponse struct {
	dataResponse
	Data patientDelResponse `json:"data"`
}

type patientsResponse struct {
	collectionItemData
	Items []m.Patient `json:"items"`
	Kind  string      `json:"kind"`
}

type patientsListResponse struct {
	dataResponse
	Data patientsResponse `json:"data"`
}

// Create Patient returns an echo handler
// @Summary Patient.Create
// @Description Create Patient
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Patient body models.Patient true "Create new Patient"
// @Success 200 {object} handler.patientGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/patients [post]
func (handler *PatientHandler) Create(c echo.Context) error {
	req := m.Patient{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	if doctID != nil {
		req.DoctID = *doctID
	}
	pat, err := handler.create(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Patient")
	}
	return c.JSON(http.StatusOK, patientGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: patientResponse{
			Kind: "Patient",
			Item: pat,
		},
	})
}

// Update returns an echo handler
// @Summary Patient.Update
// @Description Update Patient
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param patiID path string true "Patient ID" Format(string)
// @Param Patient body models.Patient true "Patient Update Body"
// @Success 200 {object} handler.patientGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/patients/{patiID} [put]
func (handler *PatientHandler) Update(c echo.Context) error {
	req := m.Patient{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	if doctID != nil {
		req.DoctID = *doctID
	}

	req.PatiID, err = uuid.FromString(c.Param("patiID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	pat, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Patient")
	}
	return c.JSON(http.StatusOK, patientGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: patientResponse{
			Kind: "Patient update",
			Item: pat,
		},
	})
}

// Delete Patient returns an echo handler
// @Summary Patient.Delete
// @Description delete Patient
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param patiID path string true "Delete Patient" Format(string)
// @Success 200 {object} handler.patientDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/patients/{patiID} [del]
func (handler *PatientHandler) Delete(c echo.Context) error {
	patiID, err := uuid.FromString(c.Param("patiID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	pat, err := handler.delete(doctID, patiID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Patient")
	}
	return c.JSON(http.StatusOK, patientDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: patientDelResponse{
			Kind: "Patient deleted",
			Item: pat,
		},
	})
}

// Get returns an echo handler
// @Summary Patient.Get
// @Description Get a Patient
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param patiID query string true "Filter Patients by type [patiID]"
// @Success 200 {object} handler.patientGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/patients/{patiID} [get]
func (handler *PatientHandler) Get(c echo.Context) error {
	patiID, err := uuid.FromString(c.Param("patiID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	pat, err := handler.get(doctID, patiID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Patients")
	}
	return c.JSON(http.StatusOK, patientGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: patientResponse{
			Kind: "Patient get",
			Item: pat,
		},
	})
}

// List returns an echo handler
// @Summary Patient.List
// @Description Get Patient list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param patiID query string false "Filter Patients by type [patiID]"
// @Param doctID query string false "Filter Patients by type [doctID]"
// @Param name query string false "Filter Patients by type [name]"
// @Param email query string false "Filter Patients by type [email]"
// @Param createdAt[gte] query string false "Filter Patients by type [createdAt[gte]]"
// @Param createdAt[lte] query string false "Filter Patients by type [createdAt[lte]]"
// @Success 200 {object} handler.patientsListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/patients [get]
func (handler *PatientHandler) List(c echo.Context) error {
	f, err := buildFilterPatient(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	pat, err := handler.list(doctID, f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Patients")
	}
	return c.JSON(http.StatusOK, patientsListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: patientsResponse{
			Kind:  "Patient list",
			Items: pat,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(pat)),
				TotalItems:       int64(len(pat)),
			},
		},
	})
}

/* buildFilterPatient - Verifying params to method List */
func buildFilterPatient(QueryParam func(string) string) (m.FilterPatient, error) {
	f := m.FilterPatient{}
	patiID := QueryParam("patiID")
	if len(patiID) > 0 {
		f.PatiID = &patiID
	}
	doctID := QueryParam("doctID")
	if len(doctID) > 0 {
		f.DoctID = &doctID
	}
	name := QueryParam("name")
	if len(name) > 0 {
		f.Name = &name
	}
	email := QueryParam("email")
	if len(email) > 0 {
		f.Email = &email
	}
	df := QueryParam("createdAt[gte]")
	if len(df) > 0 {
		dateFrom, err := time.Parse("2006-01-02", df)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse dateFrom")
		}
		f.InitialDate = &dateFrom
	}
	dt := QueryParam("createdAt[lte]")
	if len(dt) > 0 {
		dateTo, err := time.Parse("2006-01-02", dt)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse dateTo")
		}
		dateTo = dateTo.Add((time.Hour * 23) + (time.Millisecond * 60000) - 1)
		f.FinishDate = &dateTo
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
