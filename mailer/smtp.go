package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

type smtpMailer struct {
	host     string
	port     string
	email    string
	username string
	password string
}

func NewSMTP(host, port, username, email, password string) Mailer {
	return &smtpMailer{
		host:     host,
		port:     port,
		username: username,
		email:    email,
		password: password,
	}
}

func (m *smtpMailer) SendMail(ctx context.Context, to []Contact, subject, plainContent, htmlContent string) error {
	emails, err := verifyEmails(to)
	if err != nil {
		return err
	}

	msg := &message{
		to:           emails,
		from:         m.email,
		subject:      subject,
		plainContent: plainContent,
		htmlContent:  htmlContent,
	}

	encodedMsg, err := msg.encodeMessage()
	if err != nil {
		return err
	}

	// Without credentials (e.g. an internal relay) no AUTH is attempted, so
	// relays that only accept HELO keep working.
	var auth smtp.Auth
	if m.username != "" || m.password != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	if len(m.host) == 0 {
		return rerror.NewE(i18n.T("invalid smtp url"))
	}

	if err := sendMail(ctx, m.host, m.port, auth, m.email, emails, encodedMsg); err != nil {
		return err
	}

	logMail(ctx, to, subject)
	return nil
}

// smtpTimeout bounds a whole SMTP exchange when ctx has no earlier deadline.
// net/smtp.SendMail has no timeout and ignores ctx, so a stalled server would
// block the caller indefinitely.
var smtpTimeout = 30 * time.Second

// dialSMTP connects to an SMTP server. Every operation on the returned client
// fails once ctx is done or smtpTimeout has elapsed. Call closeFn when done.
func dialSMTP(ctx context.Context, host, port string) (c *smtp.Client, closeFn func(), err error) {
	deadline := time.Now().Add(smtpTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	dialer := net.Dialer{Deadline: deadline}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, nil, err
	}
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	// Unblock any pending read or write as soon as ctx is cancelled.
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })

	c, err = smtp.NewClient(conn, host)
	if err != nil {
		stop()
		_ = conn.Close()
		return nil, nil, err
	}
	return c, func() {
		stop()
		_ = c.Close()
	}, nil
}

// sendMail is net/smtp.SendMail with a dial timeout, a connection deadline and
// cancellation through ctx.
func sendMail(ctx context.Context, host, port string, a smtp.Auth, from string, to []string, msg []byte) (err error) {
	for _, l := range append([]string{from}, to...) {
		if strings.ContainsAny(l, "\r\n") {
			return errors.New("smtp: a line must not contain CR or LF")
		}
	}

	c, closeFn, err := dialSMTP(ctx, host, port)
	if err != nil {
		return err
	}
	defer closeFn()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}
	if a != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err := c.Auth(a); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}
