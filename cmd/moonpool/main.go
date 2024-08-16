package main

import (
	"context"
	"flag"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/log"
)

var DATABASE_PATH string
var config config.Config

func initFlags() {
	flag.StringVar(&DATABASE_PATH, "database", "archive.sqlite3", "path to moonpool archive")

	archive := flag.NewFlagSet("archive", flag.ExitOnError)
	archive.BoolFunc("new", "create a new moonpool archive", func(s string) error {
		db, err := archive.OpenSQLite3(databasePath)
		if err != nil {
			return err
		}

		logger := log.NewSlogger(context.Background(), log.LogLevelDebug, "api")

		api.New(db, logger, config.Config{})

		return nil
	})
}

func initConfig() {
}

func openMoonpool(databasePath, mediaPath string) error {
	db, err := archive.OpenSQLite3(databasePath)
	if err != nil {
		return err
	}

	logger := log.NewSlogger(context.Background(), log.LogLevelDebug, "api")

	api.New(db, logger, config.Config{})
}

/*
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/config"
	mpLog "github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/server"
	"github.com/dtbead/moonpool/server/www"
	"github.com/pkg/profile"
)

const DATABASE_PATH = "archive.sqlite3"

var conf config.Config

func init() {
	var err error
	conf, err = config.Open("config.json")
	if err != nil {
		fmt.Printf("failed to read config. %v\n", err)
		os.Exit(1)
	}
}

func main() {
	logger := mpLog.NewSlogger(context.Background(), mpLog.LogLevelDebug, "")

	var sql *sql.DB
	var err error
	if conf.ArchivePath != "" {
		sql, err = archive.OpenSQLite3(conf.ArchivePath)
	} else {
		sql, err = archive.OpenSQLite3(DATABASE_PATH)
	}
	if err != nil {
		fmt.Printf("failed to open archive. %v\n", err)
		os.Exit(1)
	}

	archive.InitializeSQLite3(sql)

	// initialize moonpool API
	moonpool := server.New(sql, logger, conf)
	moonpool.E.HideBanner = true
	defer moonpool.Shutdown()
	defer sql.Exec("PRAGMA schema.wal_checkpoint;")
	defer sql.Close()

	// initialize moonpool WebUI
	frontend := www.New(*api.New(sql, logger, conf), conf.MediaPath)
	frontend.E.HideBanner = true
	defer frontend.E.Shutdown(context.Background())

	// enable profling if specified
	if conf.Logging.Profiling == config.PROFILING_CPU {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".profile")).Stop()
	} else {
		if conf.Logging.Profiling == config.PROFILING_MEM {
			defer profile.Start(profile.MemProfile, profile.ProfilePath(".profile")).Stop()
		}
	}

	go moonpool.E.Start("localhost:5878")
	frontend.E.Start("localhost:9996")
}
*/
