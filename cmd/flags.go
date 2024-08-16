package cmd

import "flag"

func init() {
	flag.StringVar()

	archive := flag.NewFlagSet("archive", flag.ExitOnError)
	archive.BoolFunc("new", "create a new moonpool archive")

}
