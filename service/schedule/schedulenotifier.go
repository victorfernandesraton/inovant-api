package schedule

import (
	"database/sql"
	"log"
	"strings"

	"github.com/jasonlvhit/gocron"
	"github.com/jmoiron/sqlx"
	"gitlab.com/falqon/inovantapp/backend/service"

	sq "github.com/elgris/sqrl"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var schedNoti *gocron.Scheduler

//var psql sq.StatementBuilderType

//ScheduleNotifier service to stagger schedules
type ScheduleNotifier struct {
	DB     *sqlx.DB
	Logger *log.Logger
}

func init() {
	sched = gocron.NewScheduler()
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

// Start resets cron timer and loads from database config
func (s *ScheduleNotifier) Start() chan bool {
	s.Run()
	x := gocron.NewScheduler()
	x.Clear()
	x.Every(1).Minutes().Do(s.Run)
	<-x.Start()
	return sched.Start()
}

//Run service to run Scheduling algorithm
func (s *ScheduleNotifier) Run() error {
	sch, err := lastFifteenMinutes(s.DB)
	if err != nil {
		s.Logger.Println("Last Fifteen Minutes function error: ", err)
	}
	if len(sch) > 0 {
		for _, v := range sch {
			s.Send(v.Token, v)
		}
	}
	return nil
}

//Send service to run send notification
func (s *ScheduleNotifier) Send(token *string, schNot m.ScheduleNotifier) {
	if token == nil {
		s.Logger.Println("Invalid Token")
		return
	}
	replaceToken := strings.Replace(*token, "{", "", -1)
	replaceToken = strings.Replace(replaceToken, "}", "", -1)
	messageNotification := "Faltam apenas 15 minutos para o encerramento do seu horário"
	//To check the token is valid
	pushToken, err := expo.NewExponentPushToken(replaceToken)
	if err != nil {
		s.Logger.Println("Invalid Expo push token", err)
		return
	}
	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)
	// Publish message
	response, err := client.Publish(
		&expo.PushMessage{
			To:   []expo.ExponentPushToken{pushToken},
			Body: messageNotification,
			Data: map[string]string{
				"userID":   schNot.UserID.String(),
				"scheID":   schNot.ScheID.String(),
				"doctID":   schNot.DoctID.String(),
				"contents": messageNotification,
				"type":     "notification.newNotify",
			},
			Sound:    "default",
			Priority: expo.DefaultPriority,
		},
	)
	if err != nil {
		s.Logger.Println("Failed to publish message", err)
		return
	}

	// Validate responses
	err = response.ValidateResponse()
	if err != nil {
		s.Logger.Println("Invalid response", err)
	}
}

func lastFifteenMinutes(db service.DB) ([]m.ScheduleNotifier, error) {
	lastMinutes := "00:15:00"
	afterMinutes := "00:14:00"
	sch := []m.ScheduleNotifier{}
	args := []interface{}{lastMinutes, afterMinutes}
	query := `SELECT sche_id, doct_id, room_id, start_at, end_at, plan, s.info, s.created_at, deleted_at, u.user_id, u.push_tokens
				FROM schedule s
				LEFT JOIN doctor d USING (doct_id)
				LEFT JOIN "user" u USING (user_id)
				WHERE deleted_at IS NULL
				AND start_at::DATE = now()::DATE
				AND
				(
					(EXTRACT(EPOCH FROM (end_at)::TIME) - EXTRACT(EPOCH FROM (NOW())::TIME)) <= EXTRACT(EPOCH FROM ($1)::TIME)
					AND
					(EXTRACT(EPOCH FROM (end_at)::TIME) - EXTRACT(EPOCH FROM (NOW())::TIME)) > EXTRACT(EPOCH FROM ($2)::TIME)
				)
			`
	err := db.Select(&sch, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return sch, nil
}
