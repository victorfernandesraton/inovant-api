package appointment

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

//Creator service to create new Appointment
type Creator struct {
	DB *sqlx.DB
}

//Run create new Appointment
func (c *Creator) Run(app *m.Appointment, doctID *uuid.UUID) (*m.Appointment, error) {
	appoID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Appointment uuid")
	}
	app.AppoID = appoID
	u, err := createAppointment(c.DB, app, doctID)
	return u, err
}

//Lister service to return Appointment
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Appointment by Filter
func (l *Lister) Run(doctID *uuid.UUID, f m.FilterAppointment) ([]m.Appointment, error) {
	u, err := listAppointment(l.DB, doctID, f)
	return u, err
}

//Getter service to return Appointment
type Getter struct {
	DB *sqlx.DB
}

//Run return a Appointment by sche_id
func (g *Getter) Run(doctID *uuid.UUID, appoID uuid.UUID) (*m.Appointment, error) {
	u, err := getAppointment(g.DB, doctID, appoID)
	return u, err
}

//Updater service to update Appointment
type Updater struct {
	DB *sqlx.DB
}

//Run update Appointment data
func (g *Updater) Run(app *m.Appointment, doctID *uuid.UUID) (*m.Appointment, error) {
	u, err := updateAppointment(g.DB, app, doctID)
	return u, err
}

//Deleter service to soft delete Appointment
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete Appointment by sche_id
func (d *Deleter) Run(appoID uuid.UUID) (*m.Appointment, error) {
	u, err := deleteAppointment(d.DB, appoID)
	return u, err
}

/* Create a new Appointment to database */
func createAppointment(db service.DB, app *m.Appointment, doctID *uuid.UUID) (*m.Appointment, error) {
	args := []interface{}{app.AppoID, app.StartAt, app.ScheID, app.PatiID, app.Type, app.Status}
	filter := ""

	if doctID != nil {
		args = append(args, doctID)
		filter = ` AND doct_id = $7`
	}

	query := `
			WITH results AS (
				SELECT $1::UUID, $2::TIMESTAMP, sche_id, pati_id, $5::TEXT, $6::TEXT, doct_id
				FROM schedule
				JOIN patient USING (doct_id)
				WHERE sche_id = $3 AND pati_id = $4` + filter + `
			)

			INSERT INTO appointment(appo_id, start_at, sche_id, pati_id, type, status)
			SELECT $1, $2, sche_id, pati_id, $5, $6
			FROM results
			RETURNING *
			`
	err := db.Get(app, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return app, nil
}

/* Return a list of Appointment by filters */
func listAppointment(db service.DB, doctID *uuid.UUID, f m.FilterAppointment) ([]m.Appointment, error) {
	sch := []m.Appointment{}
	query := psql.Select("app.appo_id", "app.start_at", "app.sche_id", "app.pati_id", "pat.name as pati_name", "app.type", "app.status", "app.created_at").
		From("appointment app").
		LeftJoin("schedule sch USING (sche_id)").
		LeftJoin("patient pat USING (pati_id)")

	if doctID != nil {
		query = query.Where(`sch.doct_id = ?`, doctID)
	}
	if f.AppoID != nil {
		query = query.Where(`appo_id = ?`, f.AppoID)
	}
	if f.StartAtGte != nil {
		query = query.Where(sq.GtOrEq{"app.start_at": f.StartAtGte})
	}
	if f.StartAtLte != nil {
		query = query.Where(sq.LtOrEq{"app.start_at": f.StartAtLte})
	}
	if f.ScheID != nil {
		query = query.Where(`sche_id = ?`, f.ScheID)
	}
	if f.PatiID != nil {
		query = query.Where(`pati_id = ?`, f.PatiID)
	}
	if f.Type != nil {
		query = query.Where(`type ILIKE ?`, `%`+*f.Type+`%`)
	}
	if f.Status != nil {
		query = query.Where(`status ILIKE ?`, `%`+*f.Status+`%`)
	}
	orderQuery, err := buildOrderBy("app.", f.FieldOrder, f.TypeOrder)
	if err != nil {
		return nil, err
	}
	if len(orderQuery) > 0 {
		query = query.OrderBy(orderQuery)
	}
	if f.InitialDate != nil {
		query = query.Where(sq.GtOrEq{"created_at": f.InitialDate})
	}
	if f.FinishDate != nil {
		query = query.Where(sq.LtOrEq{"created_at": f.FinishDate})
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list of Appointments sql")
	}
	err = db.Select(&sch, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list of Appointments sql")
		}
		return nil, nil
	}
	return sch, nil
}

/* Return a Appointment by appo_id */
func getAppointment(db service.DB, doctID *uuid.UUID, appoID uuid.UUID) (*m.Appointment, error) {
	sch := m.Appointment{}
	query := psql.Select("app.appo_id", "app.start_at", "app.sche_id", "app.pati_id", "pat.name as pati_name", "app.type", "app.status", "app.created_at").
		From("appointment app").
		LeftJoin("schedule sch USING (sche_id)").
		LeftJoin("patient pat USING (pati_id)").
		Where(sq.Eq{"appo_id": appoID})

	if doctID != nil {
		query = query.Where(`sch.doct_id = ?`, doctID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Appointment sql")
	}
	err = db.Get(&sch, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Appointment sql")
		}
		return nil, nil
	}
	return &sch, nil
}

/* Update Appointment to database by appo_id */
func updateAppointment(db service.DB, app *m.Appointment, doctID *uuid.UUID) (*m.Appointment, error) {
	query := psql.Update("appointment").
		Set("start_at", app.StartAt).
		Set("sche_id", app.ScheID).
		Set("pati_id", app.PatiID).
		Set("type", app.Type).
		Set("status", app.Status).
		Suffix("RETURNING *").
		Where(sq.Eq{"appo_id": app.AppoID})

	if doctID != nil {
		query = query.Where(`EXISTS (SELECT sche_id FROM schedule WHERE doct_id = ? AND sche_id = ?)`, doctID, app.ScheID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Appointment update sql")
	}

	err = db.Get(app, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Appointment update sql")
	}

	return app, nil
}

/* Delete Appointment to database by appo_id */
func deleteAppointment(db service.DB, appoID uuid.UUID) (*m.Appointment, error) {
	app := m.Appointment{}
	query := psql.Delete("appointment").
		Suffix("RETURNING *").
		Where(sq.Eq{"appo_id": appoID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &app, errors.Wrap(err, "Error generating delete Appointment sql")
	}
	err = db.Get(&app, qSQL, args...)
	if err != nil {
		return &app, errors.Wrap(err, "Error delete Appointment sql")
	}
	return &app, nil
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
