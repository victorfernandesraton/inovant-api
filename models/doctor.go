package models

import (
	"github.com/lib/pq"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/types"
)

//Doctor is a representation of the table Doctor
type Doctor struct {
	User
	DoctID      uuid.UUID       `db:"doct_id" json:"doctID"`
	Name        string          `db:"name" json:"name"`
	Info        types.JSONText  `db:"info" json:"info"`
	UpdateAt    time.Time       `db:"update_at" json:"updateAt"`
	Specialties *pq.StringArray `db:"specialties" json:"specialties"`
}

//FilterDoctor to get a List of Doctor
type FilterDoctor struct {
	DoctID *string
	UserID *string
	Name   *string
	Limit  *int64
	Offset *int64
}
