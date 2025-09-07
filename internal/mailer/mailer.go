package mailer

import (
	"bytes"
	"embed"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// embed the templates directory to the binary so that we don't need to read it from disk at runtime
//
//go:embed "templates"
var templateFS embed.FS

// Mailer is a custom struct that holds the dialer, which will send the email, and sender name
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func NewMailerFromEnv() *Mailer {
	env := os.Getenv("ENV")

	var host string
	var port int
	var username string
	var pass string
	sender := os.Getenv("SENDER")

	if env == "PROD" {
		host = os.Getenv("MAILTRAP_HOST")
		port, _ = strconv.Atoi(os.Getenv("MAILTRAP_PORT"))
		username = os.Getenv("MAILTRAP_USERNAME")
		pass = os.Getenv("MAILTRAP_PASSWORD")
	} else {
		// dev: use MailHog
		host = "localhost"
		port = 1025
		username = ""
		pass = ""
	}

	return New(host, port, username, pass, sender)
}

func New(host string, port int, username, password, sender string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	// add a 5 sec timoout to the dialer so that processes don't take long time or forever
	dialer.Timeout = 5 * time.Second

	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (mailer *Mailer) Send(recipient, templateFile string, data map[string]any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// render subject template
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// render plain text template
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// render HTML body template
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// create email message and set headers and bodies
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", mailer.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// send the email
	return mailer.dialer.DialAndSend(msg)
}
