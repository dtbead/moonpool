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
		if c.Logging.Profiling == config.PROFILING_CPU {
			newProfiler(config.PROFILING_CPU)
		}

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
		&cli.StringFlag{
			Name:        "profiling",
			Aliases:     []string{"profile", "prof", "p"},
			Usage:       "enable performance profiling for debugging purposes ('cpu' OR 'memory')",
			Destination: &c.Logging.Profiling,
		},
	},
	Before: func(ctx *cli.Context) error {
		err := initConfig()
		if ctx.String("config") != CONFIG_DEFAULT_PATH && err != nil {
			fmt.Printf("failed to load config, %v. refusing to use defaults\n", err)
			os.Exit(1)
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
