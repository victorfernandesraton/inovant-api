package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"

	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/perm"
)

// ScheduleHandler service to create handler
type ScheduleHandler struct {
	rolesCtxKey     string
	claimsCtxKey    string
	create          func(*m.Schedule) (*m.Schedule, error)
	update          func(*m.Schedule) (*m.Schedule, error)
	updateSchedule  func(*m.Schedule) (*m.Schedule, error)
	delete          func(scheID uuid.UUID, doctID *uuid.UUID) (*m.Schedule, error)
	updateDelete    func(scheID uuid.UUID) (*m.Schedule, error)
	list            func(doctID *uuid.UUID, f m.FilterSchedule) ([]m.Schedule, error)
	get             func(doctID *uuid.UUID, scheID uuid.UUID) (*m.Schedule, error)
	calendar        func(doctID *uuid.UUID, f m.FilterCalendar) ([]m.Calendar, error)
	outdoor         func(roomID uuid.UUID) (*m.Outdoor, error)
	getErrorMessage func(error) generalError
}

type scheduleResponse struct {
	Item *m.Schedule `json:"item"`
	Kind string      `json:"kind"`
}

type scheduleGetResponse struct {
	dataResponse
	Data scheduleResponse `json:"data"`
}
type outdoorResponse struct {
	Item *m.Outdoor `json:"item"`
	Kind string     `json:"kind"`
}

type outdoorGetResponse struct {
	dataResponse
	Data outdoorResponse `json:"data"`
}

type scheduleDelResponse struct {
	Item *m.Schedule `json:"item"`
	Kind string      `json:"kind"`
}

type scheduleDeleteResponse struct {
	dataResponse
	Data scheduleDelResponse `json:"data"`
}

type schedulesResponse struct {
	collectionItemData
	Items []m.Schedule `json:"items"`
	Kind  string       `json:"kind"`
}

type schedulesListResponse struct {
	dataResponse
	Data schedulesResponse `json:"data"`
}

type calendarResponse struct {
	collectionItemData
	Items []m.Calendar `json:"items"`
	Kind  string       `json:"kind"`
}

type calendarListResponse struct {
	dataResponse
	Data calendarResponse `json:"data"`
}

// Create Schedule returns an echo handler
// @Summary Schedule.Create
// @Description Create Schedule
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Schedule body models.Schedule true "Create new Schedule"
// @Success 200 {object} handler.scheduleGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules [post]
func (handler *ScheduleHandler) Create(c echo.Context) error {
	req := m.Schedule{}
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

	sch, err := handler.create(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: handler.getErrorMessage(err),
		})
	}
	return c.JSON(http.StatusOK, scheduleGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: scheduleResponse{
			Kind: "Schedule",
			Item: sch,
		},
	})
}

// Update returns an echo handler
// @Summary Schedule.Update
// @Description Update Schedule
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param scheID path string true "Schedule ID" Format(string)
// @Param Schedule body models.Schedule true "Schedule Update Body"
// @Success 200 {object} handler.scheduleGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules/{scheID} [put]
func (handler *ScheduleHandler) Update(c echo.Context) error {
	req := m.Schedule{}
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

	req.ScheID, err = uuid.FromString(c.Param("scheID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doc, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Schedule")
	}
	return c.JSON(http.StatusOK, scheduleGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: scheduleResponse{
			Kind: "Schedule update",
			Item: doc,
		},
	})
}

// UpdateSchedule returns an echo handler
// @Summary Schedule.UpdateSchedule
// @Description UpdateSchedule Schedule
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param scheID path string true "Schedule ID" Format(string)
// @Param Schedule body models.Schedule true "Schedule UpdateSchedule Body"
// @Success 200 {object} handler.scheduleGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules/{scheID}/schedule [put]
func (handler *ScheduleHandler) UpdateSchedule(c echo.Context) error {
	req := m.Schedule{}
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

	req.ScheID, err = uuid.FromString(c.Param("scheID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doc, err := handler.updateSchedule(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Schedule")
	}
	return c.JSON(http.StatusOK, scheduleGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: scheduleResponse{
			Kind: "Schedule update",
			Item: doc,
		},
	})
}

// Delete Schedule returns an echo handler
// @Summary Schedule.Delete
// @Description delete Schedule
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param scheID path string true "Delete Schedule" Format(string)
// @Success 200 {object} handler.scheduleDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules/{scheID} [del]
func (handler *ScheduleHandler) Delete(c echo.Context) error {
	scheID, err := uuid.FromString(c.Param("scheID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	doc, err := handler.delete(scheID, doctID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Schedule")
	}
	return c.JSON(http.StatusOK, scheduleDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: scheduleDelResponse{
			Kind: "Schedule deleted",
			Item: doc,
		},
	})
}

// UpdateDeleter Schedule returns an echo handler
// @Summary Schedule.UpdateDelete
// @Description updateDelete Schedule
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param scheID path string true "UpdateDelete Schedule" Format(string)
// @Success 200 {object} handler.scheduleDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules/{scheID}/deletedAt [put]
func (handler *ScheduleHandler) UpdateDeleter(c echo.Context) error {
	scheID, err := uuid.FromString(c.Param("scheID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doc, err := handler.updateDelete(scheID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Schedule")
	}
	return c.JSON(http.StatusOK, scheduleDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: scheduleDelResponse{
			Kind: "Schedule deleted",
			Item: doc,
		},
	})
}

// Get returns an echo handler
// @Summary Schedule.Get
// @Description Get a Schedule
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param scheID query string true "Filter Schedules by type [scheID]"
// @Success 200 {object} handler.scheduleGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules/{scheID} [get]
func (handler *ScheduleHandler) Get(c echo.Context) error {
	scheID, err := uuid.FromString(c.Param("scheID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	doc, err := handler.get(doctID, scheID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Schedules")
	}
	return c.JSON(http.StatusOK, scheduleGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: scheduleResponse{
			Kind: "Schedule get",
			Item: doc,
		},
	})
}

// List returns an echo handler
// @Summary Schedule.List
// @Description Get Schedule list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param scheID query string false "Filter Schedules by type [scheID]"
// @Param doctID query string false "Filter Schedules by type [doctID]"
// @Param roomID query string false "Filter Schedules by type [roomID]"
// @Param startAt query string false "Filter Schedules by type [startAt]"
// @Param endAt query string false "Filter Schedules by type [endAt]"
// @Param plan query string false "Filter Schedules by type [plan]"
// @Success 200 {object} handler.schedulesListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/schedules [get]
func (handler *ScheduleHandler) List(c echo.Context) error {
	f, err := buildFilterSchedule(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}

	sch, err := handler.list(doctID, f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Schedule")
	}
	return c.JSON(http.StatusOK, schedulesListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: schedulesResponse{
			Kind:  "Schedule list",
			Items: sch,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(sch)),
				TotalItems:       int64(len(sch)),
			},
		},
	})
}

// Calendar returns an echo handler
// @Summary Schedule.Calendar
// @Description Get Schedule Calendar
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Success 200 {object} handler.calendarListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/calendar [get]
func (handler *ScheduleHandler) Calendar(c echo.Context) error {
	f, err := buildFilterCalendar(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}
	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	cal, err := handler.calendar(doctID, f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Schedule Calendar")
	}
	return c.JSON(http.StatusOK, calendarListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: calendarResponse{
			Kind:  "Schedule Calendar list",
			Items: cal,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(cal)),
				TotalItems:       int64(len(cal)),
			},
		},
	})
}

// Outdoor returns an echo handler
// @Summary Schedule.Outdoor
// @Description Get Schedule Outdoor
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Success 200 {object} handler.outdoorListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/outdoor/roomID [get]
func (handler *ScheduleHandler) Outdoor(c echo.Context) error {
	roomID := uuid.FromStringOrNil(c.Param("roomID"))
	/*claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}*/
	p, err := auth.ExtractPermissions(c.Get(handler.rolesCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse permissions")
	}
	// User can get his own data
	if !p.Can(perm.Outdoor) {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
		})
	}

	i, err := handler.outdoor(roomID)
	if err != nil {
		return errors.Wrap(err, "Failed to get user")
	}
	return c.JSON(http.StatusOK, outdoorGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: outdoorResponse{
			Kind: "Schedule Outdoor",
			Item: i,
		},
	})
}

/* buildFilterSchedule - Verifying params to method List */
func buildFilterSchedule(QueryParam func(string) string) (m.FilterSchedule, error) {
	f := m.FilterSchedule{}
	scheID := QueryParam("scheID")
	if len(scheID) > 0 {
		f.ScheID = &scheID
	}
	doctID := QueryParam("doctID")
	if len(doctID) > 0 {
		f.DoctID = &doctID
	}
	roomID := QueryParam("roomID")
	if len(roomID) > 0 {
		f.RoomID = &roomID
	}
	ds := QueryParam("startAt")
	if len(ds) > 0 {
		dateFrom, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse startAt")
		}
		f.StartAt = &dateFrom
	}
	de := QueryParam("endAt")
	if len(de) > 0 {
		dateTo, err := time.Parse("2006-01-02", de)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse endAt")
		}
		dateTo = dateTo.Add((time.Hour * 23) + (time.Millisecond * 60000) - 1)
		f.EndAt = &dateTo
	}
	plan := QueryParam("plan")
	if len(plan) > 0 {
		f.Plan = &plan
	}
	hour := QueryParam("hour")
	if len(hour) > 0 {
		f.Hour = &hour
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
	fieldOrder := QueryParam("fieldOrder")
	if len(fieldOrder) > 0 {
		f.FieldOrder = &fieldOrder
	}
	typeOrder := QueryParam("typeOrder")
	if len(typeOrder) > 0 {
		f.TypeOrder = &typeOrder
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

/* buildFilterCalendar - Verifying params to method List */
func buildFilterCalendar(QueryParam func(string) string) (m.FilterCalendar, error) {
	f := m.FilterCalendar{}
	doctID := QueryParam("doctID")
	if len(doctID) > 0 {
		f.DoctID = &doctID
	}
	ds := QueryParam("startAt")
	if len(ds) > 0 {
		dateFrom, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse startAt")
		}
		f.StartAt = &dateFrom
	}
	de := QueryParam("endAt")
	if len(de) > 0 {
		dateTo, err := time.Parse("2006-01-02", de)
		if err != nil {
			return f, errors.Wrap(err, "Failed to parse endAt")
		}
		dateTo = dateTo.Add((time.Hour * 23) + (time.Millisecond * 60000) - 1)
		f.EndAt = &dateTo
	}
	patiID := QueryParam("patiID")
	if len(patiID) > 0 {
		f.PatiID = &patiID
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
