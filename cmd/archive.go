package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/dtbead/moonpool/api"
	mdb "github.com/dtbead/moonpool/internal/db"
	"github.com/urfave/cli/v2"
)

var archive = cli.Command{
	Name:    "archive",
	Usage:   "manage a moonpool archive",
	Aliases: []string{"a"},
	Subcommands: []*cli.Command{
		&archiveNew,
		&archiveTags,
	},
}

var archiveNew = cli.Command{
	Name:  "new",
	Usage: "initializes a new, blank moonpool archive location",
	Action: func(cCtx *cli.Context) error {
		c.ArchivePath = cCtx.String("database")
		c.MediaPath = cCtx.String("mediapath")
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
			Name:    "database",
			Aliases: []string{"d", "db"},
			Usage:   "path to save archive & tagging data to",
			Value:   "archive.sqlite3",
		},
		&cli.StringFlag{
			Name:    "mediapath",
			Aliases: []string{"m", "media"},
			Usage:   "path to save media to",
			Value:   "/mediapath",
		},
	},
}

var archiveTags = cli.Command{
	Name:     "tags",
	Category: "tags",
	Usage:    "manage tags",
	Subcommands: []*cli.Command{
		&tagsSet,
		&tagsSearch,
		&tagsQuery,
		&tagsList,
	},
}

var tagsSet = cli.Command{
	Name:     "set",
	Category: "tags",
	Aliases:  []string{"s"},
	Usage:    "assigns or removes tags associated with a given archive id",
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
}

var tagsSearch = cli.Command{
	Name:     "search",
	Category: "tags",
	Usage:    "search for a singular tag",
	Action: func(cCtx *cli.Context) error {
		_, a, err := OpenMoonpool()
		if err != nil {
			return err
		}
		defer a.Close()

		res, err := a.SearchTag(cCtx.Context, cCtx.String("tag"))
		if err != nil {
			return err
		}

		for _, v := range res {
			fmt.Printf("found id %d\n", v)
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "tag",
			Usage: "tag to search for",
		},
	},
}

var tagsQuery = cli.Command{
	Name:     "query",
	Category: "tags",
	Usage:    "search for a custom tag query",
	Action: func(cCtx *cli.Context) error {
		_, a, err := OpenMoonpool()
		if err != nil {
			return err
		}
		defer a.Close()

		q := api.NewSearchQuery(cCtx.String("tags"))
		res, err := a.Query(cCtx.Context, q)
		if err != nil {
			return err
		}

		for _, v := range res {
			fmt.Printf("found id %d\n", v)
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "tags",
			Usage: "comma-separated tags to query for",
		},
	},
}

var tagsList = cli.Command{
	Name:     "list",
	Aliases:  []string{"l"},
	Category: "tags",
	Usage:    "list all tags associated with an archive_id",
	Args:     true,
	Action: func(cCtx *cli.Context) error {
		_, a, err := OpenMoonpool()
		if err != nil {
			return err
		}
		defer a.Close()

		tags, err := a.GetTags(context.Background(), cCtx.Int64("id"))
		if err != nil {
			return err
		}

		var tagStr strings.Builder
		tagStr.WriteString(fmt.Sprintf("found %d tag(s)\n", len(tags)))
		for i, v := range tags {
			tagStr.WriteString(fmt.Sprintf("%d. %s\n", i+1, v))
		}

		fmt.Println(tagStr.String())
		return nil
	},
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:     "id",
			Usage:    "archive_id to list tags from",
			Required: true,
		},
	},
}
