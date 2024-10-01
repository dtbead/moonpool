package cmd

import (
	"errors"
	"log/slog"
	"os"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/urfave/cli/v2"
)

const CONFIG_DEFAULT_PATH = "config.json"

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
		&cli.PathFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "path to JSON configuration file",
			Value:   CONFIG_DEFAULT_PATH,
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
			Usage:   "path to moonpool root media folder ",
			Value:   config.DefaultValues().MediaPath,
		},
	}

	app.SliceFlagSeparator = ","
	return *app
}

func OpenMoonpool(filepath string, c config.Config) (*api.API, error) {
	moonpool, err := api.Open(
		api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath},
		slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err != nil {
		return nil, err
	}

	return moonpool, nil
}

// OpenConfig() reads a config file from a path "config" taken from cli.Context.
// If fallbackDefaults is true, it returns a default config if no config path is specified by the user
// AND if the default config location does not exist.
//
// Otherwise, if a config path is specified but does not exist, it returns an error.
func OpenConfig(cCtx cli.Context, fallbackDefaults bool) (config.Config, error) {
	c, err := config.Open(cCtx.Path("config"))
	if err != nil {
		if !cCtx.IsSet("config") && errors.Is(err, os.ErrNotExist) {
			if fallbackDefaults {
				return config.DefaultValues(), err
			}
		}

		return config.Config{}, err
	}

	if cCtx.IsSet("database") {
		c.ArchivePath = cCtx.String("database")
	}

	if cCtx.IsSet("mediapath") {
		c.ArchivePath = cCtx.String("mediapath")
	}

	return c, nil
}
