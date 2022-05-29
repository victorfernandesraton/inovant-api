package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"github.com/pindamonhangaba/hermes"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	m "gitlab.com/falqon/inovantapp/backend/models"
)

// Mailer struct Mailer
type Mailer struct {
	Mailer *sendgrid.Client
	Hermes *hermes.Hermes
	Config *Config
	Create *m.CreateAccountConfirm
}

// Config struct Config
type Config struct {
	ResetPasswordHTML       string `json:"resetPasswordHtml"`
	ConfirmationAccountHTML string `json:"confirmationAccount"`
	ImagesTemplate          string `json:"imagesTemplate"`
}

// SendPwdResetRequest send password reset to email
func (m *Mailer) SendPwdResetRequest(pr m.PwdReset) error {
	config, err := paramsConfigJSON()
	if err != nil {
		return err
	}

	resetPassHTML, err := ioutil.ReadFile(config.ResetPasswordHTML)
	if err != nil {
		return err
	}

	pr.ImagesTemplate = config.ImagesTemplate

	t, err := template.New(config.ResetPasswordHTML).Parse(string(resetPassHTML))
	if err != nil {
		return err
	}

	var body bytes.Buffer

	if err := t.Execute(&body, pr); err != nil {
		log.Println(err)
	}
	emailBody := body.String()

	subject := "Solicitação de recuperação de senha"
	plainTextContent := "Recuperação de senha"
	htmlContent := emailBody
	from := mail.NewEmail(plainTextContent, os.Getenv("MAIL_FROM"))
	to := mail.NewEmail(pr.Name, pr.Email)
	mess := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SMTP_PASSWORD"))
	response, err := client.Send(mess)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(response.StatusCode)
		log.Println(response.Body)
		log.Println(response.Headers)
	}
	return err
}

// SendConfirmationAccount send account confirmation to email
func (m *Mailer) SendConfirmationAccount(cac *m.CreateAccountConfirm) error {
	fmt.Println("chegou")
	config, err := paramsConfigJSON()
	if err != nil {
		return err
	}

	confAccofHTML, err := ioutil.ReadFile(config.ConfirmationAccountHTML)
	if err != nil {
		return err
	}

	cac.ImagesTemplate = config.ImagesTemplate

	t, err := template.New(config.ConfirmationAccountHTML).Parse(string(confAccofHTML))
	if err != nil {
		return err
	}

	var body bytes.Buffer

	if err := t.Execute(&body, cac); err != nil {
		log.Println(err)
	}
	emailBody := body.String()

	name := "Usuário"
	if cac.Name != nil {
		name = *cac.Name
	}

	subject := "Confirmação de Criação de Conta"
	plainTextContent := "Criação de Conta"
	htmlContent := emailBody
	from := mail.NewEmail(plainTextContent, os.Getenv("MAIL_FROM"))
	to := mail.NewEmail(name, cac.Email)
	mess := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SMTP_PASSWORD"))
	response, err := client.Send(mess)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(response.StatusCode)
		log.Println(response.Body)
		log.Println(response.Headers)
	}
	return err
}

// SendPwdResetAlert send account confirmation to email
func (m *Mailer) SendPwdResetAlert(config ...interface{}) error {
	return nil
}

func paramsConfigJSON() (*Config, error) {
	// Open our jsonFile
	jsonFile, err := os.Open(os.Getenv("CONFIG_EMAIL"))
	if err != nil {
		return nil, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config
	json.Unmarshal(byteValue, &config)

	return &config, nil
}
