package main

import (
	"fmt"
	"os"

	"github.com/dtbead/moonpool/cmd"
)

const MOONPOOL_VERSION = "v0.1.1-alpha"

func main() {
	app := cmd.NewApp()
	app.Version = MOONPOOL_VERSION

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
