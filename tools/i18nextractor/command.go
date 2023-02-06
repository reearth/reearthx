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

	// workaround
	if !filepath.IsAbs(c.Inputdir) {
		dir := filepath.Base(lo.Must(os.Getwd()))
		c.Inputdir = filepath.Join("..", dir, c.Inputdir)
	}
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

		l := len(msgs)
		if l > 0 {
			os.Stderr.WriteString(fmt.Sprintf("%s ... %d messages\n", path, l))
		}

		return nil
	}); err != nil {
		return err
	}

	if len(messages) > 0 {
		os.Stderr.WriteString("\n")
	}
	os.Stderr.WriteString(fmt.Sprintf("%d messages found\n", len(messages)))

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

	os.Stderr.WriteString("done\n")

	if update && c.FailOnUpdate {
		return ErrUpdate
	}
	return nil
}

func (c *Config) mergeFile(path, format string, data any) (bool, error) {
	os.Stderr.WriteString(fmt.Sprintf("writing messages to %s", path))

	f, err := c.Output.Open(path)
	if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		os.Stderr.WriteString("\n")
		return false, err
	}

	a := map[string]any{}
	if f != nil {
		defer func() {
			_ = f.Close()
		}()
		if err := unmarshal(f, &a, format); err != nil {
			os.Stderr.WriteString("\n")
			return false, err
		}
	}

	merged := merge(a, data)
	if merged == nil {
		os.Stderr.WriteString(" ... no updates\n")
		return false, nil
	}

	o, err := marshal(merged, format)
	if err != nil {
		os.Stderr.WriteString("\n")
		return false, err
	}

	if err := afero.WriteFile(c.Output, path, o, 0644); err != nil {
		os.Stderr.WriteString("\n")
		return false, err
	}

	os.Stderr.WriteString(" ... done\n")
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

func merge(a, b any) any {
	am, _ := a.(map[string]any)
	bm, ok := b.(map[string]any)
	if !ok || bm == nil {
		return nil
	}

	if len(am) == 0 && len(bm) == 0 {
		return nil
	}

	unusedKeys, newKeys := lo.Difference(maps.Keys(am), maps.Keys(bm))
	if len(newKeys) == 0 && len(unusedKeys) == 0 {
		return nil
	}

	for _, k := range unusedKeys {
		delete(am, k)
	}
	for _, k := range newKeys {
		am[k] = bm[k]
	}
	return am
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
