package handler

import (
	"net/http"
	"time"

	"github.com/jinzhu/now"
	"github.com/labstack/echo"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// DashboardHandler service to create handler
type DashboardHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	view         func(m.FilterDashboard) ([]m.Dashboard, error)
}

type dashboardResponse struct {
	collectionItemData
	Items []m.Dashboard `json:"items"`
	Kind  string        `json:"kind"`
}

type dashboardViewResponse struct {
	dataResponse
	Data dashboardResponse `json:"data"`
}

// View Dashboard returns an echo handler
// @Summary Dashboard.Check
// @Description Check Dashboard
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Dashboard body models.Dashboard true "Check Dashboard"
// @Success 200 {object} handler.dashboardViewResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/Dashboard [get]
func (handler *DashboardHandler) View(c echo.Context) error {
	f, err := buildFilterDashboard(c.QueryParam)
	if err != nil {
		return err
	}

	doctID, err := doctIDOrNil(c, handler.claimsCtxKey, handler.rolesCtxKey)
	if err != nil {
		return err
	}
	f.DoctID = *doctID

	das, err := handler.view(f)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dashboardViewResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: dashboardResponse{
			Kind:  "Dashboard view",
			Items: das,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(das)),
				TotalItems:       int64(len(das)),
			},
		},
	})
}

/* buildFilterDashboard - Verifying params to method List */
func buildFilterDashboard(QueryParam func(string) string) (m.FilterDashboard, error) {
	f := m.FilterDashboard{}
	f.StartDate = now.BeginningOfMonth()
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

	return f, nil
}
