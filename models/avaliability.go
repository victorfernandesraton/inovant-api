package models

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/types"
)

//Avaliability Model
type Avaliability struct {
	DoctID   uuid.UUID      `db:"doct_id" json:"doctID"`
	DoctName string         `db:"doct_name" json:"doctName"`
	Date     time.Time      `db:"date" json:"date"`
	Slots    types.JSONText `db:"slots" json:"slots"`
}

//FilterAvaliability Model
type FilterAvaliability struct {
	DoctID    *uuid.UUID `db:"doct_id" json:"doctID"`
	StartDate time.Time  `db:"start_date" json:"startDate"`
	EndDate   time.Time  `db:"end_date" json:"endDate"`
	Plan      string     `db:"plan" json:"plan"`
}
