package chat

import (
	"encoding/json"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	types "github.com/jmoiron/sqlx/types"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

type resolver func(*incomming) error

var actionResolver = map[string]resolver{
	"getClientsStatus": getClientsStatusResolver,
	"setMessagesRead":  setMessagesReadResolver,
	"sendMessage":      sendMessageResolver,
	"listMessages":     listMessagesResolver,
	"listActivity":     listActivityResolver,
}

func echoResolver(d *incomming) error {
	p, err := json.Marshal(d.Action)
	if err != nil {
		return err
	}
	d.client.hub.broadcast <- broadcast{
		recipients: []string{d.client.identifier},
		action:     action{Type: "echo", Payload: p},
	}
	return nil
}

type getClientsStatusRequest struct {
	Clients []string `json:"clients"`
}

func getClientsStatusResolver(d *incomming) error {
	in := getClientsStatusRequest{}
	err := json.Unmarshal(d.Action.Payload, &in)
	if err != nil {
		return err
	}

	activeClients := map[string]bool{}
	for _, id := range in.Clients {
		_, ok := d.client.hub.clients[id]
		activeClients[id] = ok
	}
	p, err := json.Marshal(activeClients)
	if err != nil {
		return err
	}
	d.client.hub.broadcast <- broadcast{
		recipients: []string{d.client.identifier},
		action:     action{Type: "clientsStatusList", Payload: p},
	}
	return nil
}

type mpayload struct {
	Type  string          `json:"type"`
	Value string          `json:"value"`
	Data  json.RawMessage `json:"data"`
}

type sendMessageRequest struct {
	Targets []string `json:"targets"`
	Groups  []string `json:"groups"`
	Context string   `json:"context"`
	Message mpayload `json:"message"`
}

type sendMessageResponse struct {
	Context string    `json:"context"`
	Message m.Message `json:"message"`
}

func sendMessageResolver(d *incomming) error {
	sender := d.client.identifier

	in := sendMessageRequest{}
	err := json.Unmarshal(d.Action.Payload, &in)
	if err != nil {
		return err
	}

	if len(in.Groups) == 0 {
		return errors.New("No group target for message")
	}
	// only one receiver group supported
	receiverGroup := in.Groups[0]

	m := &m.Message{
		FromUserID: uuid.FromStringOrNil(sender),
		ProrID:     receiverGroup,
		Type:       in.Message.Type,
		TextValue:  in.Message.Value,
		Data:       types.JSONText(in.Message.Data),
	}

	m, err = d.client.hub.persister.SaveMessage(m)
	if err != nil {
		return err
	}

	targets, err := d.client.hub.persister.GetUsersForMessage(m.MessID)
	if err != nil {
		return err
	}

	err = d.client.hub.persister.Notify(m)
	if err != nil {
		d.client.hub.logger.Println(errors.Wrap(err, "Failed to notify clients for message:"+strconv.Itoa(int(m.MessID))))
	}

	p, err := json.Marshal(sendMessageResponse{Message: *m, Context: in.Context})
	if err != nil {
		return err
	}
	d.client.hub.broadcast <- broadcast{
		recipients: targets,
		action:     action{Type: "receiveMessage", Payload: p},
	}

	return nil
}

type setMessageReadRequest struct {
	Targets []string `json:"targets"`
	IDList  []int64  `json:"messages"`
}

type setMessageReadResponse struct {
	Readers []string `json:"readers"`
	IDList  []int64  `json:"messages"`
}

func setMessagesReadResolver(d *incomming) error {
	in := setMessageReadRequest{}
	err := json.Unmarshal(d.Action.Payload, &in)
	if err != nil {
		return err
	}

	if len(in.Targets) == 0 {
		return errors.New("No target for message")
	}

	alertTargets, err := d.client.hub.persister.SetMessagesRead(in.IDList, in.Targets)
	if err != nil {
		return err
	}
	p, err := json.Marshal(setMessageReadResponse{Readers: alertTargets, IDList: in.IDList})
	if err != nil {
		return err
	}
	d.client.hub.broadcast <- broadcast{
		recipients: alertTargets,
		action:     action{Type: "messageRead", Payload: p},
	}
	return nil
}

type listMessagesRequest struct {
	GroupID  uuid.UUID `json:"groupID"`
	BeforeID *int64    `json:"beforeID"`
	Limit    *int64    `json:"limit"`
}

type listMessagesResponse struct {
	Messages []m.ChatMessage `json:"messages"`
	Filter   m.FilterMessage `json:"appliedFilter"`
}

func listMessagesResolver(d *incomming) error {
	sender := d.client.identifier
	in := listMessagesRequest{}
	err := json.Unmarshal(d.Action.Payload, &in)
	if err != nil {
		return err
	}

	f := m.FilterMessage{
		Limit:    in.Limit,
		BeforeID: in.BeforeID,
		GroupID:  in.GroupID,
	}
	msgs, err := d.client.hub.persister.ListMessages(f)
	if err != nil {
		return errors.Wrap(err, "persister.ListMessages")
	}
	p, err := json.Marshal(listMessagesResponse{Messages: msgs, Filter: f})
	if err != nil {
		return errors.Wrap(err, "listMessagesResponse")
	}
	d.client.hub.broadcast <- broadcast{
		recipients: []string{sender},
		action:     action{Type: "messageList", Payload: p},
	}
	return nil
}

type listActivityRequest struct {
	//Groups []string `json:"groups"`
	Groups pq.StringArray `json:"groups"`
}

type listActivityResponse struct {
	ActivitySnapshots []m.ActivitySnapshot `json:"messages"`
}

func listActivityResolver(d *incomming) error {
	sender := d.client.identifier

	in := listActivityRequest{}
	err := json.Unmarshal(d.Action.Payload, &in)
	if err != nil {
		return err
	}
	snps, err := d.client.hub.persister.ListActivity(sender, in.Groups)
	if err != nil {
		return errors.Wrap(err, "persister.ListActivity")
	}
	p, err := json.Marshal(listActivityResponse{ActivitySnapshots: snps})
	if err != nil {
		return errors.Wrap(err, "Failed to parse listActivityResponse")
	}
	d.client.hub.logger.Println("sending to client", p)
	d.client.hub.broadcast <- broadcast{
		recipients: []string{sender},
		action:     action{Type: "activityList", Payload: p},
	}
	return nil
}
