package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/server/www"
	"github.com/urfave/cli/v2"
)

var launch = cli.Command{
	Name:  "",
	Usage: "start a new moonpool instance",
	Action: func(cCtx *cli.Context) error {
		db, a, err := newMoonpool(c.MediaPath, c.ArchivePath)
		if err != nil {
			fmt.Printf("failed to launch moonpool instance. %v\n", err)
		}
		db.Close()

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
}

func newMoonpool(mediapath, archivepath string) (*sql.DB, *api.API, error) {
	db, err := archive.OpenSQLite3(c.ArchivePath)
	if err != nil {
		return &sql.DB{}, &api.API{}, err
	}

	l := log.NewSlogger(context.Background(), slog.LevelDebug, "api")
	a := api.New(db, l, c)

	return db, a, nil
}
