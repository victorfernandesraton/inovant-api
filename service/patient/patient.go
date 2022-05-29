package patient

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

//Creator service to create new Patient
type Creator struct {
	DB *sqlx.DB
}

//Run create new Patient
func (c *Creator) Run(pat *m.Patient) (*m.Patient, error) {
	patiID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Patient uuid")
	}
	pat.PatiID = patiID
	u, err := createPatient(c.DB, pat)
	return u, err
}

//Lister service to return Patient
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Patient by Filter
func (l *Lister) Run(doctID *uuid.UUID, f m.FilterPatient) ([]m.Patient, error) {
	u, err := listPatient(l.DB, doctID, f)
	return u, err
}

//Getter service to return Patient
type Getter struct {
	DB *sqlx.DB
}

//Run return a Patient by pati_id
func (g *Getter) Run(doctID *uuid.UUID, patiID uuid.UUID) (*m.Patient, error) {
	u, err := getPatient(g.DB, doctID, patiID)
	return u, err
}

//Updater service to update Patient
type Updater struct {
	DB *sqlx.DB
}

//Run update Patient data
func (g *Updater) Run(pat *m.Patient) (*m.Patient, error) {
	u, err := updatePatient(g.DB, pat)
	return u, err
}

//Deleter service to soft delete Patient
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete Patient by pati_id
func (d *Deleter) Run(doctID *uuid.UUID, patiID uuid.UUID) (*m.Patient, error) {
	u, err := deletePatient(d.DB, doctID, patiID)
	return u, err
}

/* Create a new Patient to database */
func createPatient(db service.DB, pat *m.Patient) (*m.Patient, error) {
	query := psql.Insert("patient").
		Columns("pati_id", "doct_id", "name", "email", "info").
		Values(pat.PatiID, pat.DoctID, pat.Name, pat.Email, pat.Info).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Patient sql")
	}

	err = db.Get(pat, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting Patient in database")
	}
	return pat, nil
}

/* Return a list of Patient by filters */
func listPatient(db service.DB, doctID *uuid.UUID, f m.FilterPatient) ([]m.Patient, error) {
	pat := []m.Patient{}
	query := psql.Select("pati_id", "doct_id", "pa.name as name", "email", "pa.info", "pa.created_at", "updated_at", "start_at AS last_appointment", "doc.name AS doct_name").
		From("patient_appointment pa").
		Join("doctor doc USING (doct_id)").
		Where(sq.Eq{"ord_number": 1}).
		OrderBy("name").
		Prefix(
			`WITH patient AS (
				SELECT pati_id, doct_id, "name", email, info, pat.created_at, updated_at, app.start_at,
				row_number() OVER(PARTITION BY pati_id ORDER BY app.start_at DESC) AS ord_number
				FROM patient pat
				LEFT JOIN appointment app USING (pati_id)
			),
			available_appointments AS (
				SELECT *
				FROM appointment
				WHERE start_at <= NOW() AND status = 'confirmed'
			),
			patient_appointment AS (
				SELECT pati_id, doct_id, "name", email, info, pat.created_at, updated_at, app.start_at, app.status,
				row_number() OVER(PARTITION BY pati_id ORDER BY app.start_at DESC) AS ord_number
				FROM patient pat
				LEFT JOIN available_appointments app USING (pati_id)
			)`,
		)

	if doctID != nil {
		query = query.Where(`pa.doct_id = ?`, doctID)
	}
	if f.PatiID != nil {
		query = query.Where(`pa.pati_id = ?`, f.PatiID)
	}
	if f.DoctID != nil {
		query = query.Where(`pa.doct_id = ?`, f.DoctID)
	}
	if f.Name != nil {
		query = query.Where(`pa.name ILIKE ?`, `%`+*f.Name+`%`)
	}
	if f.Email != nil {
		query = query.Where(`email ILIKE ?`, `%`+*f.Email+`%`)
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
		return nil, errors.Wrap(err, "Error generating list of Patients sql")
	}
	err = db.Select(&pat, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list of Patients sql")
		}
		return nil, nil
	}
	return pat, nil
}

/* Return a Patient by pati_id */
func getPatient(db service.DB, doctID *uuid.UUID, patiID uuid.UUID) (*m.Patient, error) {
	pat := m.Patient{}
	query := psql.Select("pati_id", "doct_id", "name", "email", "info", "created_at", "updated_at").
		From("patient").
		Where(sq.Eq{"pati_id": patiID})

	if doctID != nil {
		query = query.Where(`doct_id = ?`, doctID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Patient sql")
	}
	err = db.Get(&pat, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Patient sql")
		}
		return nil, nil
	}
	return &pat, nil
}

/* Update Patient to database by pati_id */
func updatePatient(db service.DB, pat *m.Patient) (*m.Patient, error) {
	query := psql.Update("patient").
		Set("doct_id", pat.DoctID).
		Set("name", pat.Name).
		Set("email", pat.Email).
		Set("info", pat.Info).
		Suffix("RETURNING *").
		Where(sq.Eq{"pati_id": pat.PatiID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Patient update sql")
	}

	err = db.Get(pat, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Patient update sql")
	}

	return pat, nil
}

/* Delete Patient to database by pati_id */
func deletePatient(db service.DB, doctID *uuid.UUID, patiID uuid.UUID) (*m.Patient, error) {
	pat := m.Patient{}
	query := psql.Delete("patient").
		Suffix("RETURNING *").
		Where(sq.Eq{"pati_id": patiID})
	if doctID != nil {
		query = query.Where(`doct_id = ?`, doctID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &pat, errors.Wrap(err, "Error generating delete Patient sql")
	}
	err = db.Get(&pat, qSQL, args...)
	if err != nil {
		return &pat, errors.Wrap(err, "Error delete Patient sql")
	}
	return &pat, nil
}
