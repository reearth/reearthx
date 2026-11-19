package mailer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/email_sending"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/shared"
)

// cloudflareMailer sends mail through Cloudflare Email Service.
type cloudflareMailer struct {
	sender    Contact
	accountID string
	service   *email_sending.EmailSendingService
}

func NewCloudflare(senderName, senderEmail, accountID, apiToken string, opts ...option.RequestOption) Mailer {
	// Build only the email sending service instead of cloudflare.NewClient, which
	// would also pick up CLOUDFLARE_* credentials from the environment.
	defaults := []option.RequestOption{
		option.WithEnvironmentProduction(),
		option.WithAPIToken(apiToken),
		// The SDK retries up to 10 times by default. A retried send after a
		// timeout or 5xx can deliver the same mail twice, so don't retry.
		option.WithMaxRetries(0),
		option.WithRequestTimeout(30 * time.Second),
	}
	return &cloudflareMailer{
		sender:    Contact{Name: senderName, Email: senderEmail},
		accountID: accountID,
		service:   email_sending.NewEmailSendingService(append(defaults, opts...)...),
	}
}

func (m *cloudflareMailer) SendMail(ctx context.Context, to []Contact, subject, plainContent, htmlContent string) error {
	if _, err := verifyEmails(to); err != nil {
		return err
	}

	// Send one message per recipient so recipients don't see each other's addresses.
	for _, t := range to {
		if err := m.send(ctx, t, subject, plainContent, htmlContent); err != nil {
			return err
		}
	}

	logMail(ctx, to, subject)
	return nil
}

func (m *cloudflareMailer) send(ctx context.Context, to Contact, subject, plainContent, htmlContent string) error {
	// The SDK's "to" only accepts bare addresses, so recipient names are not sent.
	params := email_sending.EmailSendingSendParams{
		AccountID: cloudflare.F(m.accountID),
		From:      cloudflare.F(m.from()),
		To:        cloudflare.F[email_sending.EmailSendingSendParamsToUnion](shared.UnionString(to.Email)),
		Subject:   cloudflare.F(subject),
	}
	if plainContent != "" {
		params.Text = cloudflare.F(plainContent)
	}
	if htmlContent != "" {
		params.HTML = cloudflare.F(htmlContent)
	}

	res, err := m.service.Send(ctx, params)
	if err != nil {
		var apiErr *cloudflare.Error
		if errors.As(err, &apiErr) {
			msgs := make([]string, 0, len(apiErr.Errors))
			for _, e := range apiErr.Errors {
				msgs = append(msgs, fmt.Sprintf("%d: %s", e.Code, e.Message))
			}
			return fmt.Errorf("cloudflare: send failed with status %d: %s", apiErr.StatusCode, strings.Join(msgs, "; "))
		}
		return fmt.Errorf("cloudflare: send failed: %w", err)
	}

	if res != nil && len(res.PermanentBounces) > 0 {
		return fmt.Errorf("cloudflare: permanently bounced: %s", strings.Join(res.PermanentBounces, ", "))
	}

	return nil
}

func (m *cloudflareMailer) from() email_sending.EmailSendingSendParamsFromUnion {
	if m.sender.Name == "" {
		return shared.UnionString(m.sender.Email)
	}
	return email_sending.EmailSendingSendParamsFromEmailSendingEmailAddressObject{
		Address: cloudflare.F(m.sender.Email),
		Name:    cloudflare.F(m.sender.Name),
	}
}
