package cmd

import (
	"fmt"

	"github.com/dtbead/moonpool/config"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

var c config.Config

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

// openConfig() implicitly loads a JSON file to the global variable 'c'. openConfig will
// exit the program if an error occurs
func openConfig(cCtx *cli.Context) error {
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
		fmt.Println("loading config")
		configPath = cCtx.String("config")
	}

	if err := open(configPath); err != nil {
		return err
	}

	return nil
}
