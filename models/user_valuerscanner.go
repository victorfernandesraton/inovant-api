package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Value implements the driver Valuer interface.
func (i UserRole) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *UserRole) Scan(src interface{}) error {
	var source []byte
	// let's support string and []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("Incompatible type for Role")
	}
	return json.Unmarshal(source, i)
}
