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

// AppointmentHandler service to create handler
type AppointmentHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.Appointment, *uuid.UUID) (*m.Appointment, error)
	update       func(*m.Appointment, *uuid.UUID) (*m.Appointment, error)
	delete       func(appoID uuid.UUID) (*m.Appointment, error)
	list         func(doctID *uuid.UUID, f m.FilterAppointment) ([]m.Appointment, error)
	get          func(doctID *uuid.UUID, appoID uuid.UUID) (*m.Appointment, error)
}

type appointmentResponse struct {
	Item *m.Appointment `json:"item"`
	Kind string         `json:"kind"`
}

type appointmentGetResponse struct {
	dataResponse
	Data appointmentResponse `json:"data"`
}

type appointmentDelResponse struct {
	Item *m.Appointment `json:"item"`
	Kind string         `json:"kind"`
}

type appointmentDeleteResponse struct {
	dataResponse
	Data appointmentDelResponse `json:"data"`
}

type appointmentsResponse struct {
	collectionItemData
	Items []m.Appointment `json:"items"`
	Kind  string          `json:"kind"`
}

type appointmentsListResponse struct {
	dataResponse
	Data appointmentsResponse `json:"data"`
}

// Create Appointment returns an echo handler
// @Summary Appointment.Create
// @Description Create Appointment
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Appointment body models.Appointment true "Create new Appointment"
// @Success 200 {object} handler.appointmentGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/appointments [post]
func (handler *AppointmentHandler) Create(c echo.Context) error {
	req := m.Appointment{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	app, err := handler.create(&req, doctID)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Appointment")
	}
	return c.JSON(http.StatusOK, appointmentGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: appointmentResponse{
			Kind: "Appointment",
			Item: app,
		},
	})
}

// Update returns an echo handler
// @Summary Appointment.Update
// @Description Update Appointment
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param appoID path string true "Appointment ID" Format(string)
// @Param Appointment body models.Appointment true "Appointment Update Body"
// @Success 200 {object} handler.appointmentGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/appointments/{appoID} [put]
func (handler *AppointmentHandler) Update(c echo.Context) error {
	req := m.Appointment{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	req.AppoID, err = uuid.FromString(c.Param("appoID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	app, err := handler.update(&req, doctID)
	if err != nil {
		return errors.Wrap(err, "Fail to update Appointment")
	}
	return c.JSON(http.StatusOK, appointmentGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: appointmentResponse{
			Kind: "Appointment update",
			Item: app,
		},
	})
}

// Delete Appointment returns an echo handler
// @Summary Appointment.Delete
// @Description delete Appointment
// @Accept json
// @Produce json
// @Param context query string false "Context to return"
// @Param appoID path string true "Delete Appointment" Format(string)
// @Success 200 {object} handler.appointmentDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/appointments/{appoID} [del]
func (handler *AppointmentHandler) Delete(c echo.Context) error {
	appoID, err := uuid.FromString(c.Param("appoID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	app, err := handler.delete(appoID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Appointment")
	}
	return c.JSON(http.StatusOK, appointmentDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: appointmentDelResponse{
			Kind: "Appointment deleted",
			Item: app,
		},
	})
}

// Get returns an echo handler
// @Summary Appointment.Get
// @Description Get a Appointment
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param appoID query string true "Filter Appointments by type [appoID]"
// @Success 200 {object} handler.appointmentGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/appointments/{appoID} [get]
func (handler *AppointmentHandler) Get(c echo.Context) error {
	appoID, err := uuid.FromString(c.Param("appoID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	app, err := handler.get(doctID, appoID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Appointments")
	}
	return c.JSON(http.StatusOK, appointmentGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: appointmentResponse{
			Kind: "Appointment get",
			Item: app,
		},
	})
}

// List returns an echo handler
// @Summary Appointment.List
// @Description Get Appointment list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param appoID query string false "Filter Appointments by type [appoID]"
// @Param GStartAt query string false "Filter Appointments by type [GstartAt]"
// @Param LStartAt query string false "Filter Appointments by type [LstartAt]"
// @Param scheID query string false "Filter Appointments by type [scheID]"
// @Param patiID query string false "Filter Appointments by type [patiID]"
// @Param type query string false "Filter Appointments by type [type]"
// @Param status query string false "Filter Appointments by type [status]"
// @Param createdAt[gte] query string false "Filter Appointments by type [createdAt[gte]]"
// @Param createdAt[lte] query string false "Filter Appointments by type [createdAt[lte]]"
// @Success 200 {object} handler.appointmentsListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/appointments [get]
func (handler *AppointmentHandler) List(c echo.Context) error {
	f, err := buildFilterAppointment(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	app, err := handler.list(doctID, f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Appointments")
	}
	return c.JSON(http.StatusOK, appointmentsListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: appointmentsResponse{
			Kind:  "Appointment list",
			Items: app,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(app)),
				TotalItems:       int64(len(app)),
			},
		},
	})
}

/* buildFilterAppointment - Verifying params to method List */
func buildFilterAppointment(QueryParam func(string) string) (m.FilterAppointment, error) {
	f := m.FilterAppointment{}
	appoID := QueryParam("appoID")
	if len(appoID) > 0 {
		f.AppoID = &appoID
	}
	ds := QueryParam("startAt[gte]")
	if len(ds) > 0 {
		dateFrom, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse startAt")
		}
		f.StartAtGte = &dateFrom
	}
	de := QueryParam("startAt[lte]")
	if len(de) > 0 {
		dateTo, err := time.Parse("2006-01-02", de)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse startAt")
		}
		dateTo = dateTo.Add((time.Hour * 23) + (time.Millisecond * 60000) - 1)
		f.StartAtLte = &dateTo
	}
	scheID := QueryParam("scheID")
	if len(scheID) > 0 {
		f.ScheID = &scheID
	}
	patiID := QueryParam("patiID")
	if len(patiID) > 0 {
		f.PatiID = &patiID
	}
	typed := QueryParam("type")
	if len(typed) > 0 {
		f.Type = &typed
	}
	fieldOrder := QueryParam("fieldOrder")
	if len(fieldOrder) > 0 {
		f.FieldOrder = &fieldOrder
	}
	typeOrder := QueryParam("typeOrder")
	if len(typeOrder) > 0 {
		f.TypeOrder = &typeOrder
	}
	status := QueryParam("status")
	if len(status) > 0 {
		f.Status = &status
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
