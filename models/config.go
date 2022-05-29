package models

import (
	types "github.com/jmoiron/sqlx/types"
)

//Config is a representation of the table Config
type Config struct {
	Key   string         `db:"key" json:"key"`
	Value types.JSONText `db:"value" json:"value"`
}

//FilterConfig to get a List of Config
type FilterConfig struct {
	Key    *string
	Limit  *int64
	Offset *int64
}
