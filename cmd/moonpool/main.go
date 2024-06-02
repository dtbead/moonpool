package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/server"
)

const DATABASE_PATH = "archive.sqlite3"

/*
var Moonpool *server.API

func init() {

		l := log.NewSlogLogger(context.Background())
		sql, err := archive.OpenSQLite3(DATABASE_PATH)
		if err != nil {
			l.Error(err.Error())
			os.Exit(1)
		}

		//archive.InitializeSQLite3(sql)

		Moonpool = server.NewAPI(l, sql)
	}
*/
func main() {
	l := log.NewSlogLogger(context.TODO())
	sql, err := archive.OpenSQLite3(DATABASE_PATH)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	moonpool := NewServer(l, sql)
	moonpool.Init()

	go l.Error(moonpool.E.Start("localhost:5878").Error())
	os.Exit(1)
}

func NewServer(l log.Logger, d *sql.DB) *server.Moonpool {
	return server.New(l, d)
}
