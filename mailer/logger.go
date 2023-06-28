package mailer

import (
	"context"
	"fmt"
)

const loggerSep = "======================="

type logger struct{}

func NewLogger() Mailer {
	return &logger{}
}

func (m *logger) SendMail(ctx context.Context, to []Contact, subject, plainContent, _ string) error {
	logMail(ctx, to, subject)
	fmt.Printf("%s\n%s\n%s\n", loggerSep, plainContent, loggerSep)
	return nil
}
