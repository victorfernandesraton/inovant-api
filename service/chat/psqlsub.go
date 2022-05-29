package chat

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
)

type pgnotification struct {
	Table  string                 `json:"table"`
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

// Subscriber listens to changes in the persisted messages
type Subscriber struct {
	Logger   *log.Logger
	Listener *pq.Listener
}

// NewSubscriber returns a new Subscriber
func NewSubscriber(l *pq.Listener) (*Subscriber, error) {
	// Create a Subscriber and start handling connections
	pb := &Subscriber{Listener: l}
	go pb.handleIncomingNotifications()

	// Return the pgBroadcaster
	return pb, nil
}

// Listen makes the Subscriber's underlying pglistener listen to thh
// specified channel
func (pb *Subscriber) Listen(pgchannel string) error {
	return pb.Listener.Listen(pgchannel)
}

func (pb *Subscriber) handleIncomingNotifications() {
	for {
		select {
		case n := <-pb.Listener.Notify:
			// For some reason after connection loss with the postgres database,
			// the first notifications is a nil notification. Ignore it.
			if n == nil {
				continue
			}
			// Unmarshal JSON in pgnotification struct
			var pgn pgnotification
			_ = json.Unmarshal([]byte(n.Extra), &pgn)
		case <-time.After(60 * time.Second):
			// received no events for 60 seconds, ping connection")
			go func() {
				pb.Listener.Ping()
			}()
		}
	}
}
