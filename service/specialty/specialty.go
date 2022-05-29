package specialty

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Creator service to create new Specialty
type Creator struct {
	DB *sqlx.DB
}

//Run create new Specialty
func (c *Creator) Run(spe *m.Specialty) (*m.Specialty, error) {
	u, err := createSpecialty(c.DB, spe)
	return u, err
}

//Lister service to return Specialty
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Specialty by Filter
func (l *Lister) Run(f m.FilterSpecialty) ([]m.Specialty, error) {
	u, err := listSpecialty(l.DB, f)
	return u, err
}

//Getter service to return Specialty
type Getter struct {
	DB *sqlx.DB
}

//Run return a Specialty by spec_id
func (g *Getter) Run(specID int64) (*m.Specialty, error) {
	u, err := getSpecialty(g.DB, specID)
	return u, err
}

//Updater service to update Specialty
type Updater struct {
	DB *sqlx.DB
}

//Run update Specialty data
func (g *Updater) Run(spe *m.Specialty) (*m.Specialty, error) {
	u, err := updateSpecialty(g.DB, spe)
	return u, err
}

//Deleter service to soft delete Specialty
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete Specialty by spec_id
func (d *Deleter) Run(specID int64) (*m.Specialty, error) {
	u, err := deleteSpecialty(d.DB, specID)
	return u, err
}

/* Create a new Specialty to database */
func createSpecialty(db service.DB, spe *m.Specialty) (*m.Specialty, error) {
	query := psql.Insert("specialty").
		Columns("name", "description", "info").
		Values(spe.Name, spe.Description, spe.Info).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Specialty sql")
	}

	err = db.Get(spe, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting Specialty in database")
	}
	return spe, nil
}

/* Return a list of Specialty by filters */
func listSpecialty(db service.DB, f m.FilterSpecialty) ([]m.Specialty, error) {
	spe := []m.Specialty{}
	query := psql.Select("spec_id", "name", "description").
		From("specialty")

	if f.Name != nil {
		query = query.Where(`name ILIKE ?`, `%`+*f.Name+`%`)
	}
	if f.Description != nil {
		query = query.Where(`description ILIKE ?`, `%`+*f.Description+`%`)
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list of Specialty sql")
	}
	err = db.Select(&spe, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list of Specialty sql")
		}
		return nil, nil
	}
	return spe, nil
}

/* Return a Specialty by spec_id */
func getSpecialty(db service.DB, specID int64) (*m.Specialty, error) {
	spe := m.Specialty{}
	query := psql.Select("spec_id", "name", "description", "info").
		From("specialty").
		Where(sq.Eq{"spec_id": specID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Specialty sql")
	}
	err = db.Get(&spe, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Specialty sql")
		}
		return nil, nil
	}
	return &spe, nil
}

/* Update Specialty to database by spec_id */
func updateSpecialty(db service.DB, spe *m.Specialty) (*m.Specialty, error) {
	query := psql.Update("specialty").
		Set("name", spe.Name).
		Set("description", spe.Description).
		Set("info", spe.Info).
		Suffix("RETURNING *").
		Where(sq.Eq{"spec_id": spe.SpecID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Specialty update sql")
	}

	err = db.Get(spe, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Specialty update sql")
	}

	return spe, nil
}

/* Delete Specialty to database by spec_id */
func deleteSpecialty(db service.DB, specID int64) (*m.Specialty, error) {
	spe := m.Specialty{}
	query := psql.Delete("specialty").
		Suffix("RETURNING *").
		Where(sq.Eq{"spec_id": specID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &spe, errors.Wrap(err, "Error generating delete Specialty sql")
	}
	err = db.Get(&spe, qSQL, args...)
	if err != nil {
		return &spe, errors.Wrap(err, "Error delete Specialty sql")
	}
	return &spe, nil
}
