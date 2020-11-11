package notification

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	SenderNotify     *mail.Email
	SenderPostmaster *mail.Email

	templatesNames = []string{"request", "invitation", "password-reset"}
)

type TemplateMail struct {
	Sender *mail.Email

	Subject string

	ReceiverName  string
	ReceiverEmail string

	Template     string
	TemplateData interface{}
}

type MailOpts struct {
	Key             string
	Sender          string
	NotifyEmail     string
	PostmasterEmail string
	TemplatePath    string
}

type Mailer interface {
	Send(m TemplateMail) error
}

type service struct {
	client    *sendgrid.Client
	templates map[string]*template.Template
}

func New(opts MailOpts) Mailer {
	templates := make(map[string]*template.Template)

	for _, n := range templatesNames {
		path := fmt.Sprintf("%s/%s.html", opts.TemplatePath, n)
		templates[n] = FileTemplate(path)
	}

	// mail senders
	SenderNotify = mail.NewEmail(opts.Sender, opts.NotifyEmail)

	// sendgrid client
	client := sendgrid.NewSendClient(opts.Key)

	return &service{client, templates}
}

func (s *service) Send(m TemplateMail) error {
	htmlTmpl, ok := s.templates[m.Template]
	if !ok {
		msg := fmt.Sprintf("template with key \"%s\" doesn't exist", m.Template)
		panic(errors.New(msg))
	}

	rcv := mail.NewEmail(m.ReceiverName, m.ReceiverEmail)

	buf, err := ioutil.ReadAll(ExecuteTemplate(htmlTmpl, m.TemplateData))
	if err != nil {
		panic(err)
	}

	message := mail.NewSingleEmail(m.Sender, m.Subject, rcv, "Placeolder Text", string(buf))
	if res, err := s.client.Send(message); err != nil && res.StatusCode != 200 {
		return err
	} else if res.StatusCode >= 400 {
		return errors.New(res.Body)
	}

	return nil
}
