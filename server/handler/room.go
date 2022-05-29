package handler

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// RoomHandler service to create handler
type RoomHandler struct {
	rolesCtxKey  string
	claimsCtxKey string
	create       func(*m.Room) (*m.Room, error)
	update       func(*m.Room) (*m.Room, error)
	delete       func(roomID uuid.UUID) (*m.Room, error)
	list         func(m.FilterRoom) ([]m.Room, error)
	get          func(roomID uuid.UUID) (*m.Room, error)
}

type roomResponse struct {
	Item *m.Room `json:"item"`
	Kind string  `json:"kind"`
}

type roomGetResponse struct {
	dataResponse
	Data roomResponse `json:"data"`
}

type roomDelResponse struct {
	Item *m.Room `json:"item"`
	Kind string  `json:"kind"`
}

type roomDeleteResponse struct {
	dataResponse
	Data roomDelResponse `json:"data"`
}

type roomsResponse struct {
	collectionItemData
	Items []m.Room `json:"items"`
	Kind  string   `json:"kind"`
}

type roomsListResponse struct {
	dataResponse
	Data roomsResponse `json:"data"`
}

// Create Room returns an echo handler
// @Summary Room.Create
// @Description Create Room
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param Room body models.Room true "Create new Room"
// @Success 200 {object} handler.roomGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 401 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/rooms [post]
func (handler *RoomHandler) Create(c echo.Context) error {
	req := m.Room{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	rom, err := handler.create(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to create new Room")
	}
	return c.JSON(http.StatusOK, roomGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: roomResponse{
			Kind: "Room",
			Item: rom,
		},
	})
}

// Update returns an echo handler
// @Summary Room.Update
// @Description Update Room
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param roomID path string true "Room ID" Format(string)
// @Param Room body models.Room true "Room Update Body"
// @Success 200 {object} handler.roomGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/rooms/{roomID} [put]
func (handler *RoomHandler) Update(c echo.Context) error {
	req := m.Room{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	req.RoomID, err = uuid.FromString(c.Param("roomID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	rom, err := handler.update(&req)
	if err != nil {
		return errors.Wrap(err, "Fail to update Room")
	}
	return c.JSON(http.StatusOK, roomGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: roomResponse{
			Kind: "Room update",
			Item: rom,
		},
	})
}

// Delete Room returns an echo handler
// @Summary Room.Delete
// @Description   delete Room
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param roomID path string true "Delete Room" Format(string)
// @Success 200 {object} handler.roomDeleteResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/rooms/{roomID} [del]
func (handler *RoomHandler) Delete(c echo.Context) error {
	roomID, err := uuid.FromString(c.Param("roomID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}
	rom, err := handler.delete(roomID)
	if err != nil {
		return errors.Wrap(err, "Fail to delete Room")
	}
	return c.JSON(http.StatusOK, roomDeleteResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: roomDelResponse{
			Kind: "Room deleted",
			Item: rom,
		},
	})
}

// Get returns an echo handler
// @Summary Room.Get
// @Description Get a Room
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param roomID query string true "Filter Rooms by type [roomID]"
// @Success 200 {object} handler.roomGetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/rooms/{roomID} [get]
func (handler *RoomHandler) Get(c echo.Context) error {
	roomID, err := uuid.FromString(c.Param("roomID"))
	if err != nil {
		return errors.Wrap(err, "Error uuid format")
	}

	rom, err := handler.get(roomID)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Rooms")
	}
	return c.JSON(http.StatusOK, roomGetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: roomResponse{
			Kind: "Room get",
			Item: rom,
		},
	})
}

// List returns an echo handler
// @Summary Room.List
// @Description Get Room list
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param roomID query string false "Filter Rooms by type [roomID]"
// @Param label query string false "Filter Rooms by type [label]"
// @Success 200 {object} handler.roomsListResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /api/rooms [get]
func (handler *RoomHandler) List(c echo.Context) error {
	f, err := buildFilterRoom(c.QueryParam)
	if err != nil {
		return errors.Wrap(err, "Failed to parse filter queries")
	}

	rom, err := handler.list(f)
	if err != nil {
		return errors.Wrap(err, "Fail to list of Rooms")
	}
	return c.JSON(http.StatusOK, roomsListResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: roomsResponse{
			Kind:  "Room list",
			Items: rom,
			collectionItemData: collectionItemData{
				CurrentItemCount: int64(len(rom)),
				TotalItems:       int64(len(rom)),
			},
		},
	})
}

/* buildFilterRoom - Verifying params to method List */
func buildFilterRoom(QueryParam func(string) string) (m.FilterRoom, error) {
	f := m.FilterRoom{}
	roomID := QueryParam("roomID")
	if len(roomID) > 0 {
		f.RoomID = &roomID
	}
	label := QueryParam("label")
	if len(label) > 0 {
		f.Label = &label
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
