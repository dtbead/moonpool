package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	m.GetHashes()
}

func New(d *sql.DB, l *slog.Logger, c config.Config) *Moonpool {
	m := &Moonpool{
		E: echo.New(),
		A: api.New(d, l, c),
	}

	log, err := NewDatabase()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	SetLogMiddleware(m.E, log, c.Logging.FileLogging)
	m.init()
	return m
}

func (m Moonpool) Shutdown() {
	m.E.Shutdown(context.TODO())
}

// ValidateArchiveID validates a given string and determines whether it is a valid integer >=1, and integer
// exists as an archive id in entry. Returns -1 on invalid IDs
func ValidateArchiveID(a api.API, id string) int64 {
	archive_id, err := strconv.ParseInt(strings.ReplaceAll(id, "/", ""), 10, 64)
	if err != nil || archive_id <= 0 {
		return -1
	}

	if a.DoesEntryExist(context.Background(), archive_id) {
		return archive_id
	}

	return -1
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
