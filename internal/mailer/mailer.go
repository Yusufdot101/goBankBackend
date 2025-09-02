package mailer

import (
	"bytes"
	"embed"
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

func New(host string, port int, username, password, sender string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	// add a 5 sec timoout to the dialer so that processes don't take long time or forever
	dialer.Timeout = 5 * time.Second

	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (mailer *Mailer) Send(recipient, templateFile string, data any) error {
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
	msg.SetHeader("TO", recipient)
	msg.SetHeader("FROM", mailer.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// send the email
	return mailer.dialer.DialAndSend(msg)
}
