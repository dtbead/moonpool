package server

import (
	"database/sql"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/log"
	"github.com/labstack/echo/v4"
)

type Moonpool struct {
	E *echo.Echo
	a *api.API
}

func New(l log.Logger, d *sql.DB) *Moonpool {
	return &Moonpool{
		E: echo.New(),
		a: api.New(l, d),
	}
}

// Init initializes every callback function used for Echo
func (m Moonpool) Init() {
	m.Post()
	m.GetFile()
	m.Upload()
}
