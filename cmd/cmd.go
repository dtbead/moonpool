package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/internal/db"
	"github.com/dtbead/moonpool/internal/log"
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
	app.Name = "moonpool"
	app.Usage = "self-hosted media tagging server"
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

	app.Before = func(cCtx *cli.Context) error {
		openConfig(cCtx)
		return nil
	}

	app.SliceFlagSeparator = ","
	return *app
}

// openConfig implicitly loads a JSON file to the global variable 'c'. openConfig will
// exit the program if an error occurs
func openConfig(cCtx *cli.Context) {
	configPath := CONFIG_DEFAULT_PATH

	open := func(s string) error {
		tmp, err := config.Open(s)
		if err != nil {
			return err
		}
		c = tmp
		return nil
	}

	if cCtx.IsSet("config") {
		configPath = cCtx.String("config")
	}

	if err := open(configPath); err != nil {
		fmt.Printf("failed to load config, %v. refusing to use defaults\n", err)
		os.Exit(1)
	}

}
