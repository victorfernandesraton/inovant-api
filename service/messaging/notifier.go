package messaging

import (
	"log"

	"github.com/jasonlvhit/gocron"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

// MessagetNotification holds the message id for new comment event
type MessagetNotification struct {
	MessID int64
}

// Notifier is a service object for scheduling spiders
type Notifier struct {
	Logger             *log.Logger
	MessageNotificator *MessageNotificator
	commentQueue       map[int64]MessagetNotification
}

// Run starts the spider cron
func (c *Notifier) Run() chan bool {
	c.commentQueue = map[int64]MessagetNotification{}
	s := gocron.NewScheduler()
	s.Every(10).Seconds().Do(c.notify)
	return s.Start()
}

func (c *Notifier) notify() {
	for k, n := range c.commentQueue {
		u, err := c.MessageNotificator.Run(n.MessID)
		if err != nil {
			c.Logger.Println("Error retrieving users to notify", err)
			continue
		}
		for _, s := range u {
			for _, n := range s.Tokens.StringArray {
				c.send(n, s)
			}
		}
		delete(c.commentQueue, k)
	}
}

func (c *Notifier) send(token string, m m.ToNotify) {

	// To check the token is valid
	pushToken, err := expo.NewExponentPushToken(token)
	if err != nil {
		c.Logger.Println("Invalid Expo push token", err)
		return
	}

	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)

	// Publish message
	response, err := client.Publish(
		&expo.PushMessage{
			To:    []expo.ExponentPushToken{pushToken},
			Title: "Nova mensagem de \"" + m.SenderName + "\"",
			Body:  string([]rune(m.MessageText)[:15]),
			Data: map[string]string{
				"userID":     m.UserID,
				"senderID":   m.SenderID,
				"senderName": m.SenderName,
				"contents":   m.MessageText,
				"type":       "chat.newMessage",
			},
			Sound:    "default",
			Priority: expo.DefaultPriority,
		},
	)
	if err != nil {
		c.Logger.Println("Failed to publish message", err)
		return
	}

	// Validate responses
	err = response.ValidateResponse()
	if err != nil {
		c.Logger.Println("Invalid response", err)
	}
}

// NotifyComment adds a notification to the queue
func (c *Notifier) NotifyComment(n MessagetNotification) {
	c.commentQueue[n.MessID] = n
}
