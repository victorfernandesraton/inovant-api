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

// ActionVerificationHandler service to create handler
type ActionVerificationHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.ActionVerification) (*m.ActionVerification, error)
	update       func(*m.ActionVerification) (*m.ActionVerification, error)
	delete       func(acveID uuid.UUID) (*m.ActionVerification, error)
	list         func(m.FilterActionVerification) ([]m.ActionVerification, error)
	get          func(acveID uuid.UUID) (*m.ActionVerification, error)
}

type actionVerificationResponse struct {
	Item *m.ActionVerification `json:"item"`
	Kind string                `json:"kind"`
}

type actionVerificationGetResponse struct {
	dataResponse
	Data actionVerificationResponse `json:"data"`
}

type actionVerificationDelResponse struct {
	Item *m.ActionVerification `json:"item"`
	Kind string                `json:"kind"`
}

type actionVerificationDeleteResponse struct {
	dataResponse
	Data actionVerificationDelResponse `json:"data"`
}

type actionsVerificationResponse struct {
	collectionItemData
	Items []m.ActionVerification `json:"items"`
	Kind  string                 `json:"kind"`
}

type actionsVerificationListResponse struct {
	dataResponse
	Data actionsVerificationResponse `json:"data"`
}

// Create ActionVerification returns an echo handler
// @Summary ActionVerification.Create
// @Description Create ActionVerification
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param ActionVerification body models.ActionVerification true "Create new ActionVerification"
// @Success 200 {object} handler.actionVerificationGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/actions-verification [post]
func (handler *ActionVerificationHandler) Create(c echo.Context) error {
	req := m.ActionVerification{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	acv, err := handler.create(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Action Verification")
	}
	return c.JSON(http.StatusOK, actionVerificationGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: actionVerificationResponse{
			Kind: "Action Verification",
			Item: acv,
		},
	})
}

// Update returns an echo handler
// @Summary ActionVerification.Update
// @Description Update ActionVerification
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param acveID path string true "ActionVerification ID" Format(string)
// @Param ActionVerification body models.ActionVerification true "ActionVerification Update Body"
// @Success 200 {object} handler.actionVerificationGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/actions-verification/{acveID} [put]
func (handler *ActionVerificationHandler) Update(c echo.Context) error {
	req := m.ActionVerification{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	req.AcveID, err = uuid.FromString(c.Param("acveID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	acv, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Action Verification")
	}
	return c.JSON(http.StatusOK, actionVerificationGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: actionVerificationResponse{
			Kind: "Action Verification updated",
			Item: acv,
		},
	})
}

// Delete ActionVerification returns an echo handler
// @Summary ActionVerification.Delete
// @Description delete ActionVerification
// @Accept json
// @Produce json
// @Param context query string false "Context to return"
// @Param acveID path string true "Delete ActionVerification" Format(string)
// @Success 200 {object} handler.actionVerificationDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/actions-verification/{acveID} [del]
func (handler *ActionVerificationHandler) Delete(c echo.Context) error {
	acveID, err := uuid.FromString(c.Param("acveID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}
	acv, err := handler.delete(acveID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete ActionVerification")
	}
	return c.JSON(http.StatusOK, actionVerificationDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: actionVerificationDelResponse{
			Kind: "Action Verification deleted",
			Item: acv,
		},
	})
}

// Get returns an echo handler
// @Summary ActionVerification.Get
// @Description Get a ActionVerification
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param acveID query string true "Filter ActionVerifications by type [acveID]"
// @Success 200 {object} handler.actionVerificationGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/actions-verification/{acveID} [get]
func (handler *ActionVerificationHandler) Get(c echo.Context) error {
	acveID, err := uuid.FromString(c.Param("acveID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	acv, err := handler.get(acveID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of ActionVerifications")
	}
	return c.JSON(http.StatusOK, actionVerificationGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: actionVerificationResponse{
			Kind: "Action Verification get",
			Item: acv,
		},
	})
}

// List returns an echo handler
// @Summary ActionVerification.List
// @Description Get ActionVerification list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param acveID query string false "Filter ActionVerifications by type [acveID]"
// @Param userID query string false "Filter ActionVerifications by type [userID]"
// @Param type query string false "Filter ActionVerifications by type [type]"
// @Param verification query string false "Filter ActionVerifications by type [verification]"
// @Param createdAt[gte] query string false "Filter ActionVerifications by type [createdAt[gte]]"
// @Param createdAt[lte] query string false "Filter ActionVerifications by type [createdAt[lte]]"
// @Success 200 {object} handler.actionsVerificationListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/actions-verification [get]
func (handler *ActionVerificationHandler) List(c echo.Context) error {
	f, err := buildFilterActionVerification(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	acv, err := handler.list(f)
	if err != nil {
		return errors.Wrap(err, "Fail list of Actions Verification")
	}
	return c.JSON(http.StatusOK, actionsVerificationListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: actionsVerificationResponse{
			Kind:  "Action Verification list",
			Items: acv,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(acv)),
				TotalItems:       int64(len(acv)),
			},
		},
	})
}

/* buildFilterActionVerification - Verifying params to method List */
func buildFilterActionVerification(QueryParam func(string) string) (m.FilterActionVerification, error) {
	f := m.FilterActionVerification{}
	acveID := QueryParam("acveID")
	if len(acveID) > 0 {
		f.AcveID = &acveID
	}
	userID := QueryParam("userID")
	if len(userID) > 0 {
		f.UserID = &userID
	}
	types := QueryParam("type")
	if len(types) > 0 {
		f.Type = &types
	}
	verification := QueryParam("verification")
	if len(verification) > 0 {
		f.Verification = &verification
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
