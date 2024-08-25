package cmd

import (
	"fmt"

	"github.com/dtbead/moonpool/api"
	"github.com/urfave/cli/v2"
)

var mock = cli.Command{
	Name:  "mock",
	Usage: "generate mock data for testing purposes",
	Action: func(cCtx *cli.Context) error {
		db, a, err := newMoonpool(c.MediaPath, c.ArchivePath)
		if err != nil {
			fmt.Printf("failed to launch moonpool instance. %v\n", err)
		}
		defer a.Close()
		defer db.Close()

		archiveIDs, err := api.GenerateMockData(a, cCtx.Int("amount"), cCtx.Bool("tag"))
		if err != nil {
			fmt.Printf("failed to generate mock data. %v\n", err)
			return err
		}

		fmt.Printf("generated mock data with archive id %d-%d", archiveIDs[0], archiveIDs[len(archiveIDs)-1])
		return nil
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "tag",
			Usage: "insert random tags into each entry",
			Value: true,
		},
		&cli.IntFlag{
			Name:  "amount",
			Usage: "amount of entries to create",
			Value: 10,
		},
	},
	Subcommands: []*cli.Command{},
}
