package messaging

import (
	"encoding/json"
	"github.com/lib/pq"
)

// Tokens is a string array
type Tokens struct {
	pq.StringArray
}

// MarshalJSON implements the json.Marshaler interface.
func (c Tokens) MarshalJSON() ([]byte, error) {
	jsonValue, err := json.Marshal(c.StringArray)
	if err != nil {
		return nil, err
	}
	return jsonValue, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *Tokens) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &c.StringArray)
}
