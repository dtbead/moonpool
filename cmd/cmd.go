package cmd

import (
	"context"
	"database/sql"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/log"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

var c config.Config

func OpenMoonpool() (*sql.DB, *api.API, error) {
	d, err := db.OpenSQLite3(c.ArchivePath)
	if err != nil {
		return nil, nil, err
	}

	a := api.New(d, log.NewSlogger(context.Background(), log.LogLevelInfo, "api"), api.Config{MediaLocation: c.MediaPath})
	return d, a, nil
}

func NewApp() cli.App {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		&launch,
		&archive,
		&mock,
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "path to JSON configuration file",
			Value:   CONFIG_DEFAULT_PATH,
		},
	}

	app.SliceFlagSeparator = ","
	return *app
}

func initConfig() error {
	tmp, err := config.Open(CONFIG_DEFAULT_PATH)
	if err != nil {
		return err
	}

	c = tmp

	return nil
}
