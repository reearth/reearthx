package mailer

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/reearth/reearthx/log"
)

type Mailer interface {
	SendMail(toContacts []Contact, subject, plainContent, htmlContent string) error
}

type Contact struct {
	Email string
	Name  string
}

type Config struct {
	Mailer   string
	SMTP     SMTPConfig
	SendGrid SendGridConfig
	SES      SESConfig
}

type SendGridConfig struct {
	Email string
	Name  string
	API   string
}

type SMTPConfig struct {
	Host         string
	Port         string
	SMTPUsername string
	Email        string
	Password     string
}

type SESConfig struct {
	Email string
	Name  string
}

func New(conf *Config) Mailer {
	if conf.Mailer == "sendgrid" {
		log.Infoln("mailer: sendgrid is used")
		return NewSendGrid(conf.SendGrid.Name, conf.SendGrid.Email, conf.SendGrid.API)
	}
	if conf.Mailer == "smtp" {
		log.Infoln("mailer: smtp is used")
		return NewSMTP(conf.SMTP.Host, conf.SMTP.Port, conf.SMTP.SMTPUsername, conf.SMTP.Email, conf.SMTP.Password)
	}
	log.Infoln("mailer: logger is used")
	return NewLogger()
}

func verifyEmails(contacts []Contact) ([]string, error) {
	emails := make([]string, 0, len(contacts))
	for _, c := range contacts {
		_, err := mail.ParseAddress(c.Email)
		if err != nil {
			return nil, fmt.Errorf("invalid email %s", c.Email)
		}
		emails = append(emails, c.Email)
	}

	return emails, nil
}

type message struct {
	to           []string
	from         string
	subject      string
	plainContent string
	htmlContent  string
}

func (m *message) encodeContent() (string, error) {
	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()

	altBuffer, err := writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"multipart/alternative; boundary=" + boundary}})
	if err != nil {
		return "", err
	}
	altWriter := multipart.NewWriter(altBuffer)
	err = altWriter.SetBoundary(boundary)
	if err != nil {
		return "", err
	}
	var content io.Writer
	content, err = altWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain"}})
	if err != nil {
		return "", err
	}

	_, err = content.Write([]byte(m.plainContent + "\r\n\r\n"))
	if err != nil {
		return "", err
	}
	content, err = altWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html"}})
	if err != nil {
		return "", err
	}
	_, err = content.Write([]byte(m.htmlContent + "\r\n"))
	if err != nil {
		return "", err
	}
	_ = altWriter.Close()
	return buf.String(), nil
}

func (m *message) encodeMessage() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.subject))
	buf.WriteString(fmt.Sprintf("From: %s\n", m.from))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.to, ",")))
	content, err := m.encodeContent()
	if err != nil {
		return nil, err
	}
	buf.WriteString(content)

	return buf.Bytes(), nil
}

type ToList []Contact

func (l ToList) String() string {
	tos := &strings.Builder{}
	for i, t := range l {
		if t.Name != "" {
			_, _ = tos.WriteString(t.Name)
			if t.Email != "" {
				_, _ = tos.WriteString(" ")
			}
		}
		if t.Email != "" {
			_, _ = tos.WriteString("<")
			_, _ = tos.WriteString(t.Email)
			_, _ = tos.WriteString(">")
		}
		if len(l)-1 > i {
			_, _ = tos.WriteString(", ")
		}
	}
	return tos.String()
}

func logMail(to ToList, subject string) {
	log.Infof("mailer: mail sent: To: %s, Subject: %s", to, subject)
}
