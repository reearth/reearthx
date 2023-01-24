package i18n

import (
	"io/fs"

	"github.com/goccy/go-yaml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Message = i18n.Message

type Bundle struct {
	bundle *i18n.Bundle
}

type Localizer = i18n.Localizer

type LocalizeConfig = i18n.LocalizeConfig

func T(id string) *Message {
	return &Message{
		ID: id,
	}
}

func NewBundle(defaultLanguage language.Tag) *Bundle {
	b := i18n.NewBundle(defaultLanguage)
	b.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	b.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	return &Bundle{
		bundle: b,
	}
}

func (b *Bundle) AddMessages(tag language.Tag, messages ...*Message) error {
	return b.bundle.AddMessages(tag, messages...)
}

func (b *Bundle) MustAddMessages(tag language.Tag, messages ...*Message) {
	b.bundle.MustAddMessages(tag, messages...)
}

func (b *Bundle) LoadBytes(data []byte, name string) error {
	_, err := b.bundle.ParseMessageFileBytes(data, name)
	return err
}

func (b *Bundle) MustLoadBytes(data []byte, name string) {
	b.bundle.MustParseMessageFileBytes(data, name)
}

func (b *Bundle) LoadFS(fs fs.FS, paths []string) error {
	for _, p := range paths {
		_, err := b.bundle.LoadMessageFileFS(fs, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bundle) MustLoadFS(fs fs.FS, paths ...string) {
	for _, p := range paths {
		_, err := b.bundle.LoadMessageFileFS(fs, p)
		if err != nil {
			panic(err)
		}
	}
}

func (b *Bundle) LanguageTags() []language.Tag {
	return b.bundle.LanguageTags()
}

func NewLocalizer(bundle *Bundle, langs ...string) *Localizer {
	return i18n.NewLocalizer(bundle.bundle, langs...)
}

func DefaultMessage(m *Message) string {
	if m == nil {
		return ""
	}
	if m.Other != "" {
		return m.Other
	}
	return m.ID
}
