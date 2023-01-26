package i18nextractor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"golang.org/x/exp/maps"
)

type Config struct {
	Inputdir     string   `opts:"default=." help:"Source codes directory path"`
	Lang         []string `help:"The language tag of the extracted messages (e.g. en, en-US, zh-Hant-CN)."`
	Outdir       string   `opts:"default=." help:"Write message files to this directory."`
	FailOnUpdate bool     `opts:"name=fail-on-update"`
	Input        afero.Fs `opts:"-"`
	Output       afero.Fs `opts:"-"`
}

var ErrUpdate = errors.New("translation files are out of date")

const format = "yml"

func (c *Config) Execute() error {
	if c == nil {
		return nil
	}

	if len(c.Lang) == 0 {
		c.Lang = []string{"en"}
	}
	c.Lang = lo.FlatMap(c.Lang, func(l string, _ int) []string {
		return strings.Split(l, ",")
	})

	c.Input = afero.NewBasePathFs(afero.NewOsFs(), c.Inputdir)
	if c.Outdir == "" || path.Clean(c.Outdir) == "." {
		// https://github.com/spf13/afero/issues/344
		c.Output = afero.NewOsFs()
	} else {
		c.Output = afero.NewBasePathFs(afero.NewOsFs(), c.Outdir)
	}
	return c.execute()
}

func (c *Config) execute() error {
	messages := []*i18n.Message{}

	if err := afero.Walk(c.Input, "", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		buf, err := afero.ReadFile(c.Input, path)
		if err != nil {
			return err
		}

		msgs, err := extractMessages(buf)
		if err != nil {
			return err
		}

		messages = append(messages, msgs...)
		return nil
	}); err != nil {
		return err
	}

	messageTemplates := map[string]*i18n.MessageTemplate{}
	for _, m := range messages {
		mt := i18n.NewMessageTemplate(m)
		if mt != nil {
			messageTemplates[m.ID] = mt
		} else {
			messageTemplates[m.ID] = &i18n.MessageTemplate{Message: m}
		}
	}

	update := false
	content := marshalTemplates(messageTemplates, true)
	for _, l := range c.Lang {
		path := fmt.Sprintf("%s.%s", l, format)
		if u, err := c.mergeFile(path, format, content); err != nil {
			return err
		} else if u {
			update = u
		}
	}

	if update && c.FailOnUpdate {
		return ErrUpdate
	}
	return nil
}

func (c *Config) mergeFile(path, format string, data any) (bool, error) {
	f, err := c.Output.Open(path)
	if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return false, err
	}

	var a any
	if f != nil {
		defer func() {
			_ = f.Close()
		}()
		if err := unmarshal(f, &a, format); err != nil {
			return false, err
		}
	}

	merged, updated := merge(a, data)
	if merged == nil {
		return false, nil
	}

	if !updated {
		return false, nil
	}

	o, err := marshal(merged, format)
	if err != nil {
		return false, err
	}

	if err := afero.WriteFile(c.Output, path, o, 0644); err != nil {
		return false, err
	}

	return true, nil
}

func marshalTemplates(messageTemplates map[string]*i18n.MessageTemplate, sourceLanguage bool) any {
	v := make(map[string]any, len(messageTemplates))
	for id, template := range messageTemplates {
		if other := template.PluralTemplates["other"]; sourceLanguage && len(template.PluralTemplates) == 1 &&
			other != nil && template.Description == "" && template.LeftDelim == "" && template.RightDelim == "" {
			v[id] = other.Src
		} else if !onlyID(template) {
			m := map[string]string{}
			if template.Description != "" {
				m["description"] = template.Description
			}
			if !sourceLanguage {
				m["hash"] = template.Hash
			}
			for pluralForm, template := range template.PluralTemplates {
				m[string(pluralForm)] = template.Src
			}
			v[id] = m
		} else {
			v[id] = ""
		}
	}
	return v
}

func onlyID(m *i18n.MessageTemplate) bool {
	return m.ID != "" && len(m.PluralTemplates) == 0
}

func merge(a, b any) (any, bool) {
	am, _ := a.(map[string]any)
	bm, ok := b.(map[string]any)
	if !ok || bm == nil {
		return nil, false
	}

	if len(am) == 0 && len(bm) == 0 {
		return nil, false
	}

	unusedKeys, newKeys := lo.Difference(maps.Keys(am), maps.Keys(bm))
	if len(newKeys) == 0 && len(unusedKeys) == 0 {
		return nil, false
	}

	for _, k := range unusedKeys {
		delete(am, k)
	}
	for _, k := range newKeys {
		am[k] = bm[k]
	}
	return am, true
}

func marshal(v any, format string) ([]byte, error) {
	switch format {
	/*
		case "json":
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			enc.SetIndent("", "  ")
			err := enc.Encode(v)
			return buf.Bytes(), err
		case "toml":
			var buf bytes.Buffer
			enc := toml.NewEncoder(&buf)
			enc.Indent = ""
			err := enc.Encode(v)
			return buf.Bytes(), err
	*/
	case "yaml", "yml":
		return yaml.Marshal(v)
	}
	return nil, fmt.Errorf("unsupported format: %s", format)
}

func unmarshal(r io.Reader, v any, format string) error {
	switch format {
	/*
		case "json":
			enc := json.NewDecoder(r)
			return enc.Decode(v)
		case "toml":
			var buf bytes.Buffer
			enc := toml.NewDecoder(&buf)
			_, err := enc.Decode(v)
			return err
	*/
	case "yaml", "yml":
		return yaml.NewDecoder(r).Decode(v)
	}
	return fmt.Errorf("unsupported format: %s", format)
}
