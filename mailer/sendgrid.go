package mailer

import (
	"context"
	"fmt"

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
		res, err := m.client.SendWithContext(ctx, message)
		if err != nil {
			return err
		}
		// Send only fails on transport errors; API errors (401 revoked key,
		// 403 unverified sender, 429 quota, 5xx) come back as a response.
		if res == nil {
			return fmt.Errorf("sendgrid: no response")
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("sendgrid: send failed with status %d: %s", res.StatusCode, res.Body)
		}
	}

	logMail(ctx, to, subject)
	return nil
}
