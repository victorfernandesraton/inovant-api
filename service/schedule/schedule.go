package schedule

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type generalError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Errors  []detailError `json:"errors,omitempty"`
}

type detailError struct {
	Domain       string  `json:"domain"`
	Reason       string  `json:"reason"`
	Message      string  `json:"message"`
	Location     *string `json:"location,omitempty"`
	LocationType *string `json:"locationType,omitempty"`
	ExtendedHelp *string `json:"extendedHelp,omitempty"`
	SendReport   *string `json:"sendReport,omitempty"`
}

type errNotFound struct {
	err error
}

//Creator service to create new Schedule
type Creator struct {
	DB     service.DB
	Logger *log.Logger
}

//Run create new Schedule
func (c *Creator) Run(sch *m.Schedule) (*m.Schedule, error) {
	scheID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Schedule uuid")
	}
	sch.ScheID = scheID
	u, err := createSchedule(c.DB, sch)
	return u, err
}

//Lister service to return Schedule
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Schedule by Filter
func (l *Lister) Run(doctID *uuid.UUID, f m.FilterSchedule) ([]m.Schedule, error) {
	u, err := listSchedule(l.DB, doctID, f)
	return u, err
}

//Getter service to return Schedule
type Getter struct {
	DB *sqlx.DB
}

//Run return a Schedule by sche_id
func (g *Getter) Run(doctID *uuid.UUID, scheID uuid.UUID) (*m.Schedule, error) {
	u, err := getSchedule(g.DB, doctID, scheID)
	return u, err
}

//UpdateSchedule service to return Schedule
type UpdateSchedule struct {
	DB *sqlx.DB
}

//Run return a Schedule by sche_id
func (g *UpdateSchedule) Run(sch *m.Schedule) (*m.Schedule, error) {
	u, err := updateSchedule(g.DB, sch, true)
	return u, err
}

//Updater service to update Schedule
type Updater struct {
	DB     *sqlx.DB
	Logger *log.Logger
	Create Creator
}

//Run update Schedule data
func (g *Updater) Run(sch *m.Schedule) (*m.Schedule, error) {
	s := m.Schedule{
		EndAt:   sch.EndAt,
		StartAt: sch.StartAt,
		DoctID:  sch.DoctID,
		Plan:    sch.Plan,
		Info:    sch.Info,
	}
	tx, err := g.DB.Beginx()
	_, err = updateDeleteAtSchedule(tx, sch.ScheID)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Error Updating Schedule:", err)
		}
		tx.Rollback()
		return nil, err
	}
	scheC := Creator{DB: tx}
	cre, err := scheC.Run(&s)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Error Creating Schedule:", err)
		}
		tx.Rollback()
		return nil, err
	}
	sch.StartAt = cre.StartAt
	sch.EndAt = cre.EndAt
	sch.RoomID = cre.RoomID
	sch.DeletedAt.Valid = false
	upd, err := updateSchedule(tx, sch, false)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Error Updating Schedule:", err)
		}
		tx.Rollback()
		return nil, err
	}
	scheD := Deleter{DB: tx}
	_, err = scheD.Run(cre.ScheID, &cre.DoctID)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Error Deleting Schedule:", err)
		}
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return upd, err
}

//Deleter service to soft delete Schedule
type Deleter struct {
	DB service.DB
}

//Run soft delete Schedule by sche_id
func (d *Deleter) Run(scheID uuid.UUID, doctID *uuid.UUID) (*m.Schedule, error) {
	u, err := deleteSchedule(d.DB, scheID, doctID)
	return u, err
}

//Calendar service to soft list calendar view
type Calendar struct {
	DB *sqlx.DB
}

//Run service Calendar to list query calendar
func (c *Calendar) Run(doctID *uuid.UUID, fCalendar m.FilterCalendar) ([]m.Calendar, error) {
	u, err := listCalendar(c.DB, doctID, fCalendar)
	return u, err
}

//UpdateDeleter service to soft list calendar view
type UpdateDeleter struct {
	DB *sqlx.DB
}

//Run service Calendar to list query calendar
func (up *UpdateDeleter) Run(scheID uuid.UUID) (*m.Schedule, error) {
	u, err := updateDeleteAtSchedule(up.DB, scheID)
	return u, err
}

//Outdoor service to soft list Outdoor view
type Outdoor struct {
	DB *sqlx.DB
}

//Run service Outdoor to list query Outdoor
func (c *Outdoor) Run(roomID uuid.UUID) (*m.Outdoor, error) {
	u, err := outdoor(c.DB, roomID)
	return u, err
}

/* Create a new Schedule to database */
func createSchedule(db service.DB, sch *m.Schedule) (*m.Schedule, error) {

	err := validationsInsertSchedule(db, sch)
	if err != nil {
		return nil, err
	}

	needBathroom, bathroomTreatment, err := needBathroom(db, sch.DoctID)
	if err != nil {
		return nil, err
	}

	args := []interface{}{sch.ScheID, sch.DoctID, sch.StartAt, sch.EndAt, sch.Plan, sch.Info}

	filterRooms := filterRooms(needBathroom, bathroomTreatment)

	query := `
			WITH config_days AS (
				SELECT value
				FROM config
				WHERE KEY = 'schedule-hour_config_flex'
			),
			config_timezone AS (
				SELECT value->>'timezone' AS timezone
				FROM config
				WHERE KEY = 'timezone-local'
			),
			dates_to_local AS (
				SELECT (($3::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_time,
				(($4::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_time
				FROM config_timezone cti
			),
			config_result_int AS (
				SELECT row_number() OVER(ORDER BY a."key") AS id,
				a."key" AS "day",
				jsonb_array_elements(a.value) AS slot
				FROM config_days, jsonb_each(config_days.value) a
			),
			config_building_timezone AS (
				SELECT id, "day",
				(( ((slot->>'start')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE ct.timezone)::TIME AS start_time,
				(( ((slot->>'end')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE ct.timezone)::TIME AS end_time
				FROM config_result_int, config_timezone ct
			),
			config_result AS (
				SELECT id, "day", jsonb_build_object('start', substring((start_time::TIME)::TEXT, 0, 6), 'end', substring((end_time::TIME)::TEXT, 0, 6)) AS slot
				FROM config_building_timezone
			),
			slots_by_day AS (
				SELECT id, "day", slot,
				EXTRACT(EPOCH FROM (slot->>'start')::TIME)::INT AS start_slot,
				EXTRACT(EPOCH FROM (slot->>'end')::TIME)::INT AS end_slot
				FROM config_result
			),
			config_transition AS (
				SELECT value
				FROM config
				WHERE KEY = 'schedule-transition_time'
			),
			schedule_local AS (
				SELECT sche_id, doct_id, room_id,
					(( ((start_at)) AT TIME ZONE 'UTC') AT TIME ZONE ct.timezone) AS start_at,
					(( ((end_at)) AT TIME ZONE 'UTC') AT TIME ZONE ct.timezone) AS end_at,
					plan, info, created_at, deleted_at
				FROM schedule, config_timezone ct
			),
			scheduled AS (
				SELECT room_id, start_at::date as id, start_at,
				CASE
					WHEN doct_id = $2 THEN end_at
					ELSE (end_at + (value->>'transition_time' ||' minutes')::INTERVAL)::TIMESTAMP
				END AS end_at,
				slot,
				ROW_NUMBER () OVER (
					PARTITION BY room_id, start_at::date, slot
					ORDER BY start_at ASC
				),
				doct_id
				FROM config_transition ct, schedule_local s
				JOIN slots_by_day sbd ON EXTRACT(EPOCH FROM (s.start_at)::TIME) >= sbd.start_slot
					AND EXTRACT(EPOCH FROM (s.end_at)::TIME) <= sbd.end_slot
					AND TRIM(TO_CHAR(s.start_at, 'day'))::TEXT = "day"
				WHERE deleted_at IS NULL
				ORDER BY room_id, id
			),
			counted AS (
				SELECT room_id, id, count(*), slot
				FROM scheduled
				GROUP BY room_id, id, slot
				ORDER BY room_id, id
			),
			prev AS (
				SELECT counted.room_id, counted.id, NULL::TIMESTAMP AS start_at,
				((scheduled.start_at::DATE)::TEXT || ' ' || (counted.slot->>'start')::TEXT || ':00')::TIMESTAMP AS end_at, counted.slot, 0 AS ROW_NUMBER
				FROM counted
				JOIN scheduled ON scheduled.room_id = counted.room_id AND scheduled.id = counted.id
			),
			nextt AS (
				SELECT counted.room_id, counted.id,
				((scheduled.start_at::DATE)::TEXT || ' ' || (counted.slot->>'end')::TEXT || ':00')::TIMESTAMP, NULL::TIMESTAMP, counted.slot, count+1 rn
				FROM counted
				JOIN scheduled ON scheduled.room_id = counted.room_id AND scheduled.id = counted.id
			),
			joined AS (
				SELECT * FROM prev
				UNION
				SELECT room_id, id, start_at, end_at, slot, ROW_NUMBER FROM scheduled
				UNION
				SELECT * FROM nextt
			),
			order_joined as (
				SELECT COALESCE(end_at, start_at) AS ord, *
				FROM joined
			),
			ordered as (
				SELECT * FROM order_joined
				ORDER BY room_id, id, ord, ROW_NUMBER
			),
			create_sched_date AS (
				SELECT *, COALESCE(start_at, end_at) AS sched_date
				FROM ordered
			),
			order_sched_date AS (
				SELECT room_id, id, start_at, end_at, slot, ROW_NUMBER, sched_date
				FROM create_sched_date
				ORDER BY room_id, id, sched_date, ROW_NUMBER
			),
			slots_free AS (
				SELECT room_id, id, ROW_NUMBER, "count",
				order_sched_date.start_at, end_at, counted.slot,
				LAG(end_at::TIME) OVER (PARTITION BY room_id, sched_date, ROW_NUMBER) AS prev_end_at,
				CASE
					WHEN ROW_NUMBER < count THEN jsonb_build_object('start', substring( (LAG(end_at::TIME) OVER (PARTITION BY room_id) )::text, 1, 5) , 'end', substring( (order_sched_date.start_at::TIME)::TEXT, 1, 5))
					WHEN ROW_NUMBER = count THEN jsonb_build_object('start', substring( (LAG(end_at::TIME) OVER (PARTITION BY room_id) )::text, 1, 5) , 'end', substring( (order_sched_date.start_at::TIME)::TEXT, 1, 5))
					WHEN ROW_NUMBER > count THEN jsonb_build_object('start', substring( (LAG(end_at::time) OVER (PARTITION BY room_id) )::text, 1, 5) , 'end', counted.slot->>'end')
				END slot_free
				FROM order_sched_date
				JOIN counted USING (room_id, id, slot)
			),
			slots_free_by_room AS (
				SELECT room_id, start_at::DATE AS date_of_month, slot_free
				FROM slots_free
				ORDER BY start_at, room_id
			),
			generated_dates_by_filter AS (
				SELECT roo.room_id, generate_series(dtl.start_time, dtl.end_time, '1 day'::INTERVAL) AS days
				FROM room roo, dates_to_local dtl
				WHERE roo.inactive_at IS NULL
				` + filterRooms + `
			),
			slots_with_days AS (
				SELECT room_id, days::DATE AS days, slot
				FROM generated_dates_by_filter ft
				JOIN config_result cr ON cr."day" = TRIM(TO_CHAR(ft.days, 'day'))::TEXT
			),
			slots_not_full_by_day AS (
				SELECT swd.room_id AS room_id, swd.days AS days,
				COALESCE(sfbr.slot_free, swd.slot) AS slot
				FROM slots_with_days swd
				LEFT JOIN slots_free_by_room sfbr ON sfbr.room_id = swd.room_id AND sfbr.date_of_month = swd.days AND
				( EXTRACT(EPOCH FROM (sfbr.slot_free->>'start')::TIME)::INT >= EXTRACT(EPOCH FROM (swd.slot->>'start')::TIME)::INT
				AND
				EXTRACT(EPOCH FROM (sfbr.slot_free->>'end')::TIME)::INT <= EXTRACT(EPOCH FROM (swd.slot->>'end')::TIME)::INT )
			),
			closest_rooms AS (
				SELECT ABS(EXTRACT(epoch FROM start_at - start_time)) as closest_schedule, room_id
				FROM schedule_local s
				JOIN	dates_to_local d on s.start_at::DATE = d.start_time::DATE
				AND doct_id = $2
			),
			ordered_rooms AS (
				SELECT * FROM closest_rooms
				UNION SELECT 100000000000, room_id FROM room r where not exists (SELECT room_id FROM closest_rooms c where c.room_id = r.room_id)
			),
			slots_not_full_by_day_ordered AS (
				SELECT *, (SELECT MAX(od.closest_schedule) FROM ordered_rooms od WHERE s.room_id = od.room_id) as ooo
				FROM slots_not_full_by_day s
				ORDER BY (
					SELECT MIN(od.closest_schedule) FROM ordered_rooms od WHERE s.room_id = od.room_id
				) ASC
			),
			invalid_slots AS (
				SELECT sd.*
				FROM scheduled s
				JOIN slots_not_full_by_day sd ON s.id = sd.days AND s.room_id = sd.room_id AND s.start_at::TIME = (sd.slot->>'end')::TIME
				JOIN  dates_to_local dtl ON 1=1
				JOIN  config_transition ct ON 1=1
				WHERE doct_id <> $2
				AND (dtl.end_time::TIME + (ct.value->>'transition_time' ||' minutes')::INTERVAL) > (sd.slot->>'end')::TIME
			),
			valid_slots_with_transition_time AS (
				SELECT *
				FROM slots_not_full_by_day_ordered fl
				WHERE NOT EXISTS(SELECT iv.room_id, * FROM invalid_slots iv WHERE iv.room_id = fl.room_id AND iv.days = fl.days AND iv.slot = fl.slot)
			),
			room_for_insert_flex AS (
				SELECT room_id, days, slot
				FROM valid_slots_with_transition_time, dates_to_local dtl, config_transition ct
				WHERE days = (dtl.start_time)::DATE
				AND EXTRACT(EPOCH FROM (slot->>'start')::TIME) <= EXTRACT(EPOCH FROM (dtl.start_time)::TIME)
				AND EXTRACT(EPOCH FROM (slot->>'end')::TIME) >= EXTRACT(EPOCH FROM (dtl.end_time)::TIME)
				AND EXTRACT(EPOCH FROM (slot->>'end')::TIME)::INT - EXTRACT(EPOCH FROM (slot->>'start')::TIME)::INT > 0
				LIMIT 1
			)
			INSERT INTO schedule
			SELECT $1, $2, room_id, $3, $4, $5, $6
			FROM room_for_insert_flex
			join room using (room_id)
			RETURNING *
		`
	err = db.Get(sch, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error inserting Schedule in database")
		}
		return nil, errNotFound{err: err}
	}
	return sch, nil
}

/* Return a list of Schedule by filters */
func listSchedule(db service.DB, doctID *uuid.UUID, f m.FilterSchedule) ([]m.Schedule, error) {
	sch := []m.Schedule{}
	query := psql.Select("sche_id", "doct_id", "room_id", "start_at", "end_at", "plan", "info", "created_at", "deleted_at").
		From("schedule").
		Where("deleted_at IS NULL")
	if doctID != nil {
		query = query.Where(`doct_id = ?`, doctID)
	}
	if f.ScheID != nil {
		query = query.Where(`sche_id = ?`, f.ScheID)
	}
	if f.DoctID != nil {
		query = query.Where(`doct_id = ?`, f.DoctID)
	}
	if f.RoomID != nil {
		query = query.Where(`room_id = ?`, f.RoomID)
	}
	if f.StartAt != nil {
		query = query.Where(sq.GtOrEq{"start_at": f.StartAt})
	}
	if f.EndAt != nil {
		query = query.Where(sq.LtOrEq{"end_at": f.EndAt})
	}
	if f.Plan != nil {
		query = query.Where(`plan ILIKE ?`, `%`+*f.Plan+`%`)
	}
	if f.Hour != nil {
		query = query.Where(`start_at >= NOW()`)
	}
	if f.InitialDate != nil {
		query = query.Where(sq.GtOrEq{"created_at": f.InitialDate})
	}
	if f.FinishDate != nil {
		query = query.Where(sq.LtOrEq{"created_at": f.FinishDate})
	}
	orderQuery, err := buildOrderBy("schedule.", f.FieldOrder, f.TypeOrder)
	if err != nil {
		return nil, err
	}
	if len(orderQuery) > 0 {
		query = query.OrderBy(orderQuery)
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list Schedules sql")
	}
	err = db.Select(&sch, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list Schedules sql")
		}
		return nil, nil
	}
	return sch, nil
}

/* Return a Schedule by sche_id */
func getSchedule(db service.DB, doctID *uuid.UUID, scheID uuid.UUID) (*m.Schedule, error) {
	sch := m.Schedule{}
	query := psql.Select("sche_id", "doct_id", "name", "room_id", "label", "start_at", "end_at", "plan", "schedule.info", "schedule.created_at", "deleted_at").
		From("schedule").
		LeftJoin("room USING (room_id)").
		LeftJoin("doctor USING (doct_id)").
		Where(sq.Eq{"sche_id": scheID})

	if doctID != nil {
		query = query.Where(`doct_id = ?`, doctID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Schedule sql")
	}
	err = db.Get(&sch, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Schedule sql")
		}
		return nil, nil
	}
	return &sch, nil
}

/* Update Schedule to database by sche_id */
func updateSchedule(db service.DB, sch *m.Schedule, validation bool) (*m.Schedule, error) {

	if validation {
		err := validationsUpdateSchedule(db, sch)
		if err != nil {
			return nil, err

		}
	}

	query := psql.Update("schedule").
		Set("doct_id", sch.DoctID).
		Set("room_id", sch.RoomID).
		Set("start_at", sch.StartAt).
		Set("end_at", sch.EndAt).
		Set("plan", sch.Plan).
		Set("info", sch.Info).
		Set("deleted_at", sch.DeletedAt).
		Suffix("RETURNING *").
		Where(sq.Eq{"sche_id": sch.ScheID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Schedule update sql")
	}

	err = db.Get(sch, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Schedule update sql")
	}

	return sch, nil
}

/* Delete Schedule to database by sche_id */
func deleteSchedule(db service.DB, scheID uuid.UUID, doctID *uuid.UUID) (*m.Schedule, error) {
	sch := m.Schedule{}
	query := psql.Delete("schedule").
		Suffix("RETURNING *").
		Where(sq.Eq{"sche_id": scheID})
	if doctID != nil {
		query = query.Where(`doct_id = ?`, doctID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &sch, errors.Wrap(err, "Error generating delete Schedule sql")
	}
	err = db.Get(&sch, qSQL, args...)
	if err != nil {
		return &sch, errors.Wrap(err, "Error delete Schedule sql")
	}
	return &sch, nil
}

/* UpdateDeleteAtSchedule Schedule to database by sche_id */
func updateDeleteAtSchedule(db service.DB, scheID uuid.UUID) (*m.Schedule, error) {
	sch := m.Schedule{}
	query := psql.Update("schedule").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"sche_id": scheID}).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &sch, errors.Wrap(err, "Error generating delete Schedule sql")
	}
	err = db.Get(&sch, qSQL, args...)
	if err != nil {
		return &sch, errors.Wrap(err, "Error delete Schedule sql")
	}
	return &sch, nil
}

func (e errNotFound) Error() string {
	return "No schedule available"
}

/* List Calendar listing schedules to put at calendar */
func listCalendar(db service.DB, doctID *uuid.UUID, fCalendar m.FilterCalendar) ([]m.Calendar, error) {
	sch := []m.Calendar{}
	args := []interface{}{}
	filterDoctor := ""
	if doctID != nil {
		args = append(args, doctID)
		filterDoctor = `AND sche.doct_id = $` + strconv.Itoa(len(args))
	}
	if fCalendar.DoctID != nil {
		args = append(args, fCalendar.DoctID)
		filterDoctor = `AND sche.doct_id = $` + strconv.Itoa(len(args))
	}
	filterPatients := ""
	if fCalendar.PatiID != nil {
		filterPatients = `AND pat.pati_id = $` + strconv.Itoa(len(args)+1)
		args = append(args, fCalendar.PatiID)
	}
	filterDate := ""
	if fCalendar.StartAt != nil && fCalendar.EndAt != nil {
		filterDate = `AND sche.start_at::DATE >= $` + strconv.Itoa(len(args)+1) + ` AND sche.end_at::DATE <= $` + strconv.Itoa(len(args)+2)
		args = append(args, fCalendar.StartAt, fCalendar.EndAt)
	}

	query :=
		`WITH inter_calendar AS (
			SELECT sche.sche_id, sche.room_id, sche.doct_id, doc.name AS doc_name, doc.info->>'treatment' AS doc_treatment ,sche.start_at::DATE AS data_appointment,
				sche.start_at AS start_hour, sche.end_at AS end_hour,
				jsonb_build_object('patientName', pat.name, 'hourAppointment', to_char(app.start_at::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), 'status', app.status, 'appoID', app.appo_id, 'patiID', pat.pati_id, 'type', app."type", 'startAt', app.start_at) AS arr_patient
			FROM schedule sche
			LEFT JOIN doctor doc USING (doct_id)
			LEFT JOIN appointment app USING(sche_id)
			LEFT JOIN patient pat USING(pati_id)
			WHERE deleted_at IS NULL
			` + filterPatients + `
			` + filterDoctor + `
			` + filterDate + `
			ORDER BY sche.start_at::DATE, sche.start_at::TIME, app.start_at::TIME
		),
		calendar AS (
			SELECT sche_id, room_id, doct_id, doc_name, doc_treatment, data_appointment, start_hour, end_hour,jsonb_agg(arr_patient) AS patient
			FROM inter_calendar
			GROUP BY sche_id, room_id, doct_id, doc_name, doc_treatment, data_appointment , start_hour, end_hour
			ORDER BY start_hour
		),
		scheduled AS (
			SELECT sche_id, room_id, doct_id, doc_name, doc_treatment, data_appointment,
				(to_char(start_hour::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')) AS start_hour,
				(to_char(end_hour::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')) AS end_hour,
				patient
			FROM calendar c
			ORDER BY data_appointment, start_hour::TIME
		),
		config_transition AS (
			SELECT value
			FROM config
			WHERE KEY = 'schedule-transition_time'
		),
		range_calendar AS (
			SELECT sche_id, room_id, doct_id, doc_name, doc_treatment, data_appointment,
				(to_char(start_hour::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')) AS start_hour,
				(to_char(end_hour::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')) AS end_hour,
				LEAD (data_appointment) OVER (PARTITION BY room_id, doct_id, data_appointment) AS adata_appointment,
				LEAD (doct_id) OVER (PARTITION BY room_id, doct_id, data_appointment) AS adoct_id,
				LEAD (room_id) OVER (PARTITION BY room_id, doct_id, data_appointment) AS aroom_id,
				LEAD (start_hour) OVER (PARTITION BY room_id, doct_id, data_appointment) AS next_start_hour,
				patient,
				value->>'transition_time' AS transition_time,
				rank() OVER w
			FROM scheduled, config_transition
			WINDOW w as (PARTITION BY room_id, doct_id, data_appointment ORDER BY data_appointment, start_hour::TIMESTAMP)
		),
		break_time AS (
			SELECT sche_id, room_id, doct_id, doc_name, data_appointment, start_hour, end_hour,
			CASE
				WHEN data_appointment = adata_appointment AND doct_id = adoct_id AND room_id = aroom_id AND end_hour = next_start_hour
				THEN 0
				ELSE transition_time::INT
			END	AS break_time,
			patient
			FROM range_calendar
		),
		interval_calendar AS (
			SELECT sche_id, room_id, doct_id, doc_name, data_appointment,
			to_char(end_hour::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS start_hour,
			to_char((end_hour::TIMESTAMP + (break_time ||' minutes')::INTERVAL)::TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS end_hour,
			jsonb_build_object('patientName', '', 'hourAppointment', '', 'status', '') AS patient,
			break_time
			FROM break_time, config_transition
		),
		interval_calendar_agg AS (
			SELECT sche_id, room_id, doct_id, doc_name, data_appointment, start_hour, end_hour,	jsonb_agg(patient) AS patient
			FROM interval_calendar
			GROUP BY sche_id, room_id, doct_id, doc_name, data_appointment , start_hour, end_hour
			ORDER BY data_appointment
		),
		union_calendar AS (
			SELECT sche_id, room_id, doct_id, doc_name, doc_treatment, data_appointment, start_hour, end_hour, patient
			FROM range_calendar
			UNION
			SELECT sche_id, room_id, doct_id, '', '', data_appointment, start_hour, end_hour, patient
			FROM interval_calendar_agg
		)
		SELECT sche_id, room_id, doct_id, doc_name, doc_treatment, data_appointment, start_hour, end_hour, patient
		FROM union_calendar
		WHERE ( EXTRACT(EPOCH FROM (end_hour::TIMESTAMP)::TIME) - EXTRACT(EPOCH FROM (start_hour::TIMESTAMP)::TIME) ) > 0
		OR start_hour::DATE != end_hour::DATE
		ORDER BY data_appointment, room_id, start_hour, end_hour
			`
	err := db.Select(&sch, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	return sch, nil
}

//ScheduleUnavailable verifying type of error
func ScheduleUnavailable(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

func validationsInsertSchedule(db service.DB, sch *m.Schedule) error {
	err := timeMinimum(db, sch.StartAt, sch.EndAt)
	if err != nil {
		return err
	}
	if sch.Plan == "Turn" {
		err := verifyIfTurnIsTrue(db, sch.StartAt, sch.EndAt)
		if err != nil {
			return err
		}
	}
	err = intervalTime(db, sch.StartAt, sch.EndAt)
	if err != nil {
		return err
	}
	err = dateLessThanToday(db, sch.StartAt)
	if err != nil {
		return err
	}
	err = doctorAvaliability(db, sch.StartAt, sch.EndAt, sch.DoctID)
	if err != nil {
		return err
	}
	return nil
}

func validationsUpdateSchedule(db service.DB, sch *m.Schedule) error {
	err := intervalTime(db, sch.StartAt, sch.EndAt)
	if err != nil {
		return err
	}

	err = roomAvailableToExtend(db, sch)
	if err != nil {
		return err
	}

	err = usingTransitionTimeToExtend(db, sch)
	if err != nil {
		return err
	}

	return nil
}

func timeMinimum(db service.DB, startAt, endAt time.Time) error {
	con := m.Config{}
	query := psql.Select("key", "value").
		From("config").
		Where(sq.Eq{"key": "schedule-minimum_time"})
	qSQL, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "Error generating get Config sql")
	}
	err = db.Get(&con, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "Error get Config sql")
		}
		return errors.New("Error Config sql not found")
	}

	mapMinTime := map[string]int{}
	err = json.Unmarshal(con.Value, &mapMinTime)
	if err != nil {
		return errors.New("Error Unmarshal minimum time")
	}
	diff := endAt.Sub(startAt)
	minTime := int64(mapMinTime["minimum_time"]) - int64(diff.Minutes())
	if minTime > 0 {
		return errors.New("Time minimum between hours not respected")
	}
	return nil
}

/* Should have a look to fix it */
func intervalTime(db service.DB, startAt, endAt time.Time) error {
	res := struct {
		Result bool `db:"valid_invalid" json:"validInterval"`
	}{}
	args := []interface{}{startAt, endAt}
	query := `WITH config_timezone AS (
				SELECT value->>'timezone' AS timezone
				FROM config
				WHERE KEY = 'timezone-local'
			),
			dates_to_local AS (
				SELECT ((($1)::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE timezone) AS start_at, ((($2)::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE timezone) AS end_at
				FROM config_timezone
			),
			validInterval AS (
				SELECT start_at, end_at,
				CASE
					WHEN TRIM(TO_CHAR(start_at::TIMESTAMP, 'day')::TEXT) = 'saturday' THEN (start_at::DATE || ' ' || ((((value->>'saturday'))::jsonb->0)->>'start') || ':00')::TIMESTAMP
					ELSE (start_at::DATE || ' ' || ((((value->>'monday'))::jsonb->0)->>'start') || ':00')::TIMESTAMP
				END AS start_lab,
				CASE
					WHEN TRIM(TO_CHAR(start_at::TIMESTAMP, 'day')::TEXT) = 'saturday' THEN
					(start_at::DATE || ' ' || ((((value->>'saturday'))::jsonb->0)->>'end' || ':00') )::TIMESTAMP
					ELSE (start_at::DATE || ' ' || ((((value->>'monday'))::jsonb->0)->>'end') || ':00')::TIMESTAMP
				END AS end_lab
				FROM config c, dates_to_local
				WHERE KEY = 'schedule-hour_config_flex'
			),
			local_timezone AS (
				SELECT start_at, end_at,
					(start_lab::DATE || ' ' || (((start_lab)::TIME AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)::TIMESTAMP AS start_lab,
					(end_lab::DATE || ' ' || (((end_lab)::TIME AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)::TIMESTAMP AS end_lab
				FROM validInterval, config_timezone cti
			)
			SELECT
				CASE
					WHEN start_at < start_lab OR end_at > end_lab THEN FALSE
					ELSE TRUE
				END AS valid_invalid
			FROM local_timezone`
	err := db.Get(&res, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "Error get Interval Time sql")
		}
		return errors.New("Error Interval Time sql not found")
	}
	if !res.Result {
		return errors.New("Interval Time of Schedule is not respecting hour of start/end")
	}
	return nil
}

func verifyIfTurnIsTrue(db service.DB, startAt, endAt time.Time) error {
	res := struct {
		Result int64 `db:"count" json:"count"`
	}{}
	args := []interface{}{startAt, endAt}
	query := `WITH config_timezone AS (
				SELECT value->>'timezone' AS timezone
				FROM config
				WHERE KEY = 'timezone-local'
			),
			dates_to_local AS (
				SELECT (($1::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_time,
				(($2::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_time
				FROM config_timezone cti
			),
			config_days AS (
				SELECT value
				FROM config
				WHERE KEY = 'schedule-hour_config'
			),
			config_result AS (
				SELECT row_number() OVER(ORDER BY a."key") AS id,
				a."key" AS "day",
				jsonb_array_elements(a.value) AS slot
				FROM config_days, jsonb_each(config_days.value) a
			),
			localtime_slot AS (
				SELECT "day",
					(((slot->>'start')::TIME AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME AS start_slot,
					(((slot->>'end')::TIME AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME AS end_slot
				FROM config_result, config_timezone cti
			)

			SELECT count(*) AS count
			FROM localtime_slot, dates_to_local
			WHERE TRIM(TO_CHAR($1::TIMESTAMP, 'day')::TEXT) = "day"
			AND (
					EXTRACT(EPOCH FROM (start_slot)::TIME) = EXTRACT(EPOCH FROM (start_time)::TIME)
					AND
					EXTRACT(EPOCH FROM (end_slot)::TIME) = EXTRACT(EPOCH FROM (end_time)::TIME)
				)
			`
	err := db.Get(&res, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "Error get Config sql")
		}
		return errors.New("Error Config sql not found")
	}
	if res.Result == 0 {
		return errors.New("Interval Schedule is not a Turn")
	}
	return nil
}

func dateLessThanToday(db service.DB, startAt time.Time) error {
	currentTime := time.Now()
	today := currentTime.Format("2006-01-02")
	dateToday, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}
	dataRequest := startAt.Format("2006-01-02")
	dateRequested, err := time.Parse("2006-01-02", dataRequest)
	if err != nil {
		return err
	}

	diff := dateRequested.Sub(dateToday)

	//Checking if dateRequested is same as dateToday and if hourRequested is greater than currentHour
	if dateRequested.Equal(dateToday) {
		hourDateRequested := startAt.Format("15:04:05")
		hourRequested, err := time.Parse("15:04:05", hourDateRequested)
		if err != nil {
			return err
		}
		hourDateToday := currentTime.Format("15:04:05")
		currentHour, err := time.Parse("15:04:05", hourDateToday)
		if err != nil {
			return err
		}

		diff := hourRequested.Sub(currentHour)
		mins := int(diff.Minutes())
		if mins < 0 {
			return errors.New("Hour requested less than current hour")
		}
	}

	days := int(diff.Hours() / 24)
	if days < 0 {
		return errors.New("Date requested less than current date")
	}
	return nil
}

func doctorAvaliability(db service.DB, startAt, endAt time.Time, doctID uuid.UUID) error {
	sch := []m.Schedule{}
	initialDate := startAt.Format("2006-01-02 15:04:05")
	finalDate := endAt.Format("2006-01-02 15:04:05")
	args := []interface{}{doctID, startAt, initialDate, finalDate}
	query := `
		WITH config_timezone AS (
			SELECT value->>'timezone' AS timezone
			FROM config
			WHERE KEY = 'timezone-local'
		),
		dates_to_local AS (
			SELECT (($3::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_time,
			(($4::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_time
			FROM config_timezone cti
		),
		doctor_schedules AS (
			SELECT sche_id, doct_id, room_id, ((start_at AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_at, ((end_at AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_at, plan, info, created_at, deleted_at
			FROM schedule, config_timezone cti
			WHERE doct_id = $1
			AND start_at::DATE = $2::DATE
			AND deleted_at IS NULL
		)

		SELECT sche_id, doct_id, room_id, start_at, end_at, plan, info, created_at, deleted_at
		FROM doctor_schedules, dates_to_local
		WHERE
		(
			--EXTRACT(EPOCH FROM (start_time::TIMESTAMP)::TIME)::INT BETWEEN EXTRACT(EPOCH FROM (start_at)::TIME)::INT AND EXTRACT(EPOCH FROM (end_at)::TIME)::INT
			(
				EXTRACT(EPOCH FROM (start_time::TIMESTAMP)::TIME)::INT >= EXTRACT(EPOCH FROM (start_at)::TIME)::INT
				AND
				EXTRACT(EPOCH FROM (start_time::TIMESTAMP)::TIME)::INT < EXTRACT(EPOCH FROM (end_at)::TIME)::INT
			)
			OR
			--EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT BETWEEN EXTRACT(EPOCH FROM (start_at)::TIME)::INT AND EXTRACT(EPOCH FROM (end_at)::TIME)::INT
			(
				--EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT >= EXTRACT(EPOCH FROM (start_at)::TIME)::INT
				EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT > EXTRACT(EPOCH FROM (start_at)::TIME)::INT
				AND
				EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT < EXTRACT(EPOCH FROM (end_at)::TIME)::INT
			)
			OR
			(
				EXTRACT(EPOCH FROM (start_at)::TIME)::INT BETWEEN EXTRACT(EPOCH FROM (start_time::TIMESTAMP)::TIME)::INT AND EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT
				AND
				EXTRACT(EPOCH FROM (end_at)::TIME)::INT BETWEEN EXTRACT(EPOCH FROM (start_time::TIMESTAMP)::TIME)::INT AND EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT
			)
		)
		`
	err := db.Select(&sch, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return err
	}
	if len(sch) > 0 {
		return errors.New("Doctor is not avaliable between these hours")
	}
	return nil
}

func outdoor(db service.DB, roomID uuid.UUID) (*m.Outdoor, error) {
	out := m.Outdoor{}
	query := psql.Select("doct_id", "doct_id", "name", "room_id", "label", "specialties", "avatar", "treatment").
		From("outdoor").
		LeftJoin("doc_specs USING (doct_id)").
		Prefix(`
			WITH doc_specs AS (
				SELECT doct_id, json_agg(name) AS specialties
				FROM doctor_specialty ds
				JOIN specialty using(spec_id)
				GROUP BY doct_id
			),
			outdoor AS (
				SELECT doct_id, doc."name", room_id, roo."label", doc.info->>'avatar' AS avatar, doc.info->>'treatment' AS treatment
				FROM schedule sch
				LEFT JOIN doctor doc USING (doct_id)
				LEFT JOIN room roo USING (room_id)
				WHERE roo.room_id = $1
				AND NOW() BETWEEN start_at AND end_at
				LIMIT 1
			)
		`)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Outdoor sql")
	}
	args = []interface{}{roomID}
	err = db.Get(&out, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Outdoor sql")
		}
		return nil, nil
	}
	return &out, nil
}

func roomAvailableToExtend(db service.DB, sch *m.Schedule) error {
	sched := []m.Schedule{}
	initialDate := sch.StartAt.Format("2006-01-02 15:04:05")
	finalDate := sch.EndAt.Format("2006-01-02 15:04:05")
	args := []interface{}{sch.ScheID, sch.RoomID, initialDate, finalDate}
	query := `
		WITH config_timezone AS (
			SELECT value->>'timezone' AS timezone
			FROM config
			WHERE KEY = 'timezone-local'
		),
		dates_to_local AS (
			SELECT (($3::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_time,
			(($4::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_time
			FROM config_timezone cti
		),
		schedule_localtime AS (
			SELECT sche_id, doct_id, room_id,
			((start_at AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_at,
			((end_at AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_at, plan, info, created_at, deleted_at
			FROM schedule, dates_to_local, config_timezone cti
			WHERE sche_id != $1
			AND room_id = $2
			AND start_at::DATE = start_time::DATE
			AND deleted_at IS NULL
		)

		SELECT sche_id, doct_id, room_id, start_at, end_at, plan, info, created_at, deleted_at
		FROM schedule_localtime, dates_to_local
		WHERE EXTRACT(EPOCH FROM (end_time::TIMESTAMP)::TIME)::INT BETWEEN EXTRACT(EPOCH FROM (start_at)::TIME)::INT AND EXTRACT(EPOCH FROM (end_at)::TIME)::INT
	`
	err := db.Select(&sched, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return err
	}

	if len(sched) > 0 {
		return errors.New("Room is not available to extend your schedule")
	}
	return nil
}

/* Should have a look to fix it */
func usingTransitionTimeToExtend(db service.DB, sch *m.Schedule) error {
	usb := m.UsableTransition{}
	args := []interface{}{sch.ScheID, sch.EndAt}
	query := `
		WITH config_timezone AS (
			SELECT value->>'timezone' AS timezone
			FROM config
			WHERE "key" = 'timezone-local'
		),
		dates_to_local AS (
			SELECT (($2::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_time
			FROM config_timezone cti
		),
		config_days AS (
			SELECT value
			FROM config
			WHERE "key" = 'schedule-hour_config_flex'
		),
		config_usable AS (
			SELECT value
			FROM config
			WHERE "key" = 'usable-transition_time'
		),
		config_transition AS (
			SELECT value
			FROM config
			WHERE "key" = 'schedule-transition_time'
		),
		config_result AS (
			SELECT row_number() OVER(ORDER BY a."key") AS id,
			a."key" AS "day",
			jsonb_array_elements(a.value) AS slot
			FROM config_days, jsonb_each(config_days.value) a
		),
		slots_by_day AS (
			SELECT id, "day",
			json_build_object('start', SUBSTRING((((((slot->>'start')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)::TEXT, 0, 6), 'end', SUBSTRING((((((slot->>'end')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)::TEXT, 0, 6) ) AS slot,
			EXTRACT(EPOCH FROM ((((slot->>'start')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)::INT AS start_slot,
			EXTRACT(EPOCH FROM ((((slot->>'end')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)::INT AS end_slot
			FROM config_result, config_timezone cti
		),
		last_slot AS (
			SELECT id, "day", max(start_slot) AS max_start_slot, max(end_slot) AS max_end_slot
			FROM slots_by_day
			GROUP BY id, "day"
		),
		schedule_with_slot AS (
			SELECT sche_id, start_at, end_at, sbd.slot->>'start' AS start_slot, sbd.slot->>'end' AS end_slot
			FROM schedule
			JOIN last_slot ls ON EXTRACT(EPOCH FROM (schedule.start_at)::TIME) >= ls.max_start_slot
							AND EXTRACT(EPOCH FROM (schedule.end_at)::TIME) <= ls.max_end_slot
							AND TRIM(TO_CHAR(schedule.start_at, 'day'))::TEXT = ls."day"
			JOIN slots_by_day sbd ON EXTRACT(EPOCH FROM (schedule.start_at)::TIME) >= sbd.start_slot
							AND EXTRACT(EPOCH FROM (schedule.end_at)::TIME) <= sbd.end_slot
							AND TRIM(TO_CHAR(schedule.start_at, 'day'))::TEXT = sbd."day"
			WHERE deleted_at IS NULL
			AND sche_id = $1
		),
		getting_configs AS (
			SELECT sche_id, start_at, end_at, start_slot::TIME, end_slot::TIME, cu.value->>'usable' AS usable, ct.value->>'transition_time'::TEXT AS transition_time
			FROM schedule_with_slot, config_transition ct, config_usable cu
		),
		limit_schedule AS (
			SELECT sche_id, start_at, end_at, usable, transition_time,
				CASE
					WHEN usable = 'true' THEN end_at::TIME + (transition_time ||' minutes')::interval
					ELSE end_at::TIME
				END AS max_hour_schedule
			FROM getting_configs
		)
		SELECT sche_id, start_at, end_at, usable, transition_time,
			CASE WHEN (EXTRACT(EPOCH FROM (end_time)::TIME)::INT) > (EXTRACT(EPOCH FROM (max_hour_schedule)::TIME)::INT) THEN FALSE
				ELSE TRUE
			END is_able_to_extend
		FROM limit_schedule, dates_to_local
		`
	err := db.Get(&usb, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if !usb.IsAbleToExtend {
		return errors.New("Schedule is not able to extend")
	}
	return nil
}

func needBathroom(db service.DB, doctID uuid.UUID) (bool, string, error) {
	checkBath := struct {
		NeedBathroom      bool   `db:"need_bathroom"`
		BathroomTreatment string `db:"bathroom_treatment"`
	}{}
	args := []interface{}{doctID}
	query := `
		WITH config_transition AS (
			SELECT value->>'bathroom_treatment' AS bathroom_treatment
			FROM config
			WHERE KEY = 'schedule-bathroom_treatment'
		),
		check_bathroom AS (
			SELECT doc.doct_id,
			CASE
				WHEN array_agg((spe.info->>'needBathroom')::BOOL) <> '{}' THEN array_agg((spe.info->>'needBathroom')::BOOL)
				ELSE array_agg(FALSE)
			END
			AS spec_need_bathroom
			FROM doctor doc
			LEFT JOIN doctor_specialty ds USING (doct_id)
			LEFT JOIN specialty spe USING (spec_id)
			WHERE doc.doct_id = $1
			GROUP BY doct_id
		)
		SELECT bathroom_treatment,
			CASE
				WHEN TRUE = ANY (spec_need_bathroom) THEN TRUE
				ELSE FALSE
			END AS need_bathroom
		FROM check_bathroom, config_transition
	`
	err := db.Get(&checkBath, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return checkBath.NeedBathroom, checkBath.BathroomTreatment, err
		}
		return checkBath.NeedBathroom, checkBath.BathroomTreatment, err
	}
	return checkBath.NeedBathroom, checkBath.BathroomTreatment, nil
}

func filterRooms(needBathroom bool, bathroomTreatment string) string {
	whereBathroom := ""
	if needBathroom == true && bathroomTreatment == "preferencial" {
		whereBathroom = `ORDER BY (roo.info->>'hasBathroom')::BOOL DESC`
	}
	if needBathroom == true && bathroomTreatment == "obligation" {
		whereBathroom = `AND (roo.info->>'hasBathroom')::BOOL = true`
	} else {
		whereBathroom = `ORDER BY (roo.info->>'hasBathroom')::BOOL`
	}
	return whereBathroom
}

func buildOrderBy(prefix string, orderBy, order *string) (string, error) {
	allowedColumns := []string{"start_at"}
	allowedOrder := []string{"ASC", "DESC", "asc", "desc"}

	if orderBy == nil {
		return "", nil
	}
	if !inArray(*orderBy, allowedColumns) {
		return "", errors.New("Invalid order options")
	}
	a := prefix + *orderBy + " "
	if order != nil {
		if !inArray(*order, allowedOrder) {
			return "", errors.New("Invalid order options")
		}
		a += *order
	}
	return a, nil
}

func inArray(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
