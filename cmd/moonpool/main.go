package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/config"
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
	var sql *sql.DB
	var err error
	if conf.ArchivePath() != "" {
		sql, err = archive.OpenSQLite3(conf.ArchivePath())
	} else {
		sql, err = archive.OpenSQLite3(DATABASE_PATH)
	}

	if err != nil {
		fmt.Printf("failed to open archive. %v\n", err)
		os.Exit(1)
	}
	archive.InitializeSQLite3(sql)

	moonpool := server.New(sql, conf)
	moonpool.E.HideBanner = true
	defer moonpool.Shutdown()
	defer sql.Exec("PRAGMA schema.wal_checkpoint;")
	defer sql.Close()

	frontend := www.New(*api.New(sql, conf))
	frontend.E.HideBanner = true
	defer frontend.E.Shutdown(context.Background())

	if conf.EnableCPUProfiling {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".profile")).Stop()
	} else {
		if conf.EnableMemProfiling {
			defer profile.Start(profile.MemProfile, profile.ProfilePath(".profile")).Stop()
		}
	}

	go frontend.E.Start("localhost:9996")
	go fmt.Println(moonpool.E.Start("localhost:5878"))
}
