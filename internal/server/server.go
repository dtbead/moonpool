package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/config"
	"github.com/labstack/echo/v4"
)

type Server struct {
	e *echo.Echo
	a *api.API
}

func (s Server) init() {
	s.Post()
	s.Search()
	s.Upload()
	s.SetTags()
	s.SetTimestamps()
	s.GetTimestamps()
	s.RemoveTags()
	s.GetFile()
	s.GetHashes()
}

func New(a *api.API, c config.Config) Server {
	m := Server{
		e: echo.New(),
		a: a,
	}

	m.init()
	return m
}

func (s Server) Start(ListenAddress string) error {
	return s.e.Start(ListenAddress)
}

func (s Server) Shutdown() error {
	if err := s.a.Close(); err != nil {
		return err
	}
	return s.e.Shutdown(context.TODO())
}

func stringToInt64(s string) int64 {
	archive_id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}

	return archive_id
}

var ErrInvalidArchiveID = fmt.Errorf("invalid archive ID")

func isDeadlined(c echo.Context, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("[%s] WARNING: request timed-out\n", c.Request().RemoteAddr)
		c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
		return c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
	}
	return nil
}