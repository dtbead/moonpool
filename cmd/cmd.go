package cmd

import (
	"fmt"

	"github.com/dtbead/moonpool/config"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

var c config.Config

func initConfig() {
	if c == (config.Config{}) {
		var err error
		c, err = config.Open(CONFIG_DEFAULT_PATH)
		if err != nil {
			fmt.Printf("failed to open config. %v\n", err)
		}
	}
}

func NewApp() cli.App {
	initConfig()

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		&launch,
		&archive,
	}

	return *app
}
