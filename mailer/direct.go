package mailer

import (
	"context"
	"net"
	"strings"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

// NOTE: implemented but it does not work expectedly
type direct struct {
	from string
}

func NewDirect(from string) Mailer {
	if from == "" {
		from = "reearth@localhost:8080"
	}
	return &direct{from: from}
}

func (m *direct) SendMail(ctx context.Context, to []Contact, subject, plainContent, htmlContent string) error {
	emails, err := verifyEmails(to)
	if err != nil {
		return err
	}

	emailHosts, err := m.hosts(emails)
	if err != nil {
		return err
	}

	mxHosts, err := m.lookupHosts(ctx, emailHosts)
	if err != nil {
		return err
	}

	msg, err := m.message(ctx, emails, subject, plainContent, htmlContent)
	if err != nil {
		return err
	}

	for i, to := range emails {
		host := mxHosts[i]
		if err := m.send(ctx, to, host, msg); err != nil {
			return err
		}
	}

	logMail(ctx, to, subject)
	return nil
}

func (m *direct) message(ctx context.Context, emails []string, subject, plainContent, htmlContent string) ([]byte, error) {
	msg := &message{
		to:           emails,
		from:         m.from,
		subject:      subject,
		plainContent: plainContent,
		htmlContent:  htmlContent,
	}
	encodedMsg, err := msg.encodeMessage()
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, err)
	}
	return encodedMsg, nil
}

func (*direct) hosts(addresses []string) ([]string, error) {
	res := make([]string, 0, len(addresses))
	for _, a := range addresses {
		s := strings.SplitAfterN(a, "@", 2)
		if len(s) != 2 {
			return nil, rerror.NewE(i18n.T("invalid email address"))
		}
		res = append(res, s[1])
	}
	return res, nil
}

// lookupHosts resolves the preferred MX host of each domain.
func (*direct) lookupHosts(ctx context.Context, hosts []string) ([]string, error) {
	res := make([]string, 0, len(hosts))
	for _, h := range hosts {
		mxs, err := net.DefaultResolver.LookupMX(ctx, h)
		if err != nil {
			return nil, rerror.NewE(i18n.T("invalid email address"))
		}
		if len(mxs) == 0 {
			return nil, rerror.NewE(i18n.T("invalid email address"))
		}
		res = append(res, strings.TrimSuffix(mxs[0].Host, "."))
	}
	return res, nil
}

func (m *direct) send(ctx context.Context, to string, host string, msg []byte) error {
	c, closeFn, err := dialSMTP(ctx, host, "25")
	if err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	defer closeFn()
	if err := c.Mail(m.from); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	if err := c.Rcpt(to); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	wc, err := c.Data()
	if err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	if _, err = wc.Write(msg); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	if err := wc.Close(); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	if err := c.Quit(); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	return nil
}
