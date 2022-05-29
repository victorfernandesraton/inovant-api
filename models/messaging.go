package models

import (
	"time"

	"github.com/gofrs/uuid"
	types "github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
)

// Tokens is a string array
type Tokens struct {
	pq.StringArray
}

// MessageRead model - holds message read status
type MessageRead struct {
	MessID int64     `db:"mess_id" json:"messID"`
	UserID uuid.UUID `db:"user_id" json:"userID"`
}

// Message model
type Message struct {
	MessID     int64          `db:"mess_id" json:"messID"`
	ProrID     string         `db:"pror_id" json:"prorID"`
	Type       string         `db:"type" json:"type"`
	FromUserID uuid.UUID      `db:"from_user_id" json:"fromUserID"`
	TextValue  string         `db:"value" json:"value"`
	Data       types.JSONText `db:"data" json:"data"`
	CreatedAt  time.Time      `db:"created_at" json:"createdAt"`
}

// ChatMessage model
type ChatMessage struct {
	Message
	FromUserName string         `db:"from_user_name" json:"fromUserName"`
	ReadBy       types.JSONText `db:"read_by" json:"status"`
}

// WithRelated struct
type WithRelated struct {
	Message
	FromName string         `db:"from_name" json:"fromName"`
	ReadBy   pq.StringArray `db:"read_by" json:"readBy"`
}

// ActivitySnapshot holds the snapshot of activity betwen users
type ActivitySnapshot struct {
	ProrID         uuid.UUID `db:"pror_id" json:"prorID"`
	MessagePreview string    `db:"last_message" json:"lastMessage"`
	UnreadCount    int32     `db:"unread_count" json:"unreadCount"`
	ActiveAt       time.Time `db:"activity_at" json:"activeAt"`
}

// FilterMessage holds values to filter messages
type FilterMessage struct {
	GroupID  uuid.UUID
	BeforeID *int64
	Limit    *int64
}

type ToNotify struct {
	MessID      int64  `db:"mess_id" json:"messID"`
	UserID      string `db:"user_id" json:"userID"`
	SenderName  string `db:"sender_name" json:"senderName"`
	SenderID    string `db:"sender_id" json:"senderID"`
	MessageText string `db:"message_text" json:"message"`
	MessageType string `db:"message_type" json:"messageType"`
	Tokens      Tokens `db:"tokens" json:"tokens"`
}

// ActivityFilter holds values to filter activity
type ActivityFilter struct {
	UserID *string
}

// Name implements the tabua.Namer interface.
func (t Message) Name() string {
	return "message"
}
