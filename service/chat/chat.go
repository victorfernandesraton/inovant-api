package chat

import (
	"encoding/json"
)

type broadcast struct {
	recipients []string `bson:"recipients"`
	action     action   `bson:"action"`
}

type incomming struct {
	client *Client
	Action action `json:"action"`
}

type action struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
