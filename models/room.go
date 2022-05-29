package models

import (
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/types"
	"gopkg.in/guregu/null.v3"
)

//Room is a representation of the table Room
type Room struct {
	RoomID     uuid.UUID      `db:"room_id" json:"roomID"`
	Label      string         `db:"label" json:"label"`
	InactiveAt null.Time      `db:"inactive_at" json:"inactiveAt"`
	Info       types.JSONText `db:"info" json:"info"`
}

//FilterRoom to get a List of Room
type FilterRoom struct {
	RoomID *string
	Label  *string
	Limit  *int64
	Offset *int64
}
