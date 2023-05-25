package mailer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
	"github.com/reearth/reearthx/log"
	"github.com/samber/lo"
)

var (
	charSet = "UTF-8"
)

type awsMailer struct {
	sender accountgateway.Contact
	client *ses.Client
}

func NewSES(ctx context.Context, senderName, senderEmail string) accountgateway.Mailer {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Errorf("mail: filed to load ses config: %+v\n", err)
		return nil
	}

	return &awsMailer{
		sender: accountgateway.Contact{
			Email: senderEmail,
			Name:  senderName,
		},
		client: ses.NewFromConfig(cfg),
	}
}

func (m *awsMailer) SendMail(tos []accountgateway.Contact, subject, plainContent, htmlContent string) error {
	mail := &ses.SendEmailInput{
		Destination: &types.Destination{
			CcAddresses: []string{},
			ToAddresses: lo.Map(tos, func(t accountgateway.Contact, _ int) string {
				return formatContact(t)
			}),
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlContent),
				},
				Text: &types.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(plainContent),
				},
			},
			Subject: &types.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(formatContact(m.sender)),
	}

	_, err := m.client.SendEmail(context.TODO(), mail)

	if err != nil {
		return err
	}

	logMail(tos, subject)
	return nil
}

func formatContact(contact accountgateway.Contact) string {
	return fmt.Sprintf("%s <%s>", contact.Name, contact.Email)
}
