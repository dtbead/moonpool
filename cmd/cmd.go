package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/log"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

var c config.Config

func initConfig() error {
	if c == (config.Config{}) {
		var err error
		c, err = config.Open(CONFIG_DEFAULT_PATH)
		if err != nil {
			return err
		}
	}
	return nil
}

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
	app.SliceFlagSeparator = ","

	if err := initConfig(); err != nil {
		if isLaunchArgs(app) {
			fmt.Printf("failed to open config, %v. refusing to ignore error due to launch arguments\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("failed to open config, %v. using defaults\n", err)
		}
	}

	return *app
}

func isLaunchArgs(c *cli.App) bool {
	launch := c.Command("launch")
	if contains(launch.Names(), os.Args[1]) {
		return true
	} else {
		return false
	}
}
