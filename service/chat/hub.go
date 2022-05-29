package chat

import (
	"encoding/json"
	"github.com/lib/pq"
	"log"
	"os"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// Persister holds methods to persist and retrieve chat messages
type Persister struct {
	Notify             func(*m.Message) error
	SaveMessage        func(*m.Message) (*m.Message, error)
	SetMessagesRead    func(messID []int64, targets []string) ([]string, error)
	GetUsersForMessage func(messID int64) ([]string, error)
	ListMessages       func(m.FilterMessage) ([]m.ChatMessage, error)
	ListActivity       func(userID string, forGroupID pq.StringArray) ([]m.ActivitySnapshot, error)
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan broadcast

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	persister Persister
	logger    *log.Logger
}

// Options holds the options for a Hub
type Options struct {
	Logger    *log.Logger
	Persister Persister
}

// NewWithOptions creates a new Hub with the given options
func NewWithOptions(opt Options) *Hub {
	return &Hub{
		broadcast:  make(chan broadcast),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		persister:  opt.Persister,
		logger:     opt.Logger,
	}
}

// NewHub creates a Hub
func NewHub() *Hub {
	lg := log.New(os.Stderr, "hub: ", log.Lshortfile)

	return &Hub{
		broadcast:  make(chan broadcast),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		persister: Persister{
			Notify:          func(m *m.Message) error { return nil },
			SaveMessage:     func(m *m.Message) (*m.Message, error) { return m, nil },
			SetMessagesRead: func(messID []int64, targets []string) ([]string, error) { return []string{}, nil },
			ListMessages:    func(f m.FilterMessage) ([]m.ChatMessage, error) { return []m.ChatMessage{}, nil },
			ListActivity: func(userID string, forGroupID pq.StringArray) ([]m.ActivitySnapshot, error) {
				return []m.ActivitySnapshot{}, nil
			},
		},
		logger: lg,
	}
}

// Run starts the Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if prevClient, ok := h.clients[client.identifier]; ok {
				prevClient.conn.Close()
			}
			h.clients[client.identifier] = client
			h.logger.Println("before notify", h.clients)
			go updateClientsStatusNotifier(h)
		case client := <-h.unregister:
			if registeredClient, ok := h.clients[client.identifier]; ok {
				if registeredClient == client {
					delete(h.clients, client.identifier)
				}
				close(client.send)
				go updateClientsStatusNotifier(h)
			}
		case b := <-h.broadcast:
			data, err := json.Marshal(b.action)
			if err != nil {
				h.logger.Println("broadcast error: ", string(data), b, err)
			}
			recipients := strMap(b.recipients)
			for id, client := range h.clients {
				if _, ok := recipients[id]; !ok {
					h.logger.Println("broadcast error: no recipient with id ", string(id), recipients)
					continue
				}
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, id)
				}
			}
		}
	}
}

func strMap(s []string) map[string]string {
	m := map[string]string{}
	for _, v := range s {
		m[v] = v
	}
	return m
}
