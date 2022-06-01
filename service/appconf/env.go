package appconf

import (
	"os"
	"strconv"
)

var (
	userDB     string
	passwordDB string
	nameDB     string
	hostDB     string
	portDB     string

	logPath   string
	jwtSecret string

	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string
	mailto   string

	appURL      string
	appUSER     string
	appPASSWORD string
	appAddr     string

	mailFrom  string
	mailAlias string

	uploadLimit string

	accessURL string
)

// SMTP holds env. configuration for SMTP connection
var SMTP = struct {
	Host     string
	Port     int
	User     string
	Password string
	Mail     string
}{}

// Log holds env. configuration for Logging
var Log = struct {
	LogDir string
}{}

// DB holds env. configuration for database connection
var DB = struct {
	User,
	Password,
	Name,
	Host,
	Port string
}{}

// App holds env. configuration for the application
var App = struct {
	URL       string
	User      string
	Password  string
	Address   string
	AccessURL string
}{appURL, appUSER, appPASSWORD, appAddr, accessURL}

// Mail holds env. configuration for email sending
var Mail = struct {
	From,
	Alias string
}{mailFrom, mailAlias}

func init() {
	userDB = os.Getenv("DB_USER")
	passwordDB = os.Getenv("DB_PASSWORD")
	nameDB = os.Getenv("DB_NAME")
	hostDB = os.Getenv("DB_HOST")
	portDB = os.Getenv("DB_PORT")

	logPath = os.Getenv("LOGPATH")
	jwtSecret = os.Getenv("JWT_SCECRET")

	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	smtpUser = os.Getenv("SMTP_USER")
	smtpPass = os.Getenv("SMTP_PASSWORD")
	mailto = os.Getenv("SMTP_MAIL")

	appURL = os.Getenv("APP_URL")
	appUSER = os.Getenv("APP_USER")
	appPASSWORD = os.Getenv("APP_PASSWORD")
	appAddr = os.Getenv("APP_ADDRESS")

	mailFrom = os.Getenv("MAIL_FROM")
	mailAlias = os.Getenv("MAIL_ALIAS")

	uploadLimit = os.Getenv("UPLOAD_LIMIT")

	accessURL = os.Getenv("APP_ACCESSURL")
	if len(smtpHost) > 0 {
		port, err := strconv.Atoi(smtpPort)
		if err != nil {
			panic(err)
		}
		SMTP.Port = port
	}

	SMTP.Host = smtpHost
	SMTP.User = smtpUser
	SMTP.Password = smtpPass
	SMTP.Mail = mailto
	DB.User = userDB
	DB.Password = passwordDB
	DB.Name = nameDB
	DB.Host = hostDB
	DB.Port = portDB

	Log.LogDir = logPath
}

// JWT holds env. configuration for the JWT authentication
var JWT = struct {
	Secret,
	RolesCtxKey,
	ClaimsCtxKey string
}{jwtSecret, "roles", "user"}

// Server holds env. configuration for the webserver
var Server = struct {
	UploadLimit string
}{uploadLimit}
