package models

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/types"
	"gopkg.in/guregu/null.v3"
)

//Patient is a representation of the table Patient
type Patient struct {
	PatiID          uuid.UUID      `db:"pati_id" json:"patiID"`
	DoctID          uuid.UUID      `db:"doct_id" json:"doctID"`
	DoctName        string         `db:"doct_name" json:"doctName"`
	Name            string         `db:"name" json:"name"`
	Email           string         `db:"email" json:"email"`
	Info            types.JSONText `db:"info" json:"info"`
	CreatedAt       time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updatedAt"`
	LastAppointment null.Time      `db:"last_appointment" json:"lastAppointment"`
}

//FilterPatient to get a List of Patient
type FilterPatient struct {
	PatiID      *string
	DoctID      *string
	Name        *string
	Email       *string
	InitialDate *time.Time
	FinishDate  *time.Time
	Limit       *int64
	Offset      *int64
}
