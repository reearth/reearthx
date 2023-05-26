package mailer

import (
	"fmt"
)

const loggerSep = "======================="

type logger struct{}

func NewLogger() Mailer {
	return &logger{}
}

func (m *logger) SendMail(to []Contact, subject, plainContent, _ string) error {
	logMail(to, subject)
	fmt.Printf("%s\n%s\n%s\n", loggerSep, plainContent, loggerSep)
	return nil
}
