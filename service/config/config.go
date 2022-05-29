package config

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Creator service to create new Config
type Creator struct {
	DB *sqlx.DB
}

//Run create new Config
func (c *Creator) Run(con *m.Config) (*m.Config, error) {
	u, err := createConfig(c.DB, con)
	return u, err
}

//Lister service to return Config
type Lister struct {
	DB *sqlx.DB
}

//Run return a list of Config by Filter
func (l *Lister) Run(f m.FilterConfig) ([]m.Config, error) {
	u, err := listConfig(l.DB, f)
	return u, err
}

//Getter service to return Config
type Getter struct {
	DB *sqlx.DB
}

//Run return a Config by key
func (g *Getter) Run(key string) (*m.Config, error) {
	u, err := getConfig(g.DB, key)
	return u, err
}

//Updater service to update Config
type Updater struct {
	DB *sqlx.DB
}

//Run update Config data
func (g *Updater) Run(con *m.Config) (*m.Config, error) {
	u, err := updateConfig(g.DB, con)
	return u, err
}

//Deleter service to soft delete Config
type Deleter struct {
	DB *sqlx.DB
}

//Run soft delete Config by key
func (d *Deleter) Run(key string) (*m.Config, error) {
	u, err := deleteConfig(d.DB, key)
	return u, err
}

/* Create a new Config to database */
func createConfig(db service.DB, con *m.Config) (*m.Config, error) {
	query := psql.Insert("config").
		Columns("key", "value").
		Values(con.Key, con.Value).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Config sql")
	}

	err = db.Get(con, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting Config in database")
	}
	return con, nil
}

/* Return a list of Config by filters */
func listConfig(db service.DB, f m.FilterConfig) ([]m.Config, error) {
	con := []m.Config{}
	query := psql.Select("key", "value").
		From("config")

	if f.Key != nil {
		query = query.Where(`key = ?`, f.Key)
	}
	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		query = query.Offset(uint64(*f.Offset))
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating list of Configs sql")
	}
	err = db.Select(&con, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error list of Configs sql")
		}
		return nil, nil
	}
	return con, nil
}

/* Return a Config by key */
func getConfig(db service.DB, key string) (*m.Config, error) {
	con := m.Config{}
	query := psql.Select("key", "value").
		From("config").
		Where(sq.Eq{"key": key})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating get Config sql")
	}
	err = db.Get(&con, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "Error get Config sql")
		}
		return nil, nil
	}
	return &con, nil
}

/* Update Config to database by key */
func updateConfig(db service.DB, con *m.Config) (*m.Config, error) {
	query := psql.Update("config").
		Set("label", con.Value).
		Suffix("RETURNING *").
		Where(sq.Eq{"key": con.Key})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating Config update sql")
	}

	err = db.Get(con, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error Config update sql")
	}

	return con, nil
}

/* Delete Config to database by key */
func deleteConfig(db service.DB, key string) (*m.Config, error) {
	con := m.Config{}
	query := psql.Delete("config").
		Suffix("RETURNING *").
		Where(sq.Eq{"key": key})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return &con, errors.Wrap(err, "Error generating delete Config sql")
	}
	err = db.Get(&con, qSQL, args...)
	if err != nil {
		return &con, errors.Wrap(err, "Error delete Config sql")
	}
	return &con, nil
}
