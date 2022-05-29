package handler

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/now"
	"github.com/labstack/echo"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// AvaliabilityHandler service to create handler
type AvaliabilityHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	check        func(m.FilterAvaliability) ([]m.Avaliability, error)
}

type avaliabilityResponse struct {
	collectionItemData
	Items []m.Avaliability `json:"items"`
	Kind  string           `json:"kind"`
}

type avaliabilityListResponse struct {
	dataResponse
	Data avaliabilityResponse `json:"data"`
}

// Check Avaliability returns an echo handler
// @Summary Avaliability.Check
// @Description Check Avaliability
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Avaliability body models.Avaliability true "Check Avaliability"
// @Success 200 {object} handler.avaliabilityListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/avaliability [get]
func (handler *AvaliabilityHandler) Check(c echo.Context) error {
	f, err := buildFilterAvaliability(c.QueryParam)
	if err != nil {
		return err
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	if doctID != nil {
		f.DoctID = doctID
	}

	ava, err := handler.check(f)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, avaliabilityListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: avaliabilityResponse{
			Kind:  "Avaliability list",
			Items: ava,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(ava)),
				TotalItems:       int64(len(ava)),
			},
		},
	})
}

/* buildFilterAvaliability - Verifying params to method List */
func buildFilterAvaliability(QueryParam func(string) string) (m.FilterAvaliability, error) {
	f := m.FilterAvaliability{}
	f.StartDate = time.Now()
	if len(QueryParam("startDate")) > 0 {
		initialDate, err := time.Parse("2006-01-02", QueryParam("startDate"))
		if err != nil {
			return f, err
		}
		f.StartDate = initialDate.Add((time.Hour * 23) + (time.Millisecond * 60000) - 1)
	}

	f.EndDate = now.EndOfMonth()
	if len(QueryParam("endDate")) > 0 {
		finalDate, err := time.Parse("2006-01-02", QueryParam("endDate"))
		if err != nil {
			return f, err
		}
		f.EndDate = finalDate.Add((time.Hour * 23) + (time.Millisecond * 60000) - 1)
	}

	if len(QueryParam("doctID")) > 0 {
		did, err := uuid.FromString(QueryParam("doctID"))
		if err != nil {
			return f, err
		}
		f.DoctID = &did
	}

	plan := QueryParam("plan")
	if len(plan) > 0 {
		f.Plan = plan
	}
	return f, nil
}
