package chat

import (
	"encoding/json"
)

func updateClientsStatusNotifier(hub *Hub) error {

	notify := []string{}
	activeClients := map[string]bool{}
	for id, c := range hub.clients {
		notify = append(notify, id)
		activeClients[id] = c != nil
	}
	p, err := json.Marshal(activeClients)
	if err != nil {
		hub.logger.Println("Error matshaling active clients")
		return err
	}
	hub.broadcast <- broadcast{
		recipients: notify,
		action:     action{Type: "clientsStatusList", Payload: p},
	}
	return nil
}
