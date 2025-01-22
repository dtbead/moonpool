package main

import (
	"fmt"
	"os"

	"github.com/dtbead/moonpool/cmd"
)

const MOONPOOL_VERSION = "v1.0.0-alpha"
const COPYRIGHT = `GNU General Public License v3 or greater (see https://www.gnu.org/licenses/gpl-3.0.txt for more details)`

func main() {
	app := cmd.NewApp()
	app.Version = MOONPOOL_VERSION
	app.Copyright = COPYRIGHT

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
