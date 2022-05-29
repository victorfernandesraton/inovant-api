package schedule

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jasonlvhit/gocron"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var sched *gocron.Scheduler

//var psql sq.StatementBuilderType

//Scheduler service to stagger schedules
type Scheduler struct {
	DB     *sqlx.DB
	Logger *log.Logger
}

func init() {
	sched = gocron.NewScheduler()
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

// Start resets cron timer and loads from database config
func (s *Scheduler) Start() chan bool {
	s.Run()
	x := gocron.NewScheduler()
	x.Clear()
	x.Every(60).Minutes().Do(s.Run)
	<-x.Start()
	return sched.Start()
}

//Run service to run Scheduling algorithm
func (s *Scheduler) Run() error {
	var err error
	/* For por dias +30 */
	//AddDate(years int, months int, days int) Time
	date := time.Now()
	for j := 1; j <= 30; j++ {
		if j == 1 {
			date = time.Now()
		} else {
			date = date.AddDate(0, 0, 1)
		}
		err = schedulingAlgorithm(s.DB, s.Logger, date)
	}
	//err = schedulingAlgorithm(s.DB, s.Logger, date.AddDate(0, 0, 1))
	if s.Logger != nil {
		if err != nil {
			s.Logger.Println(err)
			return err
		}
	}
	return nil

}

func schedulingAlgorithm(db service.DB, log *log.Logger, date time.Time) error {

	schedSlots, err := gettingWholeSlotsByToday(db, date)
	if err != nil {
		return err
	}
	sched, err := gettingScheduledToday(db, date)
	if err != nil {
		return err
	}
	mapRoomSched := map[string][]m.SchedGroup{}
	mapRoomSchedBathroom := map[string][]m.SchedGroup{}
	scheduledMap := map[string][]m.SchedGroup{}
	scheduledMapBathroom := map[string][]m.SchedGroup{}
	for _, s := range schedSlots {
		if s.HasBathroom {
			mapRoomSchedBathroom[s.RoomID.String()] = s.Slots
			scheduledMapBathroom[s.RoomID.String()] = []m.SchedGroup{}
		} else {
			mapRoomSched[s.RoomID.String()] = s.Slots
			scheduledMap[s.RoomID.String()] = []m.SchedGroup{}
		}

	}
	arrScheduled := []m.Slot{}
	arrScheduledBathroom := []m.Slot{}
	for _, s := range sched {
		if s.SlotSched.NeedBathroom {
			arrScheduledBathroom = append(arrScheduledBathroom, s.SlotSched)
		} else {
			arrScheduled = append(arrScheduled, s.SlotSched)
		}
	}
	countSchedules := len(arrScheduled) + len(arrScheduledBathroom)
	/* Scheduler Rooms with Bathroom */
	scheduledMapSlotBathroom, err := scheduler(db, log, date, mapRoomSchedBathroom, scheduledMapBathroom, arrScheduledBathroom, countSchedules)
	if err != nil {
		return err
	}
	roomsEmpty := returnRoomWithBathEmpty(scheduledMapSlotBathroom, scheduledMapBathroom)
	if len(roomsEmpty) > 0 {
		for k, v := range roomsEmpty {
			scheduledMap[k] = v
		}
	}
	notFoundBath := checkAllSchedulesWithRoom(arrScheduledBathroom, scheduledMapSlotBathroom)
	if len(notFoundBath) > 0 {
		for _, v := range notFoundBath {
			arrScheduled = append(arrScheduled, v)
		}
	}
	/* Scheduler Rooms without Bathroom */
	scheduledMapSlot, err := scheduler(db, log, date, mapRoomSched, scheduledMap, arrScheduled, countSchedules)
	if err != nil {
		return err
	}
	notFound := checkAllSchedulesWithRoom(arrScheduled, scheduledMapSlot)
	if len(notFound) > 0 {
		err = errors.New(strconv.Itoa(len(notFound)) + " slots not scheduled to any room")
		log.Println(err, "- Date:", date)
		return err
	}

	scheduleAll := map[string][]m.Slot{}
	for k, v := range scheduledMapSlot {
		scheduleAll[k] = v
	}
	for k, v := range scheduledMapSlotBathroom {
		scheduleAll[k] = v
	}

	/* Function to update database */
	err = updatingSchedule(db, scheduleAll, countSchedules, date)
	if err != nil {
		log.Println(err, "- Date:", date)
		return err
	}
	return nil

}

func returnRoomWithBathEmpty(scheduleMap map[string][]m.Slot, scheduleMapRooms map[string][]m.SchedGroup) map[string][]m.SchedGroup {
	scheduleMapEmpty := map[string][]m.SchedGroup{}
	for k, v := range scheduleMapRooms {
		if hasEmptyRoom(k, scheduleMap) {
			continue
		}
		scheduleMapEmpty[k] = v
	}
	return scheduleMapEmpty
}

func hasEmptyRoom(needle string, haystack map[string][]m.Slot) bool {
	for room := range haystack {
		if room == needle {
			return true
		}
	}
	return false
}

func scheduler(db service.DB, log *log.Logger, date time.Time, mapRoomSched, scheduledMap map[string][]m.SchedGroup, arrScheduled []m.Slot, countSchedules int) (map[string][]m.Slot, error) {
	scheduledMapSlot := map[string][]m.Slot{}
	schedGroup, err := agroupSchedules(arrScheduled)
	if err != nil {
		return scheduledMapSlot, err
	}

	keyRoom := []string{}
	for k := range mapRoomSched {
		keyRoom = append(keyRoom, k)
	}
	/* Slots ->> Rooms */
	for _, i := range schedGroup {
		for _, k := range keyRoom {
			slots, ok := roomFitForSlot(i, mapRoomSched[k])
			if ok {
				mapRoomSched[k] = slots
				scheduledMap[k] = append(scheduledMap[k], i)
				break
			}
		}
	}
	for k := range scheduledMap {
		for _, v := range scheduledMap[k] {
			for _, p := range arrScheduled {
				for _, t := range v.Schedules {
					if t == p.ScheID {
						slot := m.Slot{
							StartAt:      p.StartAt,
							EndAt:        p.EndAt,
							ScheID:       p.ScheID,
							DoctID:       p.DoctID,
							NeedBathroom: p.NeedBathroom,
						}
						scheduledMapSlot[k] = append(scheduledMapSlot[k], slot)
					}
				}
			}
		}
	}
	return scheduledMapSlot, err
}

func agroupSchedules(arrScheduled []m.Slot) ([]m.SchedGroup, error) {
	schedGroup := []m.SchedGroup{}
	for i := 0; i < len(arrScheduled); i++ {
		res, err := collectionSchedules(arrScheduled[i:])
		if len(res.Schedules) > 1 {
			i += (len(res.Schedules) - 1)
		}
		if err != nil {
			return schedGroup, err
		}
		schedGroup = append(schedGroup, res)
	}
	return schedGroup, nil
}

func collectionSchedules(arrScheduled []m.Slot) (m.SchedGroup, error) {
	if len(arrScheduled) == 0 {
		return m.SchedGroup{}, errors.New("Unexpected error on collection schedules")
	}
	schedGroup := m.SchedGroup{
		StartAt:   arrScheduled[0].StartAt,
		EndAt:     arrScheduled[0].EndAt,
		Schedules: []uuid.UUID{arrScheduled[0].ScheID},
	}
	if len(arrScheduled) == 1 {
		return schedGroup, nil
	}
	for i, sched := range arrScheduled {
		if i == 0 {
			continue
		}
		if sched.DoctID == arrScheduled[i-1].DoctID && sched.StartAt == arrScheduled[i-1].EndAt {
			schedGroup.Schedules = append(schedGroup.Schedules, sched.ScheID)
			schedGroup.EndAt = sched.EndAt
		} else {
			break
		}
	}
	return schedGroup, nil
}

func checkAllSchedulesWithRoom(arrScheduled []m.Slot, scheduledMap map[string][]m.Slot) []m.Slot {
	found := []uuid.UUID{}
	for _, i := range arrScheduled {
		for k := range scheduledMap {
			for _, z := range scheduledMap[k] {
				if i.ScheID == z.ScheID {
					found = append(found, z.ScheID)
				}
			}
		}
	}

	arrScheduledNotFound := []m.Slot{}
	if len(found) != len(arrScheduled) {
		for _, t := range arrScheduled {
			if hasUUID(t.ScheID, found) {
				continue
			}
			arrScheduledNotFound = append(arrScheduledNotFound, t)
		}
	}

	return arrScheduledNotFound
}

func hasUUID(needle uuid.UUID, haystack []uuid.UUID) bool {
	for _, uid := range haystack {
		if uid == needle {
			return true
		}
	}
	return false
}

//updatingSchedule updating schedule table by scheduling algorithm
func updatingSchedule(db service.DB, scheduledMap map[string][]m.Slot, countSchedules int, date time.Time) error {
	roomSched := []m.RoomSched{}
	for key, value := range scheduledMap {
		s := m.RoomSched{}
		for i := range value {
			s.RoomID = key
			s.ScheID = append(s.ScheID, value[i].ScheID)
		}
		roomSched = append(roomSched, s)
	}
	err := upsertSchedule(db, roomSched, countSchedules, date)
	if err != nil {
		return err
	}
	return nil
}

//upsertSchedule query to update Schedule Table by scheduling
func upsertSchedule(db service.DB, roomSched []m.RoomSched, countSchedules int, date time.Time) error {
	a, err := json.Marshal(roomSched)
	if err != nil {
		return errors.Wrap(err, "Error inserting/updating Schedules into the database")
	}
	args := []interface{}{types.JSONText(a), countSchedules, date}
	qSQL := `
		 WITH countSchedules as (
			SELECT count(*)
			FROM schedule
			WHERE start_at::DATE = $3::DATE
			AND start_at >= now()
			AND deleted_at IS NULL
			GROUP BY start_at::DATE
		),
		checkError as (
			SELECT f_exec('ERROR! SCHEDULE ARRAY NOT EQUALS')
			FROM countSchedules
			WHERE count > $2
		),
		getRoomSched AS (
			SELECT * FROM jsonb_array_elements($1) AS value
		),
		orgRoomSched AS (
			SELECT value->>'RoomID' AS room_id,
			value->>'ScheID' AS arr_sche_id
			FROM getRoomSched
			WHERE value->>'RoomID' <> '' OR value->>'RoomID' <> NULL
		),
		groupRoomSched AS (
			SELECT room_id, jsonb_array_elements(arr_sche_id::jsonb)::text AS sche_id
			FROM orgRoomSched
			GROUP BY room_id, arr_sche_id
		),
		collectionRoomSched as(
			SELECT room_id::uuid, REPLACE(sche_id, '"', '')::uuid AS sche_id
			FROM groupRoomSched
		)

		UPDATE schedule s
		SET room_id = crs.room_id
		FROM collectionRoomSched crs
		LEFT JOIN checkError ON 1=1
		WHERE crs.sche_id = s.sche_id
		`
	_, err = db.Exec(qSQL, args...)
	if err != nil {
		return err
	}
	return nil
}

//roomFitForSlot returning slots free and it is fit for any slot
func roomFitForSlot(scheduled m.SchedGroup, slots []m.SchedGroup) ([]m.SchedGroup, bool) {
	newSlot := m.SchedGroup{}
	correctSlot := m.SchedGroup{}
	flag := false
	twoSlots := false
	for ind := range slots {
		if scheduled.StartAt >= slots[ind].StartAt && scheduled.EndAt <= slots[ind].EndAt {
			correctSlot.StartAt = slots[ind].StartAt
			correctSlot.EndAt = slots[ind].EndAt
			correctSlot.Schedules = slots[ind].Schedules

			if scheduled.StartAt > correctSlot.StartAt && scheduled.EndAt <= correctSlot.EndAt {
				newSlot.StartAt = correctSlot.StartAt
				newSlot.EndAt = scheduled.StartAt

				slots[ind] = newSlot
				twoSlots = true
			}
			if scheduled.EndAt < correctSlot.EndAt && scheduled.StartAt >= correctSlot.StartAt {
				newSlot.StartAt = scheduled.EndAt
				newSlot.EndAt = correctSlot.EndAt

				if twoSlots {
					slots = append(slots, newSlot)
				} else {
					slots[ind] = newSlot
				}
			}
			if scheduled.StartAt == correctSlot.StartAt && scheduled.EndAt == correctSlot.EndAt {
				slots[ind] = m.SchedGroup{StartAt: "00:00", EndAt: "00:00"}
			}
			flag = true
		}
	}

	return slots, flag
}

//gettingScheduledToday returning all schedules booked for today
func gettingScheduledToday(db service.DB, today time.Time) ([]m.Scheduling, error) {
	sched := []m.Scheduling{}
	dateToday := today.Format("2006-01-02")
	query := psql.Select("slot_sched").
		From("appointments_scheduled").
		Prefix(`
			WITH config_days AS (
				SELECT value
				FROM config
				WHERE "key" = 'schedule-hour_config_flex'
			),
			config_timezone AS (
				SELECT value->>'timezone' AS timezone
				FROM config
				WHERE "key" = 'timezone-local'
			),
			config_transition AS (
				SELECT value
				FROM config
				WHERE "key" = 'schedule-transition_time'
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
			building_schedules AS (
				SELECT room_id, sche_id, doct_id, start_at::TIME  AS start_at, end_at::TIME AS end_at,
					CASE
						WHEN array_agg((spe.info->>'needBathroom')::bool) <> '{}' THEN array_agg((spe.info->>'needBathroom')::bool)
						ELSE array_agg(FALSE)
					END
					AS spec_need_bathroom
				FROM schedule
				LEFT JOIN doctor_specialty USING (doct_id)
				LEFT JOIN specialty spe USING (spec_id)
				WHERE start_at::DATE = ?::DATE
				AND start_at >= NOW()
				AND deleted_at IS NULL
				GROUP BY (room_id, sche_id, doct_id, start_at::TIME, end_at::TIME)
				ORDER BY doct_id, start_at::TIME
			),
			schedule_bathroom AS (
				SELECT room_id, sche_id, doct_id, start_at, end_at,
				CASE
					WHEN TRUE = ANY (spec_need_bathroom) THEN TRUE
					ELSE FALSE
				END AS need_bathroom
				FROM building_schedules
			),
			check_after AS (
				SELECT room_id, sche_id, doct_id, start_at, end_at, need_bathroom,
				LEAD (doct_id) OVER (PARTITION BY doct_id) AS next_doct_id,
				LEAD (start_at) OVER (PARTITION BY doct_id) AS next_start_at
				FROM schedule_bathroom
			),
			building_end_at AS (
				SELECT room_id, sche_id, doct_id, start_at,
					CASE
						WHEN EXTRACT(EPOCH FROM ((((end_at)::TIME AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)) = EXTRACT(EPOCH FROM ((cd.slot->>'end')::TIME))
							THEN end_at
						WHEN EXTRACT(EPOCH FROM ((((end_at)::TIME AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TIME)) < EXTRACT(EPOCH FROM ((cd.slot->>'end')::TIME)) AND end_at = next_start_at
							THEN end_at
						ELSE (end_at + (ct.value->>'transition_time' ||' minutes')::INTERVAL)::TIME
					END  AS end_at,
					need_bathroom
				FROM check_after, config_transition ct, config_result cd, config_timezone cti
				WHERE cd."day" = TRIM(TO_CHAR((?)::TIMESTAMP, 'day'))::TEXT
			),
			appointments_scheduled AS (
				SELECT room_id,
				jsonb_build_object('start', substring(((start_at AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TEXT, 0, 6), 'end', substring(((end_at AT TIME ZONE 'UTC') AT TIME ZONE cti.timezone)::TEXT, 0, 6), 'scheID', sche_id, 'doctID', doct_id, 'needBathroom', need_bathroom) AS slot_sched
				FROM building_end_at, config_timezone cti
			)`, dateToday, dateToday,
		)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Appointments scheduled sql")
	}
	err = db.Select(&sched, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list Appointments scheduled sql")
		}
		return nil, err
	}
	return sched, nil
}

//gettingWholeSlotsByToday returning all rooms with whole slot for today
func gettingWholeSlotsByToday(db service.DB, today time.Time) ([]m.SchedulingSlots, error) {
	schedSlots := []m.SchedulingSlots{}
	dateToday := today.Format("2006-01-02")
	query := psql.Select("r.room_id", "json_agg(st.slot) AS slots",
		"CASE WHEN (r.info->>'hasBathroom')::BOOL = true THEN true WHEN (r.info->>'hasBathroom')::BOOL = false THEN false ELSE false END AS has_bathroom").
		From("slot_timezone st").
		Join("room r ON TRIM(TO_CHAR(?::DATE, 'day'))::TEXT = st.day").
		Where("r.inactive_at IS NULL").
		GroupBy("r.room_id").
		Prefix(`
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
			config_result AS (
				SELECT row_number() OVER(ORDER BY a."key") AS id,
				a."key" AS "day",
				jsonb_array_elements(a.value) AS slot
				FROM config_days, jsonb_each(config_days.value) a
			),
			slot_timezone AS (
				SELECT id, "day",
				jsonb_build_object('start', SUBSTRING((( ((slot->>'start')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE ct.timezone)::TEXT, 0, 6), 'end', SUBSTRING((( ((slot->>'end')::TIME) AT TIME ZONE 'UTC') AT TIME ZONE ct.timezone)::TEXT, 0, 6)) AS slot
				FROM config_result, config_timezone ct
			)`, dateToday,
		)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Appointments scheduled sql")
	}
	err = db.Select(&schedSlots, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list Appointments scheduled sql")
		}
		return nil, err
	}
	return schedSlots, nil
}
