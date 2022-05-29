package messaging

import (
	"database/sql"
	"github.com/lib/pq"

	sq "github.com/elgris/sqrl"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// MessageCreator service to create new message
type MessageCreator struct {
	DB *sqlx.DB
}

// MessageNotificator service to lister messages
type MessageNotificator struct {
	DB *sqlx.DB
}

// MessageLister service to lister messages
type MessageLister struct {
	DB *sqlx.DB
}

// MessageReadCreator service to create new message
type MessageReadCreator struct {
	DB *sqlx.DB
}

// ActivityLister service to list message activity
type ActivityLister struct {
	DB *sqlx.DB
}

// GetUsersForMessage service to get users for message
type GetUsersForMessage struct {
	DB *sqlx.DB
}

// Run inserts a new auto response into the database
func (u *MessageCreator) Run(t *m.Message) (*m.Message, error) {
	query := psql.Insert("message").
		Columns("pror_id", "type", "from_user_id", "value", "data").
		Values(t.ProrID, t.Type, t.FromUserID, t.TextValue, t.Data).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate Message insert sql")
	}

	err = u.DB.Get(t, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to insert Message")
	}

	return t, nil
}

// Run inserts a new auto response into the database
func (u *MessageReadCreator) Run(messID []int64, targets []string) ([]string, error) {
	tx, err := u.DB.Beginx()
	if err != nil {
		return nil, err
	}
	ts := []string{}

	for _, target := range targets {
		userID := uuid.FromStringOrNil(target)
		msgs := messID
		for _, id := range messID {
			msgs = append(msgs, id)
		}
		_, err := MessageReadInsert(tx, msgs, userID)
		if err == nil {
			ts = append(ts, target)
		}
	}
	if len(ts) == 0 {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return ts, nil

}

// Run return a list of messages
func (u *MessageLister) Run(f m.FilterMessage) ([]m.ChatMessage, error) {
	sel := []string{"fromu.name as from_user_name, coalesce(rb.readers, '[]') as read_by"}
	sel = append(sel, "message.created_at", "data", "from_user_id", "mess_id", "type", "value")
	query := psql.Select(sel...).
		Prefix(`with readers as (select mess_id, json_agg(user_id) as readers from message_read group by mess_id)`).
		From(m.Message{}.Name()).
		Join(`"user" fromu on fromu.user_id = from_user_id`).
		LeftJoin(`readers rb using(mess_id)`).
		Where(sq.Or{
			sq.Eq{"message.pror_id": f.GroupID},
		}).
		OrderBy("mess_id DESC")

	if f.Limit != nil {
		query = query.Limit(uint64(*f.Limit))
	}

	if f.BeforeID != nil {
		query = query.Where(sq.Lt{"mess_id": *f.BeforeID})
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate Message list sql")
	}
	t := []m.ChatMessage{}
	err = u.DB.Select(&t, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list Messages")
	}
	return t, nil
}

// Run returns a list of user message activity
func (u *ActivityLister) Run(userID string, forGroupID pq.StringArray) ([]m.ActivitySnapshot, error) {
	args := []interface{}{userID}
	sqlFilter := ""
	if forGroupID != nil {
		args = append(args, &forGroupID)
		sqlFilter = "WHERE pror_id = ANY($2)"
	}

	query := `
	WITH group_activity AS (
		SELECT pror_id, user_id FROM user_production_order
		WHERE user_id = $1
	),
	filtered_user_activity AS (
		SELECT mess_id, user_id, pror_id, "value", created_at
		FROM group_activity ga
		JOIN message m USING (pror_id)
		` + sqlFilter + `
	), qty_messages AS (
		SELECT mess_id, user_id, pror_id, created_at, "value", ROW_NUMBER () OVER (PARTITION BY pror_id ORDER BY created_at DESC) AS qty_messages
		FROM filtered_user_activity
	),
	last_activity AS (
		SELECT mess_id, user_id, pror_id, "value", created_at
		FROM qty_messages
		WHERE qty_messages < 2
	),
	read_message_status AS (
		SELECT uac.*, (CASE WHEN mr.user_id IS NULL THEN FALSE ELSE TRUE END) AS is_read
		FROM filtered_user_activity uac
		LEFT JOIN message_read mr USING (mess_id, user_id)
	),
	unread_messages AS (
		SELECT pror_id, COUNT (is_read) AS unread_message_count
		FROM read_message_status WHERE is_read = FALSE
		GROUP BY pror_id
	)

	SELECT la.pror_id as pror_id,  COALESCE(unread_message_count, 0) as unread_count, (row_to_json(M)->>'created_at')::timestamp as activity_at, row_to_json(M)->>'value' as last_message
	FROM last_activity la
	LEFT JOIN unread_messages USING (pror_id)
	JOIN message m USING (mess_id)
	ORDER BY m.created_at DESC
	`
	t := []m.ActivitySnapshot{}
	err := u.DB.Select(&t, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Failed to list message activity")
	}

	return t, nil
}

// Run returns a list of users that must be notified
func (u *MessageNotificator) Run(messID int64) ([]m.ToNotify, error) {
	query := psql.Select(
		"mm.from_user_id as user_id",
		"mm.mess_id",
		"mm.value as message_text",
		"mm.type as message_type",
		"uu.name as sender_name",
		"uu.user_id as sender_id",
		//"uu.push_tokens as tokens",
	).
		From("message mm").
		LeftJoin(`message_read mr ON mm.from_user_id = mr.user_id AND mm.mess_id = mr.mess_id`).
		Join(`"user" uu ON mm.from_user_id = uu.user_id`).
		Where(sq.Expr(`mr.mess_id IS NULL AND mm.mess_id = ?`, messID)).
		Where(`mm.created_at >= (now() - INTERVAL '1 h')::timestamp`)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate Message list sql")
	}
	t := []m.ToNotify{}
	err = u.DB.Select(&t, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list Messages")
	}

	return t, nil
}

//MessageReadInsert Run returns a list of users that must be notified
func MessageReadInsert(tx *sqlx.Tx, messID []int64, userID uuid.UUID) ([]m.MessageRead, error) {
	query := psql.Insert("message_read").
		Columns("mess_id", "user_id").
		Suffix("ON CONFLICT DO NOTHING RETURNING *")

	for _, id := range messID {
		query = query.Values(id, userID)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate Message Read insert sql")
	}

	res := []m.MessageRead{}
	err = tx.Select(&res, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to insert Message Read")
	}

	return res, nil
}

// Run inserts a new auto response into the database
func (u *GetUsersForMessage) Run(messID int64) ([]string, error) {
	res := struct {
		Ids pq.StringArray `db:"campo"`
	}{}
	query := psql.Select("array_agg(upo.user_id) as campo").
		From("message mess").
		LeftJoin("user_production_order upo USING (pror_id)").
		Where(sq.Eq{"mess_id": messID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get Users for Message sql")
	}

	err = u.DB.Get(&res, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get Users for Message")
	}

	return []string(res.Ids), nil
}
