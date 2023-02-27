package mailer

import (
	"sync"

	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
)

type Mock struct {
	lock  sync.Mutex
	mails []Mail
}

type Mail struct {
	To           []accountgateway.Contact
	Subject      string
	PlainContent string
	HTMLContent  string
}

func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) SendMail(to []accountgateway.Contact, subject, text, html string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.mails = append(m.mails, Mail{
		To:           append([]accountgateway.Contact{}, to...),
		Subject:      subject,
		PlainContent: text,
		HTMLContent:  html,
	})
	return nil
}

func (m *Mock) Mails() []Mail {
	return append([]Mail{}, m.mails...)
}
