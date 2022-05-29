package actionverification

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

//Creator service to create new ActionVerification
type Creator struct {
	DB *sqlx.DB
}

//Run create new ActionVerification
func (c *Creator) Run(acv *m.ActionVerification) (*m.ActionVerification, error) {
	acveID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating ActionVerification uuid")
	}
	acv.AcveID = acveID
	u, err := createActionVerification(c.DB, acv)
	return u, err
}

//Lister service to return ActionVerification
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of ActionVerification by Filter
func (l *Lister) Run(f m.FilterActionVerification) ([]m.ActionVerification, error) {
	u, err := listActionVerification(l.DB, f)
	return u, err
}

//Getter service to return ActionVerification
type Getter struct {
	DB *sqlx.DB
}

//Run return a ActionVerification by sche_id
func (g *Getter) Run(acveID uuid.UUID) (*m.ActionVerification, error) {
	u, err := getActionVerification(g.DB, acveID)
	return u, err
}

//Updater service to update ActionVerification
type Updater struct {
	DB *sqlx.DB
}

//Run update ActionVerification data
func (g *Updater) Run(app *m.ActionVerification) (*m.ActionVerification, error) {
	u, err := updateActionVerification(g.DB, app)
	return u, err
}

//Deleter service to soft delete ActionVerification
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete ActionVerification by acve_id
func (d *Deleter) Run(acveID uuid.UUID) (*m.ActionVerification, error) {
	u, err := deleteActionVerification(d.DB, acveID)
	return u, err
}

/* Create a new ActionVerification to database */
func createActionVerification(db service.DB, acv *m.ActionVerification) (*m.ActionVerification, error) {
	query := psql.Insert("action_verification").
		Columns("acve_id", "user_id", "type", "verification").
		Values(acv.AcveID, acv.UserID, acv.Type, acv.Verification).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Action Verification sql")
	}

	err = db.Get(acv, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting Action Verification in database")
	}
	return acv, nil
}

/* Return a list of ActionVerification by filters */
func listActionVerification(db service.DB, f m.FilterActionVerification) ([]m.ActionVerification, error) {
	pat := []m.ActionVerification{}
	query := psql.Select("acve_id", "user_id", "type", "verification", "created_at", "deleted_at").
		From("action_verification").
		Where("deleted_at IS NULL")

	if f.AcveID != nil {
		query = query.Where(`acve_id = ?`, f.AcveID)
	}
	if f.UserID != nil {
		query = query.Where(`user_id = ?`, f.UserID)
	}
	if f.Type != nil {
		query = query.Where(`type ILIKE ?`, `%`+*f.Type+`%`)
	}
	if f.Verification != nil {
		query = query.Where(`verification ILIKE ?`, `%`+*f.Verification+`%`)
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
		return nil, errors.Wrap(err, "Error generating list of Actions Verification sql")
	}
	err = db.Select(&pat, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list of Actions Verification sql")
		}
		return nil, nil
	}
	return pat, nil
}

/* Return a ActionVerification by acve_id */
func getActionVerification(db service.DB, acveID uuid.UUID) (*m.ActionVerification, error) {
	acv := m.ActionVerification{}
	query := psql.Select("acve_id", "user_id", "type", "verification", "created_at", "deleted_at").
		From("action_verification").
		Where(sq.Eq{"acve_id": acveID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Action Verification sql")
	}
	err = db.Get(&acv, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Action Verification sql")
		}
		return nil, nil
	}
	return &acv, nil
}

/* Update ActionVerification to database by acve_id */
func updateActionVerification(db service.DB, acv *m.ActionVerification) (*m.ActionVerification, error) {
	query := psql.Update("action_verification").
		Set("user_id", acv.UserID).
		Set("type", acv.Type).
		Set("verification", acv.Verification).
		Suffix("RETURNING *").
		Where(sq.Eq{"acve_id": acv.AcveID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Action Verification update sql")
	}

	err = db.Get(acv, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Action Verification update sql")
	}

	return acv, nil
}

/* Delete ActionVerification to database by acve_id */
func deleteActionVerification(db service.DB, acveID uuid.UUID) (*m.ActionVerification, error) {
	acv := m.ActionVerification{}
	query := psql.Update(`action_verification`).
		Set("deleted_at", time.Now()).
		Suffix("RETURNING *").
		Where(sq.Eq{"acve_id": acveID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &acv, errors.Wrap(err, "Error generating delete Action Verification sql")
	}
	err = db.Get(&acv, qSQL, args...)
	if err != nil {
		return &acv, errors.Wrap(err, "Error delete Action Verification sql")
	}
	return &acv, nil
}
