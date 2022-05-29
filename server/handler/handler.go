package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/pindamonhangaba/hermes"
	"github.com/sendgrid/sendgrid-go"

	"gitlab.com/falqon/inovantapp/backend/service"
	"gitlab.com/falqon/inovantapp/backend/service/chat"
	"gitlab.com/falqon/inovantapp/backend/service/mailer"
	"gitlab.com/falqon/inovantapp/backend/service/messaging"
	"gitlab.com/falqon/inovantapp/backend/service/user"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth/rolecache"

	"gitlab.com/falqon/inovantapp/backend/service/actionverification"
	"gitlab.com/falqon/inovantapp/backend/service/appointment"
	"gitlab.com/falqon/inovantapp/backend/service/avaliability"
	"gitlab.com/falqon/inovantapp/backend/service/config"
	"gitlab.com/falqon/inovantapp/backend/service/dashboard"
	"gitlab.com/falqon/inovantapp/backend/service/doctorspecialty"
	"gitlab.com/falqon/inovantapp/backend/service/patient"
	"gitlab.com/falqon/inovantapp/backend/service/room"
	"gitlab.com/falqon/inovantapp/backend/service/schedule"
	"gitlab.com/falqon/inovantapp/backend/service/specialty"

	mw "github.com/labstack/echo/middleware"
	echoSwagger "github.com/pindamonhangaba/echo-swagger"
	m "gitlab.com/falqon/inovantapp/backend/models"
	appconf "gitlab.com/falqon/inovantapp/backend/service/appconf"
	fileman "gitlab.com/falqon/inovantapp/backend/service/filemanager"
	amw "gitlab.com/falqon/inovantapp/backend/service/user/auth/rolecache/mw"
)

// Based on Google JSONC styleguide
// https://google.github.io/styleguide/jsoncstyleguide.xml

type errorResponse struct {
	Error generalError `json:"error"`
}

type generalError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Errors  []detailError `json:"errors,omitempty"`
}

type detailError struct {
	Domain       string  `json:"domain"`
	Reason       string  `json:"reason"`
	Message      string  `json:"message"`
	Location     *string `json:"location,omitempty"`
	LocationType *string `json:"locationType,omitempty"`
	ExtendedHelp *string `json:"extendedHelp,omitempty"`
	SendReport   *string `json:"sendReport,omitempty"`
}

type dataResponse struct {
	// Client sets this value and server echos data in the response
	Context string `json:"context,omitempty"`
	Data    dataer `json:"data"`
}

type dataer interface {
	Data()
}

type dataDetail struct {
	// The kind property serves as a guide to what type of information this particular object stores
	Kind string `json:"kind" example:"resource"`
	// Indicates the language of the rest of the properties in this object (BCP 47)
	Language string `json:"lang,omitempty" example:"pt-br"`
}

func (d dataDetail) Data() {}

type singleItemData struct {
	dataDetail
	Item interface{} `json:"item"`
}

func (d singleItemData) Data() {}

type collectionItemData struct {
	dataDetail
	Items []interface{} `json:"items"`
	// The number of items in this result set
	CurrentItemCount int64 `json:"currentItemCount" example:"1"`
	// The number of items in the result
	ItemsPerPage int64 `json:"itemsPerPage" example:"10"`
	// The index of the first item in data.items
	StartIndex int64 `json:"startIndex" example:"1"`
	// The total number of items available in this set
	TotalItems int64 `json:"totalItems" example:"100"`
	// The index of the current page of items
	PageIndex int64 `json:"pageIndex" example:"1"`
	// The total number of pages in the result set.
	TotalPages int64 `json:"totalPages" example:"10"`
}

func (d collectionItemData) Data() {}

func httpErrorHandler(err error, c echo.Context) {

	// since it's an api, it should always be in json
	// won't be using xml anytime soon
	//isJsonRequest := c.Request().Header().Get("Content-Type") == "application/json"

	if e, ok := err.(*echo.HTTPError); ok {
		c.JSON(e.Code, errorResponse{
			Error: generalError{
				Code:    int64(e.Code),
				Message: e.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, errorResponse{
		Error: generalError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	})
}
func handleHome(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, "inovantappp 1.0")
}

// JWTConfig holds configuration for JWT context
type JWTConfig struct {
	Secret,
	ClaimsCtxKey,
	RolesCtxKey string
}

// ServerConf holds configuration data for the webserver
type ServerConf struct {
	BodyLimit     string
	FileDirectory string
	Address       string
	AppAddress    string
}

// HTTPServer create a service to echo server
type HTTPServer struct {
	DB         *sqlx.DB
	DB2        *sqlx.DB
	Roles      *rolecache.RoleCache
	Auth       *user.Authenticator
	JWTConfig  JWTConfig
	ServerConf ServerConf
	Mailer     *mailer.Mailer
}

// Run create a new echo server
func (h *HTTPServer) Run() {
	// configure rolecache
	h.Roles.GetUserRoles = func(userID string) ([]string, error) {
		roles := []string{}
		UID, err := uuid.FromString(userID)
		if err != nil {
			return roles, err
		}
		g := user.Getter{DB: h.DB}

		usr, err := g.Run(UID)
		if err != nil {
			return roles, err
		}
		return usr.Roles, nil
	}

	// Echo instance
	e := echo.New()
	e.Use(mw.Recover())
	e.Use(mw.Logger())
	e.Use(mw.BodyLimit("12M"))
	e.HTTPErrorHandler = httpErrorHandler

	/// CORS restricted
	// Allows requests from all origins
	// wth GET, PUT, POST or DELETE method.
	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/", handleHome)

	gAPI := e.Group("/api")
	jwtConfig := mw.JWTConfig{
		Claims:     &auth.Claims{},
		SigningKey: []byte(h.JWTConfig.Secret),
		ContextKey: h.JWTConfig.ClaimsCtxKey,
	}
	gAPI.Use(mw.JWTWithConfig(jwtConfig))
	gAPI.Use(amw.EchoMiddleware(h.Roles, amw.JWTConfig{
		RolesCtxKey: h.JWTConfig.RolesCtxKey,
		TokenCtxKey: h.JWTConfig.ClaimsCtxKey,
	}))

	PublicRoutes(h.DB, e, h.Auth)
	PrivateRoutes(h.DB, gAPI, h.JWTConfig)

	e.Logger.Fatal(e.Start(h.ServerConf.Address))
}

/*
 * public routes
 */
func PublicRoutes(db *sqlx.DB, e *echo.Echo, ua *user.Authenticator) error {
	mm := mailer.Mailer{
		Mailer: sendgrid.NewSendClient(appconf.SMTP.Password),
		Config: &mailer.Config{},
		Hermes: &hermes.Hermes{
			Product: hermes.Product{
				Copyright:   "Copyright © 2020 Inovant. Todos direitos reservados.",
				Name:        "Inovant",
				TroubleText: "Se está tendo problemas com o botão '{ACTION}', copie e cole a URL abaixo no navegador.",
			},
		},
	}
	servconf := &service.ServicesConfig{
		APPURL: appconf.App.URL,
		APIURL: appconf.App.Address,
	}

	pv := user.PwdRecoverer{
		DB:     db,
		Mailer: &mm,
		Config: servconf,
	}
	pr := user.PwdReseter{
		DB:     db,
		Mailer: &mm,
	}

	ah := &AuthHandler{
		signin:     ua.Run,
		pwdReset:   pr.Run,
		pwdRecover: pv.Run,
	}
	e.POST("/auth/signin", ah.EmailLogin)
	e.POST("/auth/password-recover", ah.PasswordRecover)
	e.POST("/auth/password-reset/:resetID/:verification", ah.PasswordReset)

	uf := &fileman.Uploader{AccessURL: appconf.App.AccessURL}
	fh := &FileHandler{upload: uf.Run}
	e.GET("/files/:file", fh.Get)
	e.GET("/templates/images/:file", fh.GetTemplateImages)

	return nil
}

/*
 * private routes
 */
// PrivateRoutes create routes to private access
func PrivateRoutes(db *sqlx.DB, gAPI *echo.Group, JWTConfig JWTConfig) error {
	mm := mailer.Mailer{
		Mailer: sendgrid.NewSendClient(appconf.SMTP.Password),
		Config: &mailer.Config{},
		Hermes: &hermes.Hermes{
			Product: hermes.Product{
				Copyright:   "Copyright © 2020 Inovant. Todos direitos reservados.",
				Name:        "Inovant",
				TroubleText: "Se está tendo problemas com o botão '{ACTION}', copie e cole a URL abaixo no navegador.",
			},
		},
	}

	servconf := &service.ServicesConfig{
		APPURL: appconf.App.URL,
		APIURL: appconf.App.Address,
	}
	// File routes
	uf := &fileman.Uploader{AccessURL: appconf.App.AccessURL}
	fh := &FileHandler{upload: uf.Run}
	gAPI.POST("/files", fh.Upload)

	// Chat notification service
	mmn := messaging.MessageNotificator{DB: db}
	noti := messaging.Notifier{
		Logger:             log.New(os.Stderr, "notificator: ", log.Lshortfile),
		MessageNotificator: &mmn,
	}
	go noti.Run()

	// Chat webservice route
	clm := messaging.MessageLister{DB: db}
	cla := messaging.ActivityLister{DB: db}
	csm := messaging.MessageCreator{DB: db}
	csmr := messaging.MessageReadCreator{DB: db}
	cufm := messaging.GetUsersForMessage{DB: db}
	hub := chat.NewWithOptions(chat.Options{
		Logger: log.New(os.Stderr, "hub: ", log.Lshortfile),
		Persister: chat.Persister{
			Notify: func(m *m.Message) error {
				noti.NotifyComment(messaging.MessagetNotification{MessID: m.MessID})
				return nil
			},
			SaveMessage:        csm.Run,
			SetMessagesRead:    csmr.Run,
			GetUsersForMessage: cufm.Run,
			ListMessages:       clm.Run,
			ListActivity:       cla.Run,
		},
	})
	go hub.Run()
	ch := ChatHandler{db: db, hub: hub, claimsCtxKey: JWTConfig.ClaimsCtxKey, rolesCtxKey: JWTConfig.RolesCtxKey}
	gAPI.GET("/connect", ch.Chat)

	// User routes
	uc := &user.Lister{DB: db}
	u := &user.Creator{DB: db, Mailer: &mm, Config: servconf}
	up := &user.Updater{DB: db}
	ug := &user.Getter{DB: db}
	ui := &user.Inactiver{DB: db}
	uac := &user.Activer{DB: db}
	upt := &user.PushTokenSetter{DB: db}
	uh := &UserHandler{
		create:       u.Run,
		update:       up.Run,
		get:          ug.Run,
		inactive:     ui.Run,
		active:       uac.Run,
		list:         uc.Run,
		setPushToken: upt.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.GET("/users/:userID", uh.Get)
	gAPI.PUT("/users/:userID", uh.Update)
	gAPI.GET("/users", uh.List)
	gAPI.POST("/users", uh.Create)
	gAPI.DELETE("/users/:userID", uh.Inactive)
	gAPI.PUT("/users/active/:userID", uh.Active)
	gAPI.POST("/users/:userID/push-tokens", uh.SetPushToken)

	//Doctor routes
	doctC := &user.DoctorCreator{DB: db, Mailer: &mm, Config: servconf}
	doctU := &user.DoctorUpdater{DB: db}
	doctD := &user.DoctorDeleter{DB: db}
	doctL := &user.DoctorLister{DB: db}
	doctG := &user.DoctorGetter{DB: db}
	doctH := &DoctorHandler{
		create:       doctC.Run,
		update:       doctU.Run,
		delete:       doctD.Run,
		list:         doctL.Run,
		get:          doctG.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.POST("/doctors", doctH.Create)
	gAPI.PUT("/doctors/:doctID", doctH.Update)
	gAPI.DELETE("/doctors/:doctID", doctH.Delete)
	gAPI.GET("/doctors", doctH.List)
	gAPI.GET("/doctors/:doctID", doctH.Get)

	//Schedule routes
	scheC := &schedule.Creator{DB: db}
	scheU := &schedule.Updater{DB: db}
	scheD := &schedule.Deleter{DB: db}
	scheUd := &schedule.UpdateDeleter{DB: db}
	scheUs := &schedule.UpdateSchedule{DB: db}
	scheL := &schedule.Lister{DB: db}
	scheG := &schedule.Getter{DB: db}
	scheCa := &schedule.Calendar{DB: db}
	scheOut := &schedule.Outdoor{DB: db}
	scheH := &ScheduleHandler{
		create:         scheC.Run,
		update:         scheU.Run,
		delete:         scheD.Run,
		updateDelete:   scheUd.Run,
		updateSchedule: scheUs.Run,
		list:           scheL.Run,
		get:            scheG.Run,
		calendar:       scheCa.Run,
		outdoor:        scheOut.Run,
		rolesCtxKey:    JWTConfig.RolesCtxKey,
		claimsCtxKey:   JWTConfig.ClaimsCtxKey,
		getErrorMessage: func(err error) generalError {
			if schedule.ScheduleUnavailable(err) {
				return generalError{
					Code: 002,
					//Message: "Schedule unavailable: " + err.Error(),
					Message: err.Error(),
				}
			}
			return generalError{
				Code: 001,
				//Message: "Unspecified error " + err.Error(),
				Message: err.Error(),
			}
		},
	}
	gAPI.POST("/schedules", scheH.Create)
	gAPI.PUT("/schedules/:scheID", scheH.Update)
	gAPI.PUT("/schedules/:scheID/schedule", scheH.UpdateSchedule)
	gAPI.PUT("/schedules/:scheID/deletedAt", scheH.UpdateDeleter)
	gAPI.DELETE("/schedules/:scheID", scheH.Delete)
	gAPI.GET("/schedules", scheH.List)
	gAPI.GET("/schedules/:scheID", scheH.Get)
	gAPI.GET("/calendar", scheH.Calendar)
	gAPI.GET("/outdoor/:roomID", scheH.Outdoor)

	//Appointment routes
	appoC := &appointment.Creator{DB: db}
	appoU := &appointment.Updater{DB: db}
	appoD := &appointment.Deleter{DB: db}
	appoL := &appointment.Lister{DB: db}
	appoG := &appointment.Getter{DB: db}
	appoH := &AppointmentHandler{
		create:       appoC.Run,
		update:       appoU.Run,
		delete:       appoD.Run,
		list:         appoL.Run,
		get:          appoG.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.POST("/appointments", appoH.Create)
	gAPI.PUT("/appointments/:appoID", appoH.Update)
	gAPI.DELETE("/appointments/:appoID", appoH.Delete)
	gAPI.GET("/appointments", appoH.List)
	gAPI.GET("/appointments/:appoID", appoH.Get)

	//Patient routes
	patiC := &patient.Creator{DB: db}
	patiU := &patient.Updater{DB: db}
	patiD := &patient.Deleter{DB: db}
	patiL := &patient.Lister{DB: db}
	patiG := &patient.Getter{DB: db}
	patiH := &PatientHandler{
		create:       patiC.Run,
		update:       patiU.Run,
		delete:       patiD.Run,
		list:         patiL.Run,
		get:          patiG.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.POST("/patients", patiH.Create)
	gAPI.PUT("/patients/:patiID", patiH.Update)
	gAPI.DELETE("/patients/:patiID", patiH.Delete)
	gAPI.GET("/patients", patiH.List)
	gAPI.GET("/patients/:patiID", patiH.Get)

	//ActionVerification routes
	acveC := &actionverification.Creator{DB: db}
	acveU := &actionverification.Updater{DB: db}
	acveD := &actionverification.Deleter{DB: db}
	acveL := &actionverification.Lister{DB: db}
	acveG := &actionverification.Getter{DB: db}
	acveH := &ActionVerificationHandler{
		create: acveC.Run,
		update: acveU.Run,
		delete: acveD.Run,
		list:   acveL.Run,
		get:    acveG.Run,
	}
	gAPI.POST("/actions-verification", acveH.Create)
	gAPI.PUT("/actions-verification/:acveID", acveH.Update)
	gAPI.DELETE("/actions-verification/:acveID", acveH.Delete)
	gAPI.GET("/actions-verification", acveH.List)
	gAPI.GET("/actions-verification/:acveID", acveH.Get)

	//Room routes
	roomC := &room.Creator{DB: db}
	roomU := &room.Updater{DB: db}
	roomD := &room.Deleter{DB: db}
	roomL := &room.Lister{DB: db}
	roomG := &room.Getter{DB: db}
	roomH := &RoomHandler{
		create: roomC.Run,
		update: roomU.Run,
		delete: roomD.Run,
		list:   roomL.Run,
		get:    roomG.Run,
	}
	gAPI.POST("/rooms", roomH.Create)
	gAPI.PUT("/rooms/:roomID", roomH.Update)
	gAPI.DELETE("/rooms/:roomID", roomH.Delete)
	gAPI.GET("/rooms", roomH.List)
	gAPI.GET("/rooms/:roomID", roomH.Get)

	//Avaliability routes
	avalC := &avaliability.Checker{DB: db}
	avalH := &AvaliabilityHandler{
		check:        avalC.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.GET("/avaliability", avalH.Check)

	//Specialty routes
	specialtyC := &specialty.Creator{DB: db}
	specialtyU := &specialty.Updater{DB: db}
	specialtyD := &specialty.Deleter{DB: db}
	specialtyL := &specialty.Lister{DB: db}
	specialtyG := &specialty.Getter{DB: db}
	specialtyH := &SpecialtyHandler{
		create: specialtyC.Run,
		update: specialtyU.Run,
		delete: specialtyD.Run,
		list:   specialtyL.Run,
		get:    specialtyG.Run,
	}
	gAPI.POST("/specialty", specialtyH.Create)
	gAPI.PUT("/specialty/:specID", specialtyH.Update)
	gAPI.DELETE("/specialty/:specID", specialtyH.Delete)
	gAPI.GET("/specialty", specialtyH.List)
	gAPI.GET("/specialty/:specID", specialtyH.Get)

	//DoctorSpecialty routes
	doctorspecialtyC := &doctorspecialty.Creator{DB: db}
	doctorspecialtyU := &doctorspecialty.Updater{DB: db}
	doctorspecialtyD := &doctorspecialty.Deleter{DB: db}
	doctorspecialtyL := &doctorspecialty.Lister{DB: db}
	doctorspecialtyG := &doctorspecialty.Getter{DB: db}
	doctorspecialtyH := &DoctorSpecialtyHandler{
		create: doctorspecialtyC.Run,
		update: doctorspecialtyU.Run,
		delete: doctorspecialtyD.Run,
		list:   doctorspecialtyL.Run,
		get:    doctorspecialtyG.Run,
	}
	gAPI.POST("/doctor-specialty", doctorspecialtyH.Create)
	gAPI.PUT("/doctor-specialty/:doctID", doctorspecialtyH.Update)
	gAPI.DELETE("/doctor-specialty/:doctID/:specID", doctorspecialtyH.Delete)
	gAPI.GET("/doctor-specialty", doctorspecialtyH.List)
	gAPI.GET("/doctor-specialty/:doctID/:specID", doctorspecialtyH.Get)

	//Config routes
	configC := &config.Creator{DB: db}
	configU := &config.Updater{DB: db}
	configD := &config.Deleter{DB: db}
	configL := &config.Lister{DB: db}
	configG := &config.Getter{DB: db}
	configH := &ConfigHandler{
		create:       configC.Run,
		update:       configU.Run,
		delete:       configD.Run,
		list:         configL.Run,
		get:          configG.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.POST("/configs", configH.Create)
	gAPI.PUT("/configs/:key", configH.Update)
	gAPI.DELETE("/configs/:key", configH.Delete)
	gAPI.GET("/configs", configH.List)
	gAPI.GET("/configs/:key", configH.Get)

	//Dashboard routes
	dashboardV := &dashboard.Viewer{DB: db}
	dashboardH := &DashboardHandler{
		view:         dashboardV.Run,
		rolesCtxKey:  JWTConfig.RolesCtxKey,
		claimsCtxKey: JWTConfig.ClaimsCtxKey,
	}
	gAPI.GET("/dashboard", dashboardH.View)

	return nil
}
