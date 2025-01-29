package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/dtbead/moonpool/internal/www"

	"github.com/urfave/cli/v2"
)

var launch = cli.Command{
	Name:  "launch",
	Usage: "run a new moonpool instance",
	Action: func(cCtx *cli.Context) error {
		if cCtx.IsSet("address") {
			moonpoolConfig.ListenAddress = cCtx.String("address")
		}

		if cCtx.IsSet("webui") {
			moonpoolConfig.WebUIPort = cCtx.Int("webui")
		}

		loggerMain := log.New(log.StringToLogLevel(moonpoolConfig.Logging.LogLevel))
		loggerWebUI := loggerMain.WithGroup("webui")

		apiConfig := api.Config{ArchiveLocation: moonpoolConfig.ArchivePath, MediaLocation: moonpoolConfig.MediaPath, ThumbnailLocation: moonpoolConfig.ThumbnailPath}
		moonpoolAPI, err := api.Open(apiConfig, loggerMain)
		if err != nil {
			return err
		}

		webFrontend, err := www.New(moonpoolAPI, www.Config{
			DynamicWebReloading:     moonpoolConfig.Debug.DynamicWebReloading.Enable,
			DynamicWebReloadingPath: moonpoolConfig.Debug.DynamicWebReloading.Path,
			Log:                     loggerWebUI,
		})
		if err != nil {
			return err
		}

		shutdown := func() error {
			return errors.Join(
				moonpoolAPI.Close(context.Background()),
				webFrontend.Shutdown(context.Background()),
			)
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
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

		services := make(chan error, 1)
		go func() {
			err := webFrontend.Start(fmt.Sprintf("%s:%d", moonpoolConfig.ListenAddress, moonpoolConfig.WebUIPort))
			if err != nil {
				services <- err
			}
		}()

		for err := range services {
			if err != nil {
				return err
			}
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
	},
}
