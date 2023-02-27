package mailer

import (
	"fmt"

	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
)

const loggerSep = "======================="

type logger struct{}

func NewLogger() accountgateway.Mailer {
	return &logger{}
}

func (m *logger) SendMail(to []accountgateway.Contact, subject, plainContent, _ string) error {
	logMail(to, subject)
	fmt.Printf("%s\n%s\n%s\n", loggerSep, plainContent, loggerSep)
	return nil
}
