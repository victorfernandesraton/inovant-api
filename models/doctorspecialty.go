package models

import (
	"time"

	"github.com/gofrs/uuid"
)

//DoctorSpecialty is a representation of the table DoctorSpecialty
type DoctorSpecialty struct {
	DoctID    uuid.UUID `db:"doct_id" json:"doctID"`
	SpecID    int64     `db:"spec_id" json:"specID"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

//FilterDoctorSpecialty to get a List of DoctorSpecialty
type FilterDoctorSpecialty struct {
	DoctID *string
	SpecID *int64
	Limit  *int64
	Offset *int64
}
