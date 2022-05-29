package room

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Creator service to create new Room
type Creator struct {
	DB *sqlx.DB
}

//Run create new Room
func (c *Creator) Run(rom *m.Room) (*m.Room, error) {
	roomID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Room uuid")
	}
	rom.RoomID = roomID
	u, err := createRoom(c.DB, rom)
	return u, err
}

//Lister service to return Room
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Room by Filter
func (l *Lister) Run(f m.FilterRoom) ([]m.Room, error) {
	u, err := listRoom(l.DB, f)
	return u, err
}

//Getter service to return Room
type Getter struct {
	DB *sqlx.DB
}

//Run return a Room by room_id
func (g *Getter) Run(roomID uuid.UUID) (*m.Room, error) {
	u, err := getRoom(g.DB, roomID)
	return u, err
}

//Updater service to update Room
type Updater struct {
	DB *sqlx.DB
}

//Run update Room data
func (g *Updater) Run(rom *m.Room) (*m.Room, error) {
	u, err := updateRoom(g.DB, rom)
	return u, err
}

//Deleter service to soft delete Room
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete Room by room_id
func (d *Deleter) Run(roomID uuid.UUID) (*m.Room, error) {
	u, err := deleteRoom(d.DB, roomID)
	return u, err
}

/* Create a new Room to database */
func createRoom(db service.DB, rom *m.Room) (*m.Room, error) {
	query := psql.Insert("room").
		Columns("room_id", "label", "inactive_at", "info").
		Values(rom.RoomID, rom.Label, rom.InactiveAt, rom.Info).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Room sql")
	}

	err = db.Get(rom, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting Room in database")
	}
	return rom, nil
}

/* Return a list of Room by filters */
func listRoom(db service.DB, f m.FilterRoom) ([]m.Room, error) {
	rom := []m.Room{}
	query := psql.Select("room_id", "label", "inactive_at", "info").
		From("room")

	if f.RoomID != nil {
		query = query.Where(`room_id = ?`, f.RoomID)
	}
	if f.Label != nil {
		query = query.Where(`label ILIKE ?`, `%`+*f.Label+`%`)
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list of Rooms sql")
	}
	err = db.Select(&rom, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list of Rooms sql")
		}
		return nil, nil
	}
	return rom, nil
}

/* Return a Room by room_id */
func getRoom(db service.DB, roomID uuid.UUID) (*m.Room, error) {
	rom := m.Room{}
	query := psql.Select("room_id", "label", "inactive_at", "info").
		From("room").
		Where(sq.Eq{"room_id": roomID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Room sql")
	}
	err = db.Get(&rom, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Room sql")
		}
		return nil, nil
	}
	return &rom, nil
}

/* Update Room to database by room_id */
func updateRoom(db service.DB, rom *m.Room) (*m.Room, error) {
	query := psql.Update("room").
		Set("label", rom.Label).
		Set("inactive_at", rom.InactiveAt).
		Set("info", rom.Info).
		Suffix("RETURNING *").
		Where(sq.Eq{"room_id": rom.RoomID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Room update sql")
	}

	err = db.Get(rom, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Room update sql")
	}

	return rom, nil
}

/* Delete Room to database by room_id */
func deleteRoom(db service.DB, roomID uuid.UUID) (*m.Room, error) {
	err := verifyRoomsToInactive(db, roomID)
	if err != nil {
		return nil, err
	}
	rom := m.Room{}
	query := psql.Update("room").
		Set("inactive_at", time.Now()).
		Suffix("RETURNING *").
		Where(sq.Eq{"room_id": roomID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &rom, errors.Wrap(err, "Error generating delete Room sql")
	}
	err = db.Get(&rom, qSQL, args...)
	if err != nil {
		return &rom, errors.Wrap(err, "Error delete Room sql")
	}
	return &rom, nil
}

func verifyRoomsToInactive(db service.DB, roomID uuid.UUID) error {
	res := struct {
		Result int64 `db:"count_schedules" json:"countSchedules"`
	}{}
	query := psql.Select("count(sche_id) as count_schedules").
		From("schedule").
		Where(sq.Eq{"room_id": roomID}).
		Where("deleted_at IS NULL").
		Where(sq.GtOrEq{"start_at": time.Now()})
	qSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}
	err = db.Get(&res, qSQL, args...)
	if err != nil {
		return err
	}
	if res.Result > 0 {
		return errors.New("Exists future schedules, so it is not possible inactive this room")
	}
	return nil
}
