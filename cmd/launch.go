package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/dtbead/moonpool/internal/profile"
	"github.com/dtbead/moonpool/internal/server"
	"github.com/dtbead/moonpool/internal/www"

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

		var p profile.Profile
		if cCtx.IsSet("profile") {
			c.Logging.Profiling = cCtx.String("profile")
		}

		if strings.EqualFold(c.Logging.Profiling, config.PROFILING_CPU) {
			p, err = profile.New(config.PROFILING_CPU)
			if err != nil {
				return err
			}
		}

		log := log.New(log.StringToLogLevel(c.Logging.LogLevel))
		apiConfig := api.Config{ArchiveLocation: c.ArchivePath, MediaLocation: c.MediaPath, ThumbnailLocation: c.ThumbnailPath}

		api, err := api.Open(apiConfig, log)
		if err != nil {
			return err
		}

		services := make(chan error, 2)
		webAPI := server.New(api, c)
		webFrontend, err := www.New(apiConfig, www.Config{
			DynamicWebReloading:     c.Debug.DynamicWebReloading.Enable,
			DynamicWebReloadingPath: c.Debug.DynamicWebReloading.Path,
		})
		if err != nil {
			return err
		}

		shutdown := func() error {
			api.Close()
			webFrontend.Shutdown(context.Background())
			webAPI.Shutdown()

			if p != (profile.Profile{}) {
				if err := p.Stop(); err != nil {
					fmt.Printf("failed to stop profiler, %v\n", err)
					return err
				}
			}
			return nil
		}

		go func() {
			err := webAPI.Start(fmt.Sprintf("%s:%d", c.ListenAddress, c.APIPort))
			if err != nil {
				services <- err
			}
		}()

		go func() {
			err := webFrontend.Start(fmt.Sprintf("%s:%d", c.ListenAddress, c.WebUIPort))
			if err != nil {
				services <- err
			}
		}()

		for err := range services {
			if err != nil {
				return err
			}
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			for s := range sig {
				fmt.Printf("received %s signal, shutting down...\n", s.String())
				if err := shutdown(); err != nil {
					fmt.Printf("error during graceful shutdown, %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			}
		}()

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
