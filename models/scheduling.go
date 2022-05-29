package models

import (
	"encoding/json"
	"errors"

	"database/sql/driver"
	"github.com/gofrs/uuid"
)

//RoomSched using to scheduling algorithm
type RoomSched struct {
	RoomID string
	ScheID []uuid.UUID
}

//Slot using to scheduling algorithm
type Slot struct {
	StartAt      string    `json:"start"`
	EndAt        string    `json:"end"`
	ScheID       uuid.UUID `json:"scheID"`
	DoctID       uuid.UUID `json:"doctID"`
	NeedBathroom bool      `json:"needBathroom"`
}

//SchedGroup using to scheduling algorithm
type SchedGroup struct {
	StartAt   string      `json:"start"`
	EndAt     string      `json:"end"`
	Schedules []uuid.UUID `json:"scheID"`
}

//Slots using to scheduling algorithm
type Slots []Slot

//SchedsGroup using to scheduling algorithm
type SchedsGroup []SchedGroup

//Scheduling is a representation query Scheduling
type Scheduling struct {
	SlotSched Slot `db:"slot_sched" json:"slotSched"`
}

//SchedulingSlots is a representation query Scheduling
type SchedulingSlots struct {
	RoomID      uuid.UUID   `db:"room_id" json:"roomID"`
	Slots       SchedsGroup `db:"slots" json:"slots"`
	HasBathroom bool        `db:"has_bathroom" json:"hasBathroom"`
}

// Value implements the driver Valuer interface.
func (i Slot) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *Slot) Scan(src interface{}) error {
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

// Value implements the driver Valuer interface.
func (i Slots) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *Slots) Scan(src interface{}) error {
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

// Value implements the driver Valuer interface.
func (i SchedGroup) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *SchedGroup) Scan(src interface{}) error {
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

// Value implements the driver Valuer interface.
func (i SchedsGroup) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *SchedsGroup) Scan(src interface{}) error {
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
