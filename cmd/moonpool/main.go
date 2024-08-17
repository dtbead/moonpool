package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/server"
	"github.com/dtbead/moonpool/server/www"
	"github.com/urfave/cli/v2"
)

var c config.Config

const (
	CONFIG_DEFAULT_PATH = "config.json"
)

func main() {
	if c == (config.Config{}) {
		var err error
		c, err = config.Open(CONFIG_DEFAULT_PATH)
		if err != nil {
			fmt.Printf("failed to open config. %v\n", err)
		}
	}

	defer fmt.Sprintln(c)
	if err := start().Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func start() *cli.App {
	return &cli.App{
		Name:  "",
		Usage: "start a new moonpool instance",
		Action: func(cCtx *cli.Context) error {
			db, err := archive.OpenSQLite3(c.ArchivePath)
			if err != nil {
				return err
			}

			l := log.NewSlogger(context.Background(), slog.LevelDebug, "api")
			a := api.New(db, l, c)

			m := server.New(a, c)
			defer m.Shutdown()
			// m.Start("127.0.0.1:" + fmt.Sprint(c.WebUIPort))

			web := www.New(*a, c.MediaPath)
			web.Start("127.0.0.1:" + fmt.Sprint(c.WebUIPort))
			return nil
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "webui",
				Usage:       "port to launch webui on",
				Value:       9996,
				Destination: &c.WebUIPort,
			},
			&cli.IntFlag{
				Name:        "api",
				Usage:       "port to launch api on",
				Value:       9995,
				Destination: &c.WebUIPort,
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "config",
				Category: "configuration",
				Usage:    "path to config file",
				Aliases:  []string{"c, conf"},
				Action: func(cCtx *cli.Context) error {
					var err error
					c, err = config.Open(cCtx.Args().First())
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:     "webui",
				Category: "configuration",
				Usage:    "port to serve webui on",
				Action: func(cCtx *cli.Context) error {
					p, err := strconv.Atoi(cCtx.Args().First())
					if err != nil {
						return err
					}

					c.WebUIPort = p
					return nil
				},
			},
			{
				Name:     "api",
				Category: "configuration",
				Usage:    "port to serve api on",
				Action: func(cCtx *cli.Context) error {
					p, err := strconv.Atoi(cCtx.Args().First())
					if err != nil {
						return err
					}

					c.WebUIPort = p
					return nil
				},
			},
		},
	}

}
