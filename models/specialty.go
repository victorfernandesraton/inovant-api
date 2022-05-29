package models

import "github.com/jmoiron/sqlx/types"

//Specialty is a representation of the table Specialty
type Specialty struct {
	SpecID      int64          `db:"spec_id" json:"specID"`
	Name        string         `db:"name" json:"name"`
	Description string         `db:"description" json:"description"`
	Info        types.JSONText `db:"info" json:"info"`
}

//FilterSpecialty to get a List of Specialty
type FilterSpecialty struct {
	Name        *string
	Description *string
	Limit       *int64
	Offset      *int64
}
