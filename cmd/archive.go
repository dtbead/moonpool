package cmd

import (
	mdb "github.com/dtbead/moonpool/db"
	"github.com/urfave/cli/v2"
)

var archive = cli.Command{
	Name:  "archive",
	Usage: "manage a moonpool archive",
	Subcommands: []*cli.Command{
		{
			Name:  "new",
			Usage: "initializes a new, blank moonpool archive location",
			Action: func(cCtx *cli.Context) error {
				db, err := mdb.OpenSQLite3(c.ArchivePath)
				if err != nil {
					return err
				}
				defer db.Close()

				if err := mdb.InitializeSQLite3(db); err != nil {
					return err
				}

				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "database",
					Usage: "path to save archive data & tagging to",
					Value: "archive.sqlite3",
				},
				&cli.StringFlag{
					Name:  "mediapath",
					Usage: "path to save media to",
					Value: "./mediapath",
				},
			},
		},
	},
}
