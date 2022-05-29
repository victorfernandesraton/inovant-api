package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/types"
	"gopkg.in/guregu/null.v3"
)

//Schedule is a representation of the table Schedule
type Schedule struct {
	ScheID    uuid.UUID      `db:"sche_id" json:"scheID"`
	DoctID    uuid.UUID      `db:"doct_id" json:"doctID"`
	NameDoct  *string        `db:"name" json:"name"`
	RoomID    *uuid.UUID     `db:"room_id" json:"roomID"`
	Label     *string        `db:"label" json:"label"`
	StartAt   time.Time      `db:"start_at" json:"startAt"`
	EndAt     time.Time      `db:"end_at" json:"endAt"`
	Plan      string         `db:"plan" json:"plan"`
	Info      types.JSONText `db:"info" json:"info"`
	CreatedAt time.Time      `db:"created_at" json:"createdAt"`
	DeletedAt null.Time      `db:"deleted_at" json:"deletedAt"`
}

//ScheduleNotifier is a representation to notify of last fifteen minutes schedules
type ScheduleNotifier struct {
	Schedule
	UserID uuid.UUID `db:"user_id" json:"userID"`
	Token  *string   `db:"push_tokens" json:"pushToken"`
}

//FilterSchedule to get a List of Schedule
type FilterSchedule struct {
	ScheID      *string
	DoctID      *string
	RoomID      *string
	StartAt     *time.Time
	EndAt       *time.Time
	Plan        *string
	InitialDate *time.Time
	FinishDate  *time.Time
	FieldOrder  *string
	TypeOrder   *string
	Hour        *string
	Limit       *int64
	Offset      *int64
}

//Calendar is a representation of listCalendar query
type Calendar struct {
	ScheID          uuid.UUID       `db:"sche_id" json:"scheID"`
	RoomID          uuid.UUID       `db:"room_id" json:"roomID"`
	DoctID          uuid.UUID       `db:"doct_id" json:"doctID"`
	DocName         string          `db:"doc_name" json:"docName"`
	DocTreatment    *string         `db:"doc_treatment" json:"docTreatment"`
	DataAppointment time.Time       `db:"data_appointment" json:"dataAppointment"`
	StartHour       string          `db:"start_hour" json:"startHour"`
	EndHour         string          `db:"end_hour" json:"endHour"`
	Patient         patientCalendar `db:"patient" json:"patient"`
}

//FilterCalendar is a representation to filter listCalendar query
type FilterCalendar struct {
	DoctID  *string
	PatiID  *string
	StartAt *time.Time
	EndAt   *time.Time
	Limit   *int64
	Offset  *int64
}

//CalendarPatient is a representation of Patient on Calendar
type CalendarPatient struct {
	HourAppointment *string `db:"hour_appointment" json:"hourAppointment"`
	PatientName     *string `db:"patient_name" json:"patientName"`
	Status          *string `db:"status" json:"status"`
	Type            *string `db:"type" json:"type"`
	AppoID          *string `db:"appo_id" json:"appoID"`
	PatiID          *string `db:"pati_id" json:"patiID"`
	StartAt         *string `db:"start_at" json:"startAt"`
}

//Outdoor is a representation of the table Outdoor
type Outdoor struct {
	DoctID    uuid.UUID      `db:"doct_id" json:"doctID"`
	NameDoct  string         `db:"name" json:"name"`
	Avatar    *string        `db:"avatar" json:"avatar"`
	Treatment *string        `db:"treatment" json:"treatment"`
	RoomID    uuid.UUID      `db:"room_id" json:"roomID"`
	LabelSala string         `db:"label" json:"label"`
	Specialty types.JSONText `db:"specialties" json:"specialties"`
}

//UsableTransition is a representation of the query UsableTransition
type UsableTransition struct {
	ScheID         uuid.UUID `db:"sche_id" json:"ScheID"`
	StartAt        time.Time `db:"start_at" json:"startAt"`
	EndAt          time.Time `db:"end_at" json:"endAt"`
	Usable         bool      `db:"usable" json:"usable"`
	TransitionTime string    `db:"transition_time" json:"transitionTime"`
	IsAbleToExtend bool      `db:"is_able_to_extend" json:"isAbleToExtend"`
}

type patientCalendar []CalendarPatient

// Value implements the driver Valuer interface.
func (i patientCalendar) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *patientCalendar) Scan(src interface{}) error {
	var source []byte
	// let's support string and []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("Incompatible type for Schedule")
	}
	return json.Unmarshal(source, i)
}
