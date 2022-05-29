package dashboard

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Viewer service to list Dashboard
type Viewer struct {
	DB *sqlx.DB
}

//Run Dashboard
func (v *Viewer) Run(fDashboard m.FilterDashboard) ([]m.Dashboard, error) {
	u, err := checkDashboard(v.DB, fDashboard)
	return u, err
}

/* Query to view dashboard */
func checkDashboard(db service.DB, f m.FilterDashboard) ([]m.Dashboard, error) {
	ava := []m.Dashboard{}
	args := []interface{}{f.DoctID, f.StartDate, f.EndDate}
	query :=
		`SELECT doct_id, start_at, end_at, plan
		FROM schedule
		WHERE start_at >= $2::TIMESTAMP
		AND end_at <= $3::TIMESTAMP
		AND doct_id = $1
		AND deleted_at IS NULL`
	err := db.Select(&ava, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list Dashboard sql")
		}
		return nil, nil
	}
	return ava, nil
}
