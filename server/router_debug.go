package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func (m Moonpool) Callback() {
	m.E.GET("debug/", func(c echo.Context) error {
		m.E.Logger.Debug(fmt.Println("wat"))
		return nil
	})
}
