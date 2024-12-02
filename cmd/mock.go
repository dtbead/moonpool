package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/urfave/cli/v2"
)

var mock = cli.Command{
	Name:  "mock",
	Usage: "generate mock data for testing purposes",
	Action: func(cCtx *cli.Context) error {
		c, err := config.Open(cCtx.Path("config"))
		if err != nil {
			if cCtx.IsSet("config") && errors.Is(err, os.ErrNotExist) {
				return err
			}

			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}

		archiveIDs, err := api.GenerateMockData(moonpool, cCtx.Int("amount"), cCtx.Bool("tag"), cCtx.Bool("timestamp"))
		if err != nil {
			fmt.Printf("failed to generate mock data. %v\n", err)
			return err
		}

		fmt.Printf("generated mock data with archive id %d-%d", archiveIDs[0], archiveIDs[len(archiveIDs)-1])
		return nil
	},
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "amount",
			Usage: "amount of entries to create",
			Value: 10,
		},
		&cli.BoolFlag{
			Name:  "tag",
			Usage: "insert random tags into each entry",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "timestamp",
			Usage: "insert random timestamps into each entry",
			Value: true,
		},
	},
	Subcommands: []*cli.Command{},
}
