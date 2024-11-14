package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/importer"
	mdb "github.com/dtbead/moonpool/internal/db"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/urfave/cli/v2"
)

var archive = cli.Command{
	Name:    "archive",
	Usage:   "manage a moonpool archive",
	Aliases: []string{"a"},
	Subcommands: []*cli.Command{
		&archiveNew,
		&archiveRemove,
		&archiveTags,
		&archiveImport,
		&archiveThumbnails,
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

var archiveThumbnails = cli.Command{
	Name:     "thumbnails",
	Category: "thumbnail",
	Usage:    "manage thumbnails",
	Subcommands: []*cli.Command{
		&thumbnailGenerateIcons,
		&thumbnailGenerateBlurHash,
	},
}

var archiveNew = cli.Command{
	Name:  "new",
	Usage: "initializes a new, blank moonpool archive location",
	Action: func(cCtx *cli.Context) error {
		archive, err := mdb.OpenSQLite3(cCtx.Path("database"))
		if err != nil {
			return err
		}
		defer archive.Close()

		thumbnail, err := mdb.OpenSQLite3(cCtx.Path("thumbnail"))
		if err != nil {
			return err
		}
		defer thumbnail.Close()

		if err := mdb.InitializeArchive(archive); err != nil {
			return err
		}

		if err := mdb.InitializeThumbnail(thumbnail); err != nil {
			return err
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "database",
			Aliases: []string{"d"},
			Usage:   "path to create new archive database",
			Value:   config.DefaultValues().ArchivePath,
		},
		&cli.PathFlag{
			Name:    "thumbnail",
			Aliases: []string{"t"},
			Usage:   "path to create new thumbnail database",
			Value:   config.DefaultValues().ThumbnailPath,
		},
	},
}

var archiveImport = cli.Command{
	Name:  "import",
	Usage: "imports a new file into moonpool",
	Action: func(cCtx *cli.Context) error {
		c, err := OpenConfig(*cCtx, false)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath, ThumbnailLocation: c.ThumbnailPath},
			log.New(log.StringToLogLevel(c.Logging.LogLevel)))
		if err != nil {
			return err
		}
		defer moonpool.Close()

		f, err := os.Open(cCtx.Path("file"))
		if err != nil {
			return err
		}

		importer, err := importer.New(f, path.Ext(cCtx.Path("file")))
		if err != nil {
			return err
		}

		archive_id, err := moonpool.Import(context.Background(), importer)
		if err != nil {
			return err
		}

		if err := moonpool.SetTags(context.Background(), archive_id, cCtx.StringSlice("tags")); err != nil {
			return err
		}

		if err := moonpool.GenerateThumbnail(context.Background(), archive_id); err != nil {
			return err
		}

		fmt.Printf("imported new entry with archive id %d\n", archive_id)
		return nil
	},
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "file to import",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:    "tags",
			Aliases: []string{"t"},
			Usage:   "tags to assign",
		},
	},
}

var archiveRemove = cli.Command{
	Name:  "remove",
	Usage: "completely remove an entry from moonpool",
	Action: func(cCtx *cli.Context) error {
		c, err := OpenConfig(*cCtx, false)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(api.Config{
			ArchiveLocation:   c.ArchivePath,
			ThumbnailLocation: c.ThumbnailPath,
			MediaLocation:     c.MediaPath,
		}, log.New(log.StringToLogLevel(c.Logging.LogLevel)))
		if err != nil {
			return err
		}

		if err := moonpool.RemoveArchive(context.Background(), cCtx.Int64("id")); err != nil {
			return err
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:     "id",
			Usage:    "archive to remove",
			Required: true,
		},
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
		c, err := OpenConfig(*cCtx, false)
		if err != nil {

		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}

		if err := moonpool.NewSavepoint(context.Background(), "tagupdate"); err != nil {
			return err
		}
		defer moonpool.RollbackSavepoint(context.Background(), "tagupdate")

		tagsOld, err := moonpool.GetTags(context.Background(), cCtx.Int64("id"))
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

		if err := moonpool.RemoveTags(context.Background(), cCtx.Int64("id"), remove); err != nil {
			return err
		}

		if err := moonpool.SetTags(context.Background(), cCtx.Int64("id"), add); err != nil {
			return err
		}

		tagsNew, err := moonpool.GetTags(context.Background(), cCtx.Int64("id"))
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

		if err := moonpool.ReleaseSavepoint(context.Background(), "tagupdate"); err != nil {
			return err
		}

		fmt.Printf("%d tag(s) affected (%d added | %d removed)\n", differenceAdd+differenceRemove, differenceAdd, differenceRemove)
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
		c, err := OpenConfig(*cCtx, true)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}

		res, err := moonpool.SearchTag(cCtx.Context, cCtx.String("tag"))
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
		c, err := OpenConfig(*cCtx, true)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}
		defer moonpool.Close()

		q := api.BuildQuery(cCtx.String("tags"))
		res, err := moonpool.QueryTags(cCtx.Context, q)
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
		c, err := OpenConfig(*cCtx, true)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}
		defer moonpool.Close()

		if !moonpool.DoesEntryExist(context.Background(), cCtx.Int64("id")) {
			fmt.Println("id does not exist")
			return nil
		}

		tags, err := moonpool.GetTags(context.Background(), cCtx.Int64("id"))
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
			Usage:    "archive id to list tags from",
			Required: true,
		},
	},
}

var thumbnailGenerateIcons = cli.Command{
	Name:     "thumbnail",
	Aliases:  []string{"t"},
	Category: "thumbnail",
	Usage:    "generate thumbnails for a given archive_id",
	Args:     true,
	Action: func(cCtx *cli.Context) error {
		c, err := OpenConfig(*cCtx, true)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath, ThumbnailLocation: c.ThumbnailPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}
		defer moonpool.Close()

		if err := moonpool.GenerateThumbnail(context.Background(), cCtx.Int64("id")); err != nil {
			return err
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:     "id",
			Usage:    "archive id to generate thumbnail for",
			Required: true,
		},
	},
}

var thumbnailGenerateBlurHash = cli.Command{
	Name:     "blurhash",
	Aliases:  []string{"b"},
	Category: "thumbnail",
	Usage:    "generate blurhash for a given archive_id",
	Args:     true,
	Action: func(cCtx *cli.Context) error {
		c, err := OpenConfig(*cCtx, true)
		if err != nil {
			return err
		}

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath, ThumbnailLocation: c.ThumbnailPath},
			slog.New(slog.NewTextHandler(os.Stdout, nil)))
		if err != nil {
			return err
		}
		defer moonpool.Close()

		if err := moonpool.GenerateBlurHash(cCtx.Context, cCtx.Int64("id")); err != nil {
			return err
		}

		hash, err := moonpool.GetBlurHashString(cCtx.Context, cCtx.Int64("id"))
		if err != nil {
			return err
		}

		fmt.Printf("generated blur hash: %s\n", hash)
		return nil
	},
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:     "id",
			Usage:    "archive id to generate blurhash for",
			Required: true,
		},
	},
}
