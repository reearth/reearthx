package rerror

import (
	"errors"
	"fmt"
	"slices"

	"github.com/reearth/reearthx/i18n"
)

type Localizable interface {
	LocalizeError(*i18n.Localizer) error
}

type E struct {
	m      *i18n.Message
	format bool
	args   []any
	err    error
}

func NewE(m *i18n.Message) *E {
	return &E{
		m: m,
	}
}

func FmtE(m *i18n.Message, args ...any) *E {
	return &E{
		m:      m,
		format: true,
		args:   args,
		err:    errors.Unwrap(fmt.Errorf(i18n.DefaultMessage(m), args...)),
	}
}

func WrapE(m *i18n.Message, err error) *E {
	return &E{
		m:   m,
		err: err,
	}
}

func (e *E) LocalizeError(l *i18n.Localizer) error {
	s, err := l.LocalizeMessage(e.m)
	if err != nil || s == "" {
		return errors.New(i18n.DefaultMessage(e.m))
	}

	if e.format {
		args := slices.Clone(e.args)
		for i, a := range args {
			if e2, ok := a.(Localizable); ok {
				args[i] = e2.LocalizeError(l)
			}
		}
		return fmt.Errorf(s, args...)
	}

	if e.err != nil {
		werr := e.err
		if e2, ok := werr.(Localizable); ok {
			werr = e2.LocalizeError(l)
		}
		return &W{Err: werr, Msg: s}
	}

	return errors.New(s)
}

func (e *E) Unwrap() error {
	return e.err
}

func (e *E) Error() string {
	if e.format {
		return fmt.Errorf(i18n.DefaultMessage(e.m), e.args...).Error()
	}
	return i18n.DefaultMessage(e.m)
}

func Localize(l *i18n.Localizer, err error) error {
	if err == nil {
		return nil
	}
	if e2, ok := err.(Localizable); ok {
		return e2.LocalizeError(l)
	}
	return err
}

type Localizer struct {
	l   *i18n.Localizer
	err error
}

func NewLocalizer(err error, l *i18n.Localizer) *Localizer {
	return &Localizer{
		l:   l,
		err: err,
	}
}

func (l *Localizer) Error() string {
	return Localize(l.l, l.err).Error()
}
