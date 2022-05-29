package doctorspecialty

import (
	"database/sql"
	"log"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Creator service to create new DoctorSpecialty
type Creator struct {
	DB     *sqlx.DB
	Logger *log.Logger
}

//Run create new DoctorSpecialty
func (c *Creator) Run(doc *m.DoctorSpecialty) (*m.DoctorSpecialty, error) {
	u, err := createDoctorSpecialty(c.DB, doc)
	return u, err
}

//Lister service to return DoctorSpecialty
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Production Orders by Filter
func (l *Lister) Run(f m.FilterDoctorSpecialty) ([]m.DoctorSpecialty, error) {
	u, err := listDoctorSpecialty(l.DB, f)
	return u, err
}

//Getter service to return DoctorSpecialty
type Getter struct {
	DB *sqlx.DB
}

//Run return a doctorSpecialty of Production Orders by doct_id
func (g *Getter) Run(doctID uuid.UUID, specID int64) (*m.DoctorSpecialty, error) {
	u, err := getDoctorSpecialty(g.DB, doctID, specID)
	return u, err
}

//Updater service to update DoctorSpecialty
type Updater struct {
	DB *sqlx.DB
}

//Run update DoctorSpecialty data
func (g *Updater) Run(po *m.DoctorSpecialty) (*m.DoctorSpecialty, error) {
	u, err := updateDoctorSpecialty(g.DB, po)
	return u, err
}

//Deleter service to soft delete DoctorSpecialty
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete DoctorSpecialty by doct_id
func (d *Deleter) Run(doctID uuid.UUID, specID int64) (*m.DoctorSpecialty, error) {
	u, err := deleteDoctorSpecialty(d.DB, doctID, specID)
	return u, err
}

/* Create a new DoctorSpecialty to database */
func createDoctorSpecialty(db service.DB, doc *m.DoctorSpecialty) (*m.DoctorSpecialty, error) {
	query := psql.Insert("doctor_specialty").
		Columns("doct_id", "spec_id").
		Values(doc.DoctID, doc.SpecID).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating DoctorSpecialty sql")
	}

	err = db.Get(doc, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting DoctorSpecialty in database")
	}
	return doc, nil
}

/* Return a list of DoctorSpecialty by filters */
func listDoctorSpecialty(db service.DB, f m.FilterDoctorSpecialty) ([]m.DoctorSpecialty, error) {
	doc := []m.DoctorSpecialty{}
	query := psql.Select("doct_id", "spec_id", "created_at").
		From("doctor_specialty")

	if f.DoctID != nil {
		query = query.Where(`doct_id = ?`, f.DoctID)
	}
	if f.SpecID != nil {
		query = query.Where(`spec_id = ?`, f.SpecID)
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list DoctorSpecialty sql")
	}
	err = db.Select(&doc, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list DoctorSpecialty sql")
		}
		return nil, nil
	}
	return doc, nil
}

/* Return a DoctorSpecialty by doct_id */
func getDoctorSpecialty(db service.DB, doctID uuid.UUID, specID int64) (*m.DoctorSpecialty, error) {
	doc := m.DoctorSpecialty{}
	query := psql.Select("doct_id", "spec_id", "created_at").
		From("doctor_specialty").
		Where(sq.Eq{"doct_id": doctID, "spec_id": specID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get DoctorSpecialty sql")
	}
	err = db.Get(&doc, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get DoctorSpecialty sql")
		}
		return nil, nil
	}
	return &doc, nil
}

/* Update DoctorSpecialty to database by doct_id */
func updateDoctorSpecialty(db service.DB, doc *m.DoctorSpecialty) (*m.DoctorSpecialty, error) {
	query := psql.Update("doctor_specialty").
		Set("doct_id", doc.DoctID).
		Set("spec_id", doc.SpecID).
		Suffix("RETURNING *").
		Where(sq.Eq{"doct_id": doc.DoctID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating DoctorSpecialty update sql")
	}

	err = db.Get(doc, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error DoctorSpecialty update sql")
	}

	return doc, nil
}

/* Delete DoctorSpecialty to database by doct_id */
func deleteDoctorSpecialty(db service.DB, doctID uuid.UUID, specID int64) (*m.DoctorSpecialty, error) {
	spe := m.DoctorSpecialty{}
	query := psql.Delete("doctor_specialty").
		Suffix("RETURNING *").
		Where(sq.Eq{"doct_id": doctID, "spec_id": specID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &spe, errors.Wrap(err, "Error generating delete DoctorSpecialty sql")
	}
	err = db.Get(&spe, qSQL, args...)
	if err != nil {
		return &spe, errors.Wrap(err, "Error delete DoctorSpecialty sql")
	}
	return &spe, nil
}
