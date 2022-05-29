package models

import (
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v3"
	"time"
)

//UserRole if the user type for roles
type UserRole []string

//User is a representation of the table user
type User struct {
	UserID     uuid.UUID `db:"user_id" json:"userID"`
	Email      string    `db:"email" json:"email"`
	Password   []byte    `db:"password" json:"-"`
	Roles      UserRole  `db:"roles" json:"roles"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	InactiveAt null.Time `db:"inactive_at" json:"inactiveAt"`
	Token      *string   `db:"push_tokens" json:"pushToken"`
}

//UserWithDoctor is a representation of the table UserWithDoctor
type UserWithDoctor struct {
	User
	DoctID   *uuid.UUID `db:"doct_id" json:"doctID"`
	DoctName *string    `db:"doct_name" json:"doctName"`
}
