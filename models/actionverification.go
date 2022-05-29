package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v3"
)

//ActionVerification is a representation of the table ActionVerification
type ActionVerification struct {
	AcveID       uuid.UUID `db:"acve_id" json:"acveID"`
	UserID       uuid.UUID `db:"user_id" json:"userID"`
	Type         string    `db:"type" json:"type"`
	Verification string    `db:"verification" json:"verification"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	DeletedAt    null.Time `db:"deleted_at" json:"deletedAt"`
}

//FilterActionVerification to get a List of ActionVerification
type FilterActionVerification struct {
	AcveID       *string
	UserID       *string
	Type         *string
	Verification *string
	InitialDate  *time.Time
	FinishDate   *time.Time
	Limit        *int64
	Offset       *int64
}
