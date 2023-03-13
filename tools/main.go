package main

import (
	"fmt"
	"os"

	"github.com/jpillora/opts"
	"github.com/reearth/reearthx/tools/i18nextractor"
	"github.com/reearth/reearthx/tools/migrategen"
)

type Config struct {
	Command     string                `opts:"mode=cmdname"`
	I18nExtract *i18nextractor.Config `opts:"name=i18n-extract,mode=cmd"`
	Migrategen  *migrategen.Config    `opts:"name=migrategen,mode=cmd"`
}

func main() {
	c := Config{}
	opts.Parse(&c)

	var err error
	switch c.Command {
	case "i18n-extract":
		err = c.I18nExtract.Execute()
	case "migrategen":
		err = c.Migrategen.Execute()
	}

	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%s\n", err))
		os.Exit(1)
	}
}
