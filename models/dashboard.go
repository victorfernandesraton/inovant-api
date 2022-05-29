package models

import (
	"time"

	"github.com/gofrs/uuid"
)

//Dashboard model
type Dashboard struct {
	DoctID    uuid.UUID `db:"doct_id" json:"doctID"`
	DateStart time.Time `db:"start_at" json:"startAt"`
	DateEnd   time.Time `db:"end_at" json:"endAt"`
	Type      string    `db:"plan" json:"plan"`
}

//FilterDashboard Model
type FilterDashboard struct {
	DoctID    uuid.UUID `db:"doct_id" json:"doctID"`
	StartDate time.Time `db:"start_date" json:"startDate"`
	EndDate   time.Time `db:"end_date" json:"endDate"`
}
