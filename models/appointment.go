package models

import (
	"time"

	"github.com/gofrs/uuid"
)

//Appointment is a representation of the table Appointment
type Appointment struct {
	AppoID    uuid.UUID `db:"appo_id" json:"appoID"`
	StartAt   time.Time `db:"start_at" json:"startAt"`
	ScheID    uuid.UUID `db:"sche_id" json:"scheID"`
	PatiID    uuid.UUID `db:"pati_id" json:"patiID"`
	PatiName  *string   `db:"pati_name" json:"patiName"`
	Type      string    `db:"type" json:"type"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

//FilterAppointment to get a List of Appointment
type FilterAppointment struct {
	AppoID      *string
	StartAtGte  *time.Time
	StartAtLte  *time.Time
	ScheID      *string
	PatiID      *string
	Type        *string
	Status      *string
	InitialDate *time.Time
	FinishDate  *time.Time
	FieldOrder  *string
	TypeOrder   *string
	Limit       *int64
	Offset      *int64
}
