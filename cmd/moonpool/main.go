package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/server"
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
	defer moonpool.Shutdown()
	defer sql.Exec("PRAGMA schema.wal_checkpoint;")
	defer sql.Close()

	if conf.EnableCPUProfiling {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".profile")).Stop()
	} else {
		if conf.EnableMemProfiling {
			defer profile.Start(profile.MemProfile, profile.ProfilePath(".profile")).Stop()
		}
	}

	fmt.Println(moonpool.E.Start("localhost:5878"))
}
