package avaliability

import (
	"database/sql"

	"github.com/gofrs/uuid"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Checker service to list avaliability
type Checker struct {
	DB *sqlx.DB
}

//Run Avaliability
func (c *Checker) Run(fAvaliability m.FilterAvaliability) ([]m.Avaliability, error) {
	u, err := checkAvaliability(c.DB, fAvaliability)
	return u, err
}

/* Create a new Schedule to database */
func checkAvaliability(db service.DB, f m.FilterAvaliability) ([]m.Avaliability, error) {
	if f.DoctID == nil {
		return nil, errors.New("doctor required")

	}
	ava := []m.Avaliability{}
	args := []interface{}{f.StartDate, f.EndDate, *f.DoctID}

	filterRooms := ""

	needBathroom, bathroomTreatment, err := needBathroom(db, *f.DoctID)
	if err != nil {
		return nil, err
	}
	filterRooms = filterRoomsBuilder(needBathroom, bathroomTreatment)

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
				SELECT (($1::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS start_time,
				(($2::TIMESTAMP AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone) AS end_time
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
								WHEN doct_id = $3 THEN end_at
								ELSE (end_at + (value->>'transition_time' ||' minutes')::INTERVAL)::TIMESTAMP
				END AS end_at,
				slot,
				ROW_NUMBER () OVER (
								PARTITION BY room_id, start_at::date, slot
								ORDER BY start_at ASC
				)
				FROM config_transition ct, schedule_local s
				JOIN slots_by_day sbd ON EXTRACT(EPOCH FROM (s.start_at)::TIME) >= sbd.start_slot
								AND EXTRACT(EPOCH FROM (s.end_at)::TIME) <= sbd.end_slot
								AND TRIM(TO_CHAR(s.start_at, 'day'))::TEXT = "day"
				WHERE start_at::DATE BETWEEN '2020-11-09 23:00:59.999999999'::DATE AND '2020-11-14 23:00:59.999999999'::DATE
				AND deleted_at IS NULL
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
				SELECT * FROM scheduled
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
				WHERE ( EXTRACT(EPOCH FROM (slot_free->>'end')::TIME)::INT - EXTRACT(EPOCH FROM (slot_free->>'start')::TIME)::INT ) > 0
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
				LEFT JOIN slots_free_by_room sfbr ON sfbr.room_id = swd.room_id AND sfbr.date_of_month = swd.days
				AND	( EXTRACT(EPOCH FROM (sfbr.slot_free->>'start')::TIME)::INT >= EXTRACT(EPOCH FROM (swd.slot->>'start')::TIME)::INT
				AND	EXTRACT(EPOCH FROM (sfbr.slot_free->>'end')::TIME)::INT <= EXTRACT(EPOCH FROM (swd.slot->>'end')::TIME)::INT )
			),
			closest_rooms AS (
				SELECT ABS(EXTRACT(epoch FROM start_at - start_time)) as closest_schedule, room_id
				FROM schedule_local s
				JOIN dates_to_local d on s.start_at::DATE = d.start_time::DATE
				AND doct_id = $3
			),
			ordered_rooms AS (
				SELECT * FROM closest_rooms
				UNION SELECT 100000000000, room_id FROM room r where not exists (SELECT room_id FROM closest_rooms c where c.room_id = r.room_id)
			),
			slots_not_full_by_day_ordered AS (
				SELECT *
				FROM slots_not_full_by_day s
				ORDER BY (
								SELECT MIN(od.closest_schedule) FROM ordered_rooms od WHERE s.room_id = od.room_id
				) ASC
			),
			room_for_insert_flex AS (
				SELECT room_id, days, slot
				FROM slots_not_full_by_day_ordered, dates_to_local dtl
			),
			slots_back_to_utc AS (
			SELECT DISTINCT ON (slot->>'start', slot->>'end', days) room_id, days,
			jsonb_build_object('start', ((days+(slot->>'start')::TIME)::TIMESTAMP AT TIME ZONE cti.timezone)::TIME, 'end', ((days+(slot->>'end')::TIME)::TIMESTAMP AT TIME ZONE cti.timezone)::TIME) as slot
			FROM room_for_insert_flex, config_timezone cti
			),
			results AS (
				select doct_id, d.name as doct_name, days::TIMESTAMP AT TIME ZONE 'UTC' as "date",
				jsonb_agg(slot) as slots
				from doctor d
				join slots_back_to_utc ON doct_id = $3
				group by doct_id, d.name, days
			)

			select * from results
			`
	err = db.Select(&ava, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list Avaliability sql")
		}
		return nil, nil
	}
	return ava, nil
}

func filterRoomsBuilder(needBathroom bool, bathroomTreatment string) string {
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
