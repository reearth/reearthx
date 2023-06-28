package mailer

import (
	"context"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendgridMailer struct {
	name   string
	email  string
	client *sendgrid.Client
}

func NewSendGrid(senderName, senderEmail, api string) Mailer {
	return &sendgridMailer{
		name:   senderName,
		email:  senderEmail,
		client: sendgrid.NewSendClient(api),
	}
}

func (m *sendgridMailer) SendMail(ctx context.Context, to []Contact, subject, plainContent, htmlContent string) error {
	for _, t := range to {
		sender := mail.NewEmail(m.name, m.email)
		receiver := mail.NewEmail(t.Name, t.Email)
		message := mail.NewSingleEmail(sender, subject, receiver, plainContent, htmlContent)
		_, err := m.client.Send(message)
		if err != nil {
			return err
		}
	}

	logMail(ctx, to, subject)
	return nil
}
