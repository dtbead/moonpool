package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
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
			moonpool <- api.Start(c.ListenAddress + ":" + fmt.Sprint(c.APIPort))
		}()

		go func() {
			moonpool <- web.Start(c.ListenAddress + ":" + fmt.Sprint(c.WebUIPort))
		}()

		errChan := <-moonpool
		if errChan != nil {
			return errChan
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "database",
			Aliases: []string{"db"},
			Usage:   "path to moonpool database",
			Value:   config.DefaultValues().ArchivePath,
		},
		&cli.StringFlag{
			Name:    "media",
			Aliases: []string{"m"},
			Usage:   "path to root media folder",
			Value:   config.DefaultValues().MediaPath,
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"ip"},
			Usage:   "ip to listen on",
			Value:   config.DefaultValues().ListenAddress,
		},
		&cli.IntFlag{
			Name:  "webui",
			Usage: "port to launch webui on",
			Value: config.DefaultValues().WebUIPort,
		},
		&cli.IntFlag{
			Name:  "api",
			Usage: "port to launch api on",
			Value: config.DefaultValues().APIPort,
		},
		&cli.StringFlag{
			Name:     "profile",
			Category: "debug",
			Aliases:  []string{"prof", "p"},
			Usage:    "enable performance profiling for debugging purposes ('cpu' OR 'memory')",
			Value:    config.DefaultValues().Logging.Profiling,
		},
	},
	Before: func(cCtx *cli.Context) error {
		if cCtx.IsSet("database") {
			c.ArchivePath = cCtx.String("database")
		}

		if cCtx.IsSet("media") {
			c.MediaPath = cCtx.String("media")
		}

		if cCtx.IsSet("address") {
			c.ListenAddress = cCtx.String("address")
		}

		if cCtx.IsSet("webui") {
			c.WebUIPort = cCtx.Int("webui")
		}

		if cCtx.IsSet("api") {
			c.APIPort = cCtx.Int("api")
		}

		if cCtx.IsSet("profile") {
			c.Logging.Profiling = cCtx.String("profile")
		}

		if c.Logging.Profiling == config.PROFILING_CPU {
			newProfiler(config.PROFILING_CPU)
		}
		return nil
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

func newProfiler(profilerType string) {
	if err := os.Mkdir("./profile", os.ModeDir); err != nil && !errors.Is(err, os.ErrExist) {
		fmt.Printf("failed to create profile folder, %v\n", err)
		os.Exit(1)
	}

	fileCPU, err := os.Create("./profile/cpu.prof")
	if err != nil {
		fmt.Printf("failed to create cpu.prof. %v\n", err)
		os.Exit(1)
	}
	defer fileCPU.Close()

	fileMem, err := os.Create("./profile/mem.prof")
	if err != nil {
		fmt.Printf("failed to create mem.prof. %v\n", err)
		os.Exit(1)
	}
	defer fileMem.Close()

	switch profilerType {
	case config.PROFILING_CPU:
		pprof.StartCPUProfile(fileCPU)
		defer pprof.StopCPUProfile()
	default:
		fmt.Println("unknown profiling method type")
		os.Exit(1)
	}

	runtime.GC()
	if err := pprof.WriteHeapProfile(fileMem); err != nil {
		fmt.Printf("failed to write mem.prof. %v\n", err)
		os.Exit(1)
	}
}
