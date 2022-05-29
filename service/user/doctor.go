package user

import (
	"database/sql"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"
	"gitlab.com/falqon/inovantapp/backend/service/mailer"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psqlx = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//DoctorCreator service to create new Doctor
type DoctorCreator struct {
	DB     *sqlx.DB
	Logger *log.Logger
	Mailer *mailer.Mailer
	Config *service.ServicesConfig
}

//Run create new Doctor
func (c *DoctorCreator) Run(doc *m.Doctor) (*m.Doctor, error) {
	tx, err := c.DB.Beginx()
	user, err := newUser(&doc.User, string(doc.User.Password))
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println("Doctor Create New User:", err)
		}
	}
	userSaved, err := saveUser(tx, user)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println("Doctor Create Save User:", err)
		}
		return nil, err
	}

	ac, ver, err := newActConfirmation(userSaved.UserID, vPwd)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "Failed to create action confirmation")
	}

	err = confirmationSave(tx, ac)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "Failed to insert action confirmation")
	}

	doc.User = *userSaved
	doc.UserID = user.UserID
	doctID, err := uuid.NewV4()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	doc.DoctID = doctID
	u, err := createDoctor(tx, doc)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println("Doctor Create:", err)
		}
	}
	err = addDoctorSpecialties(tx, u.DoctID, doc.Specialties)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Println("Doctor Create Specialties:", err)
		}
	}

	cac := m.CreateAccountConfirm{
		Name:            &u.Name,
		Email:           u.Email,
		City:            "Salvador - Bahia",
		ConfirmationURL: c.Config.APPURL + "/password_reset/" + ac.AcveID.String() + "/" + ver,
	}
	err = c.Mailer.SendConfirmationAccount(&cac)
	if err != nil {
		return u, errors.Wrap(err, "Failed to send confirmation Email")
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return u, err
}

//DoctorLister service to return Doctor
type DoctorLister struct {
	DB *sqlx.DB
}

//Run return a list of Production Orders by Filter
func (l *DoctorLister) Run(doctID *uuid.UUID, f m.FilterDoctor) ([]m.Doctor, error) {
	u, err := listDoctor(l.DB, doctID, f)
	return u, err
}

//DoctorGetter service to return Doctor
type DoctorGetter struct {
	DB *sqlx.DB
}

//Run return a doctor of Production Orders by doct_id
func (g *DoctorGetter) Run(doctID uuid.UUID) (*m.Doctor, error) {
	u, err := getDoctor(g.DB, doctID)
	return u, err
}

//DoctorUpdater service to update Doctor
type DoctorUpdater struct {
	DB     *sqlx.DB
	Logger *log.Logger
}

//Run update Doctor data
func (g *DoctorUpdater) Run(doc *m.Doctor) (*m.Doctor, error) {
	tx, err := g.DB.Beginx()
	u, err := updateDoctor(g.DB, doc)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Doctor Update:", err)
		}
	}
	user := m.User{
		UserID:     doc.UserID,
		Email:      doc.User.Email,
		Roles:      doc.User.Roles,
		InactiveAt: doc.InactiveAt,
	}
	_, err = updateUser(tx, &user)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Doctor Update User:", err)
		}
	}

	err = delDoctorSpecialties(tx, u.DoctID, doc.Specialties)
	if err != nil {
		if g.Logger != nil {
			g.Logger.Println("Doctor Update Specialties:", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return u, err
}

//DoctorDeleter service to soft delete Doctor
type DoctorDeleter struct {
	DB *sqlx.DB
}

//Run soft delete Doctor by doct_id
func (d *DoctorDeleter) Run(doctID uuid.UUID) (*m.Doctor, error) {
	tx, err := d.DB.Beginx()
	u, err := deleteDoctor(tx, doctID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = cancelAppoitments(tx, doctID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = cancelSchedules(tx, doctID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return u, err
}

/* Create a new Doctor to database */
func createDoctor(db service.DB, doc *m.Doctor) (*m.Doctor, error) {
	query := psql.Insert("doctor").
		Columns("doct_id", "user_id", "name", "info").
		Values(doc.DoctID, doc.UserID, doc.Name, doc.Info).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = db.Get(doc, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func addDoctorSpecialties(db service.DB, doctID uuid.UUID, specialties *pq.StringArray) error {
	docSpe := m.DoctorSpecialty{}
	args := []interface{}{specialties, doctID}
	query := `
		WITH insert_doc_spe AS (
			SELECT UNNEST($1::int[]) AS spec_id
		)
		INSERT INTO doctor_specialty (doct_id, spec_id)
		SELECT $2::UUID, spec_id::INT FROM insert_doc_spe
	`
	err := db.Get(&docSpe, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func delDoctorSpecialties(db service.DB, doctID uuid.UUID, specialties *pq.StringArray) error {
	docSpe := m.DoctorSpecialty{}
	args := []interface{}{specialties, doctID}
	query := `
		WITH del_doctor_specialty AS (
			DELETE FROM doctor_specialty
			WHERE doct_id = $2
		),
		insert_doc_spe AS (
			SELECT UNNEST($1::INT[]) AS spec_id
		)
		INSERT INTO doctor_specialty (doct_id, spec_id)
		SELECT $2::UUID, spec_id::INT FROM insert_doc_spe
	`
	err := db.Get(&docSpe, query, args...)
	if err != nil {
		return err
	}
	return nil
}

/* Return a list of Doctor by filters */
func listDoctor(db service.DB, doctID *uuid.UUID, f m.FilterDoctor) ([]m.Doctor, error) {
	doc := []m.Doctor{}
	query := psql.Select("doc.doct_id", "doc.user_id", "doc.name", "u.email", "u.roles", "doc.info", "doc.created_at", "u.inactive_at", "array_remove(array_agg(spe.spec_id), NULL) AS specialties").
		From("doctor doc").
		Join(`"user" u USING (user_id)`).
		LeftJoin("doctor_specialty doc_spe USING (doct_id)").
		LeftJoin("specialty spe USING (spec_id)").
		GroupBy("doc.doct_id, doc.user_id, doc.name, u.email, u.roles, doc.info, doc.created_at, u.inactive_at").
		OrderBy("doc.name ASC")

	if doctID != nil {
		query = query.Where(`doc.doct_id = ?`, doctID)
	}
	if f.DoctID != nil {
		query = query.Where(`doc.doct_id = ?`, f.DoctID)
	}
	if f.UserID != nil {
		query = query.Where(`doc.user_id = ?`, f.UserID)
	}
	if f.Name != nil {
		query = query.Where(`doc.name ILIKE ?`, `%`+*f.Name+`%`)
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = db.Select(&doc, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	return doc, nil
}

/* Return a Doctor by doct_id */
func getDoctor(db service.DB, doctID uuid.UUID) (*m.Doctor, error) {
	doc := m.Doctor{}
	query := psql.Select("doc.doct_id", "doc.user_id", "doc.name", "u.email", "u.roles", "doc.info", "doc.created_at", "u.inactive_at", "array_remove(array_agg(spe.spec_id), NULL) AS specialties").
		From("doctor doc").
		Join(`"user" u USING (user_id)`).
		LeftJoin("doctor_specialty doc_spe USING (doct_id)").
		LeftJoin("specialty spe USING (spec_id)").
		Where(sq.Eq{"doc.doct_id": doctID}).
		GroupBy("doc.doct_id, doc.user_id, doc.name, u.email, u.roles, doc.info, doc.created_at, u.inactive_at")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = db.Get(&doc, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	return &doc, nil
}

/* Update Doctor to database by doct_id */
func updateDoctor(db service.DB, doc *m.Doctor) (*m.Doctor, error) {
	query := psql.Update("doctor").
		Set("name", doc.Name).
		Set("info", doc.Info).
		Set("update_at", time.Now()).
		Suffix("RETURNING *").
		Where(sq.Eq{"doct_id": doc.DoctID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = db.Get(doc, qSQL, args...)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

/* Delete Doctor to database by doct_id */
func deleteDoctor(db service.DB, doctID uuid.UUID) (*m.Doctor, error) {
	doc := m.Doctor{}
	query := psqlx.Update(`"user"`).
		Set(`inactive_at`, time.Now()).
		From("doctor doc").
		Where(`"user".user_id = doc.user_id`).
		Suffix("RETURNING *").
		Where(sq.Eq{"doc.doct_id": doctID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &doc, err
	}
	err = db.Get(&doc, qSQL, args...)
	if err != nil {
		return &doc, err
	}
	return &doc, nil
}

func cancelAppoitments(db service.DB, doctID uuid.UUID) error {
	query := psqlx.Update("appointment").
		Set("status", "cancelled").
		From("schedule").
		Where("schedule.sche_id = appointment.sche_id").
		Suffix("RETURNING *").
		Where(sq.Eq{"schedule.doct_id": doctID}).
		Where(sq.GtOrEq{"appointment.start_at": time.Now()})
	qSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(qSQL, args...)
	if err != nil {
		return err
	}
	return nil
}

func cancelSchedules(db service.DB, doctID uuid.UUID) error {
	query := psqlx.Update("schedule").
		Set("deleted_at", time.Now()).
		Suffix("RETURNING *").
		Where(sq.Eq{"doct_id": doctID}).
		Where(sq.GtOrEq{"start_at": time.Now()})
	qSQL, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(qSQL, args...)
	if err != nil {
		return err
	}
	return nil
}
