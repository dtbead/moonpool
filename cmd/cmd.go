package cmd

import (
	"fmt"
	"strings"

	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/internal/profile"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

var moonpoolConfig config.Config
var profiler profile.Profile

func NewApp() cli.App {
	app := cli.NewApp()
	app.Name = "moonpool"
	app.Usage = "self-hosted media tagging server"
	app.Commands = []*cli.Command{
		&launch,
		&archive,
		&mock,
	}

	app.SliceFlagSeparator = ","

	app.Flags = []cli.Flag{
		&cli.PathFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "path to JSON configuration file",
			DefaultText: CONFIG_DEFAULT_PATH,
			Value:       CONFIG_DEFAULT_PATH,
		},
		&cli.PathFlag{
			Name:    "database",
			Aliases: []string{"db"},
			Usage:   "path to moonpool database file",
			Value:   config.DefaultValues().ArchivePath,
		},
		&cli.PathFlag{
			Name:    "media",
			Aliases: []string{"m"},
			Usage:   "path to moonpool root media folder",
			Value:   config.DefaultValues().MediaPath,
		},
		&cli.PathFlag{
			Name:    "thumbnail",
			Aliases: []string{"t"},
			Usage:   "path to moonpool thumbnail database file",
			Value:   config.DefaultValues().ThumbnailPath,
		},
		&cli.StringFlag{
			Name:     "profile",
			Category: "debug",
			Aliases:  []string{"prof", "p"},
			Usage:    "enable performance profiling for debugging purposes ('cpu' OR 'memory')",
			Value:    config.DefaultValues().Logging.Profiling,
		},
	}

	app.Before = func(cCtx *cli.Context) error {
		if !cCtx.IsSet("config") {
			moonpoolConfig = config.DefaultValues()
			fmt.Println("using default config settings")
			return nil
		}

		c, err := config.Open(cCtx.Path("config"))
		if err != nil {
			return err
		}
		fmt.Println("using custom config settings")

		moonpoolConfig = c

		if cCtx.IsSet("profile") {
			moonpoolConfig.Logging.Profiling = cCtx.String("profile")
		}

		if strings.EqualFold(moonpoolConfig.Logging.Profiling, config.PROFILING_CPU) {
			profiler, err = profile.New(config.PROFILING_CPU)
			if err != nil {
				return err
			}
		}

		return nil
	}

	app.After = func(cCtx *cli.Context) error {
		return profiler.Stop()
	}

	return *app
}
