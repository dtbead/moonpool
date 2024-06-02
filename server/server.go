package server

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/log"
	"github.com/labstack/echo/v4"
)

type Moonpool struct {
	E *echo.Echo
	A *api.API
}

func New(l log.Logger, d *sql.DB) *Moonpool {
	m := &Moonpool{
		E: echo.New(),
		A: api.New(l, d),
	}

	db, err := NewDatabase()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	SetLogMiddleware(m.E, db)
	return m
}

// Init initializes every callback function used for Echo
func (m Moonpool) Init() {
	m.Post()
	m.GetFile()
	m.Upload()
	m.SetTags()
	m.RemoveTags()
	m.SetTimestamps()
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
