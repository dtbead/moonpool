package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/server"
	"github.com/dtbead/moonpool/server/www"
	"github.com/urfave/cli/v2"
)

var launch = cli.Command{
	Name:    "launch",
	Usage:   "run a new moonpool instance",
	Aliases: []string{"run", "start", ""},
	Action: func(cCtx *cli.Context) error {
		_, a, err := newMoonpool(c.MediaPath, c.ArchivePath)
		if err != nil {
			fmt.Printf("failed to launch moonpool instance. %v\n", err)
		}
		defer a.Close()

		moonpool := make(chan error, 2)
		api := server.New(a, c)
		web := www.New(a, c.MediaPath)

		go func() {
			moonpool <- api.Start("127.0.0.1:" + fmt.Sprint(c.APIPort))
		}()

		go func() {
			moonpool <- web.Start("127.0.0.1:" + fmt.Sprint(c.WebUIPort))
		}()

		errChan := <-moonpool
		if errChan != nil {
			return errChan
		}

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
			Destination: &c.APIPort,
		},
	},
}

func newMoonpool(mediaPath, dbPath string) (*sql.DB, *api.API, error) {
	db, err := db.OpenSQLite3(dbPath)
	if err != nil {
		return &sql.DB{}, &api.API{}, err
	}

	l := log.NewSlogger(context.Background(), slog.LevelDebug, "api")
	a := api.New(db, l, api.Config{MediaLocation: mediaPath})

	return db, a, nil
}
