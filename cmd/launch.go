package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/dtbead/moonpool/server"
	"github.com/dtbead/moonpool/server/www"
	"github.com/urfave/cli/v2"
)

var launch = cli.Command{
	Name:  "launch",
	Usage: "run a new moonpool instance",
	Action: func(cCtx *cli.Context) error {
		c, err := OpenConfig(*cCtx, false)
		if err != nil {
			return err
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

		l := log.NewSlogger(context.Background(), log.StringToLogLevel(c.Logging.LogLevel), "api")

		moonpool, err := api.Open(
			api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath, ThumbnailLocation: c.ThumbnailPath}, l)
		if err != nil {
			return err
		}

		services := make(chan error, 2)
		webAPI := server.New(moonpool, c)
		webFrontend := www.New(moonpool)

		go func() {
			services <- webAPI.Start(c.ListenAddress + ":" + fmt.Sprint(c.APIPort))
		}()

		go func() {
			services <- webFrontend.Start(c.ListenAddress + ":" + fmt.Sprint(c.WebUIPort))
		}()

		errChan := <-services
		if errChan != nil {
			return errChan
		}

		return nil
	},
	Flags: []cli.Flag{
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
