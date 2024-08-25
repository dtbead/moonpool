package cmd

import (
	"context"
	"fmt"

	mdb "github.com/dtbead/moonpool/db"
	"github.com/urfave/cli/v2"
)

var archive = cli.Command{
	Name:    "archive",
	Usage:   "manage a moonpool archive",
	Aliases: []string{"a"},
	Subcommands: []*cli.Command{
		&archiveNew,
		&archiveTag,
	},
}

var archiveNew = cli.Command{
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
}

var archiveTag = cli.Command{
	Name:  "tag",
	Usage: "manage tags",
	Subcommands: []*cli.Command{
		{
			Name:    "update",
			Aliases: []string{"u", "set", "s"},
			Usage:   "assigns or removes tags associated with a given archive id",
			Description: `modify tags of a given archive id. setting tags can be done by with
			adding tags: --tag "foo, bar, 123"
			removing tags: --tag "-foo, -bar"
			`,
			Action: func(cCtx *cli.Context) error {
				_, a, err := OpenMoonpool()
				if err != nil {
					return err
				}
				defer a.Close()

				if err := a.NewSavepoint(context.Background(), "tagupdate"); err != nil {
					return err
				}
				defer a.RollbackSavepoint(context.Background(), "tagupdate")

				tagsOld, err := a.GetTags(context.Background(), cCtx.Int64("id"))
				if err != nil {
					return err
				}

				var add, remove []string
				for _, v := range cCtx.StringSlice("tags") {
					if []rune(v)[0] == '-' {
						str := string([]rune(v)[1:])
						remove = append(remove, str)
					} else {
						add = append(add, v)
					}
				}

				if err := a.RemoveTags(context.Background(), cCtx.Int64("id"), remove); err != nil {
					return err
				}

				if err := a.SetTags(context.Background(), cCtx.Int64("id"), add); err != nil {
					return err
				}

				tagsNew, err := a.GetTags(context.Background(), cCtx.Int64("id"))
				if err != nil {
					return err
				}

				differenceAdd, differenceRemove := 0, 0
				for _, removed := range remove {
					if !contains(tagsNew, removed) && contains(tagsOld, removed) {
						differenceRemove++
					}
				}

				for _, added := range add {
					if contains(tagsNew, added) && !contains(tagsOld, added) {
						differenceAdd++
					}
				}

				if err := a.ReleaseSavepoint(context.Background(), "tagupdate"); err != nil {
					return err
				}

				fmt.Printf("%d tag(s) affected (%d added | %d removed)", differenceAdd+differenceRemove, differenceAdd, differenceRemove)
				return nil
			},
			Flags: []cli.Flag{
				&cli.Int64Flag{
					Name:     "id",
					Aliases:  []string{"i, a"},
					Value:    -1,
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:     "tags",
					Aliases:  []string{"t"},
					Usage:    "comma separated tags to insert/remove",
					Required: true,
				},
			},
		},
	},
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
