package user

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"gitlab.com/falqon/inovantapp/backend/service"
	"gitlab.com/falqon/inovantapp/backend/service/mailer"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"

	sq "github.com/Masterminds/squirrel"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Lister service to return users
type Lister struct {
	DB *sqlx.DB
}

// PushTokenSetter service to set a user's push token
type PushTokenSetter struct {
	DB *sqlx.DB
}

// PushTokens is the column type for the table "User", a pq.StringArray
type PushTokens struct {
	pq.StringArray
}

// Creator service to create new user
type Creator struct {
	DB     *sqlx.DB
	Mailer *mailer.Mailer
	Config *service.ServicesConfig
}

// Getter service to return user
type Getter struct {
	DB *sqlx.DB
}

// Updater service to update user
type Updater struct {
	DB *sqlx.DB
}

// Inactiver service to soft inative user
type Inactiver struct {
	DB *sqlx.DB
}

// Activer service to soft active user
type Activer struct {
	DB *sqlx.DB
}

type UserCreateModel struct {
	UserID    uuid.UUID `json:"userID"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"createdAt"`
}

// Run return user from id
func (g *Lister) Run() ([]m.User, error) {
	tx, err := g.DB.Beginx()
	if err != nil {
		return nil, err
	}
	u, err := listAll(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return u, err
}

// Run create new user
func (u *Creator) Run(i *m.User, password string) (*m.User, error) {
	tx, err := u.DB.Beginx()
	if err != nil {
		return nil, err
	}
	user, err := newUser(i, password)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	user, err = saveUser(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ac, ver, err := newActConfirmation(user.UserID, vPwd)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrap(err, "Failed to create action confirmation")
	// }

	// err = confirmationSave(tx, ac)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrap(err, "Failed to insert action confirmation")
	// }

	// cac := m.CreateAccountConfirm{
	// 	Email:           user.Email,
	// 	City:            "Salvador - Bahia",
	// 	ConfirmationURL: u.Config.APPURL + "/password_reset/" + ac.AcveID.String() + "/" + ver,
	// }
	// err = u.Mailer.SendConfirmationAccount(&cac)
	// if err != nil {
	// 	return user, errors.Wrap(err, "Failed to send confirmation Email")
	// }

	err = tx.Commit()
	return user, err
}

// Run return user from id
func (g *Getter) Run(userID uuid.UUID) (*m.User, error) {
	tx, err := g.DB.Beginx()
	if err != nil {
		return nil, err
	}
	u, err := fromID(tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return u, err
}

// Run update user data
func (g *Updater) Run(user *m.User) (*m.User, error) {
	tx, err := g.DB.Beginx()
	if err != nil {
		return nil, err
	}
	u, err := updateUser(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return u, err
}

// Run inactive user from user_id
func (g *Inactiver) Run(userID uuid.UUID) (*m.User, error) {
	tx, err := g.DB.Beginx()
	if err != nil {
		return nil, err
	}
	u, err := inactiveUser(tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return u, err
}

// Run active user from user_id
func (g *Activer) Run(userID uuid.UUID) (*m.User, error) {
	tx, err := g.DB.Beginx()
	if err != nil {
		return nil, err
	}
	u, err := activeUser(tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return u, err
}

// Run create new user
func (c *PushTokenSetter) Run(userID uuid.UUID, token string) error {
	query := psql.Update(`"user"`).
		Set("push_tokens", sq.Expr(`Array[?]`, token)).
		Where(sq.Eq{"user_id": userID}).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "Error generating user sql")
	}

	_, err = c.DB.Exec(qSQL, args...)
	if err != nil {
		return errors.Wrap(err, "Error inserting user")
	}
	return nil
}

// Create a new user
func newUser(u *m.User, password string) (*m.User, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating user uuid")
	}

	passHash, err := auth.PasswordGen(password)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to hash user password")
	}
	u.UserID = id
	u.Password = passHash
	return u, nil
}

// Save a user in the database
func saveUser(tx service.DB, u *m.User) (*m.User, error) {
	newUser := m.User{}
	query := psql.Insert(`"user"`).
		Columns("user_id", "email", "password", "roles", "push_tokens").
		Values(u.UserID, u.Email, u.Password, u.Roles, u.Token).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating User sql")
	}

	err = tx.Get(&newUser, qSQL, args...)
	pqError, ok := err.(*pq.Error)
	if ok && pqError.Code == "23505" {
		return nil, errors.New("User with this Email already registered")
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting User")
	}
	return &newUser, nil
}

// Update updates a user in the database
func updateUser(tx *sqlx.Tx, u *m.User) (*m.User, error) {
	query := psql.Update(`"user"`).
		Set("email", u.Email).
		Set("roles", u.Roles).
		Set("inactive_at", u.InactiveAt).
		Suffix("RETURNING *")

	query = query.Where(sq.Eq{"user_id": u.UserID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating user update sql")
	}

	err = tx.Get(u, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error user update sql")
	}

	return u, nil
}

// listAll returns users
func listAll(tx *sqlx.Tx) ([]m.User, error) {
	u := []m.User{}
	query := psql.Select("u.user_id", "u.email", "u.password", "u.roles", "u.created_at", "u.inactive_at", "u.push_tokens").
		From(`"user" u`).
		LeftJoin(`doctor doc USING (user_id)`).
		Where(sq.Eq{"doc.user_id": nil})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list Users sql")
	}
	err = tx.Select(&u, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return u, nil
		}
		return nil, errors.Wrap(err, "Error listing Users sql")
	}
	return u, nil
}

// fromID returns an user from user_id
func fromID(tx *sqlx.Tx, userID uuid.UUID) (*m.User, error) {
	u := m.User{}
	query := psql.Select("user_id", "email", "password", "roles", "created_at", "inactive_at", "push_tokens").
		From(`"user"`).
		Where(sq.Eq{"user_id": userID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get User sql")
	}
	err = tx.Get(&u, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error get User sql")
	}
	return &u, nil
}

// inactiveUser soft inactive user from user_id
func inactiveUser(tx service.DB, userID uuid.UUID) (*m.User, error) {
	u := m.User{}
	query := psql.Update(`"user"`).
		Set("inactive_at", time.Now()).
		Where(sq.Eq{"user_id": userID}).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = tx.Get(&u, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// activeUser soft active user from user_id
func activeUser(tx service.DB, userID uuid.UUID) (*m.User, error) {
	u := m.User{}
	query := psql.Update(`"user"`).
		Set("inactive_at", nil).
		Where(sq.Eq{"user_id": userID}).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = tx.Get(&u, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// fromEmail return User from email
func fromEmail(db *sqlx.DB, email string) (usr *m.UserWithDoctor, err error) {
	usr = &m.UserWithDoctor{}
	query := psql.Select("u.user_id", "doc.doct_id", "doc.name as doct_name", "u.email", "u.password", "u.created_at", "u.inactive_at").
		From(`"user" u`).
		LeftJoin("doctor doc USING (user_id)").
		Where(sq.Eq{"u.inactive_at": nil}).
		Where("u.email ILIKE ?", strings.TrimSpace(email))
	qSQL, args, err := query.ToSql()
	if err != nil {
		return usr, err
	}
	err = db.Get(usr, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return usr, &auth.UserNotFoundError{
				Message: "No User with this email: " + email,
			}
		}
		return usr, err
	}
	return usr, err
}
