package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/importer"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/urfave/cli/v2"
)

var archiveImport = cli.Command{
	Name:  "import",
	Usage: "imports a new file into moonpool",
	Action: func(cCtx *cli.Context) error {
		path := cCtx.Path("path")

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: moonpoolConfig.ArchivePath, MediaLocation: moonpoolConfig.MediaPath, ThumbnailLocation: moonpoolConfig.ThumbnailPath},
			log.New(log.StringToLogLevel(moonpoolConfig.Logging.LogLevel)))
		if err != nil {
			return err
		}
		defer moonpool.Close(cCtx.Context)

		var imported, failed int
		var ext strings.Builder

		var supportedExt []string = []string{
			".png",
			".jpg",
			".jpeg",
			".webp",
			".gif",
			".mp4",
		}

		var scan = func(path string, d os.DirEntry, inpErr error) (err error) {
			ext.Reset()
			ext.WriteString(filepath.Ext(path))

			if !slices.Contains(supportedExt, ext.String()) {
				fmt.Printf("skipped \"%s\" (unsupported format)\n", path)
				failed++
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				failed++
				return err
			}
			defer f.Close()

			archive_id, err := fileImport(*cCtx, *moonpool, f, ext.String())
			if err != nil {
				failed++
				return err
			}

			_ = moonpool.GeneratePerceptualHash(cCtx.Context, archive_id, "", f)
			_ = moonpool.GenerateThumbnail(cCtx.Context, archive_id)

			imported++
			return nil
		}

		err = filepath.WalkDir(path, scan)
		if err != nil {
			return err
		}

		fmt.Printf("imported %d entries (%d failed)\n", imported, failed)
		return nil
	},
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:     "path",
			Aliases:  []string{"f, p"},
			Usage:    "file or folder to import from",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:    "tags",
			Aliases: []string{"t"},
			Usage:   "tags to assign with during import",
		},
	},
}

func fileImport(cCtx cli.Context, moonpool api.API, f *os.File, ext string) (archive_id int64, err error) {
	importer, err := importer.New(f, ext)
	if err != nil {
		return -1, err
	}

	err = moonpool.NewSavepoint(cCtx.Context, "folderimport")
	if err != nil {
		return -1, err
	}
	defer moonpool.RollbackSavepoint(cCtx.Context, "folderimport")

	archive_id, err = moonpool.Import(cCtx.Context, importer)
	if err != nil {
		return -1, err
	}

	err = moonpool.SetTimestamps(cCtx.Context, archive_id, importer.Timestamp())
	if err != nil {
		return -1, err
	}

	err = moonpool.AssignTags(cCtx.Context, archive_id, cCtx.StringSlice("tags"))
	if err != nil {
		return -1, err
	}

	err = moonpool.GenerateFileMetadata(cCtx.Context, archive_id)
	if err != nil {
		return -1, err
	}

	err = moonpool.ReleaseSavepoint(cCtx.Context, "folderimport")
	if err != nil {
		return -1, err
	}

	f.Seek(0, io.SeekStart)
	return -1, nil
}
