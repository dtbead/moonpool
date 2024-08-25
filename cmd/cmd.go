package cmd

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/server"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

var c config.Config
var a *api.API
var w server.Moonpool

func initConfig() {
	if c == (config.Config{}) {
		var err error
		c, err = config.Open(CONFIG_DEFAULT_PATH)
		if err != nil {
			fmt.Printf("failed to open config. %v. using defaults\n", err)
			c = config.DefaultValues()
		}
	}
}

func OpenMoonpool() (*sql.DB, *api.API, error) {
	d, err := db.OpenSQLite3(c.ArchivePath)
	if err != nil {
		return nil, nil, err
	}

	a := api.New(d, log.NewSlogger(context.Background(), log.LogLevelVerbose, "api"), api.Config{MediaLocation: c.MediaPath})
	return d, a, nil
}

func NewApp() cli.App {
	initConfig()

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		&launch,
		&archive,
		&mock,
	}
	app.SliceFlagSeparator = ","

	return *app
}
