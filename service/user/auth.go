package user

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"gitlab.com/falqon/inovantapp/backend/service"
	"gitlab.com/falqon/inovantapp/backend/service/mailer"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"

	m "gitlab.com/falqon/inovantapp/backend/models"

	sq "github.com/Masterminds/squirrel"
)

// AuthResponse Service Object Authentication
type AuthResponse struct {
	m.UserWithDoctor
	Jwt string
}

// Authenticator Object
type Authenticator struct {
	DB        *sqlx.DB
	JWTConfig JWTConfig
}

// JWTConfig Object
type JWTConfig struct {
	Secret          string
	HoursTillExpire time.Duration
	SigningMethod   *jwt.SigningMethodHMAC
}

//Run Authenticator User
func (u *Authenticator) Run(email, password string) (a *AuthResponse, err error) {

	usr, err := fromEmail(u.DB, strings.TrimSpace(email))
	if err != nil {
		return nil, err
	}
	jwt, err := authenticate(*usr, password, u.JWTConfig)
	if err != nil {
		return nil, err
	}
	a = &AuthResponse{UserWithDoctor: *usr, Jwt: jwt}

	return a, err
}

func authenticate(usr m.UserWithDoctor, password string, cfg JWTConfig) (jwttoken string, err error) {
	// err = bcrypt.CompareHashAndPassword(usr.Password, []byte(password))
	// if err != nil {
	// 	return jwttoken, &auth.ValidationError{
	// 		Messages: map[string]string{"password": "Wrong user/password combination"},
	// 	}
	// }
	doctID := ""
	if usr.DoctID != nil {
		doctID = usr.DoctID.String()
	}

	claims := auth.Claims{
		InstID: "",
		UserID: usr.UserID.String(),
		DoctID: doctID,
		Email:  usr.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(cfg.HoursTillExpire).UTC().Unix(),
		},
	}

	token := jwt.NewWithClaims(cfg.SigningMethod, claims)
	jwttoken, err = token.SignedString([]byte(cfg.Secret))

	return jwttoken, err
}

// PwdRecoverer is the service object to recover an User's password
type PwdRecoverer struct {
	DB     *sqlx.DB
	Mailer *mailer.Mailer
	Config *service.ServicesConfig
}

// Run starts an User`s password reset flow
func (p *PwdRecoverer) Run(email string) error {
	u, err := fromEmail(p.DB, email)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve user for email "+email)
	}

	ac, ver, err := newActConfirmation(u.UserID, vPwd)
	if err != nil {
		return errors.Wrap(err, "Failed to create action confirmation")
	}

	tx, err := p.DB.Beginx()
	if err != nil {
		return errors.Wrap(err, "Failed to begin transaction")
	}

	err = confirmationSave(tx, ac)
	if err != nil {
		return errors.Wrap(err, "Failed to insert action confirmation")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Failed to commit")
	}

	var name string
	if u.DoctName != nil {
		name = *u.DoctName
	}

	pr := m.PwdReset{
		Name:            name,
		ConfirmationURL: p.Config.APPURL + "/password_reset/" + ac.AcveID.String() + "/" + ver,
		Email:           u.Email,
	}
	err = p.Mailer.SendPwdResetRequest(pr)
	if err != nil {
		return errors.Wrap(err, "Failed to send")
	}
	return nil
}

// PwdReseter service resets an User`s password
type PwdReseter struct {
	DB     *sqlx.DB
	Mailer *mailer.Mailer
}

// Run resets an user's password
func (p *PwdReseter) Run(acveID, verification, password string) error {
	tx, err := p.DB.Beginx()
	if err != nil {
		return err
	}

	psrt, err := confirmationFromID(tx, acveID)
	if err != nil {
		return err
	}

	// validate verification
	err = bcrypt.CompareHashAndPassword([]byte(psrt.Verification), []byte(verification))
	if err != nil {
		return &auth.ValidationError{
			Messages: map[string]string{"verification": "Invalid verification id"},
		}
	}

	// update user password
	passHash, err := auth.PasswordGen(password)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = updatePassword(tx, psrt.UserID, passHash)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Failed to update user password")
	}

	// remove verification
	err = confirmationDelete(tx, &psrt)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Failed to update user password")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Failed to commit password reset")
	}

	// TODO: send user name, email
	err = p.Mailer.SendPwdResetAlert()
	if err != nil {
		return errors.Wrap(err, "Failed to insert action confirmation")
	}
	return nil
}

func updatePassword(tx *sqlx.Tx, userID uuid.UUID, pass []byte) error {
	query := psql.Update(`"user"`).
		Set("password", pass).
		Where(sq.Eq{"user_id": userID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "Error generating user password update sql")
	}

	_, err = tx.Exec(qSQL, args...)
	if err != nil {
		return errors.Wrap(err, "Error updating user password")
	}
	return nil
}
