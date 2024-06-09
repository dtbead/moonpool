package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/labstack/echo/v4"
)

type Moonpool struct {
	E *echo.Echo
	A *api.API
}

// Init initializes every callback function used for Echo
func (m Moonpool) init() {
	m.Post()
	m.Search()
	m.GetFile()
	m.Upload()
	m.SetTags()
	m.RemoveTags()
	m.SetTimestamps()
	m.GetTimestamps()
}

func New(d *sql.DB, c config.Config) *Moonpool {
	m := &Moonpool{
		E: echo.New(),
		A: api.New(d, c),
	}

	log, err := NewDatabase()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	SetLogMiddleware(m.E, log, c.EnableFileLogging)
	m.init()
	return m
}

func (m Moonpool) Shutdown() {
	m.E.Shutdown(context.TODO())
}

// returns -1 on invalid id string. returns int64 >= 1 otherwise
func (m Moonpool) parseArchiveID(id string) int64 {
	archive_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return -1
	}

	return archive_id
}

var ErrInvalidArchiveID = fmt.Errorf("invalid archive ID")
