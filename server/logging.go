package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const SQL_SCHEMA = `CREATE TABLE "logs" (
	"time"	TEXT NOT NULL,
	"latency"	TEXT,
	"remote_ip" TEXT NOT NULL,
	"user_agent" TEXT,
	"uri"	TEXT NOT NULL,
	"method" TEXT NOT NULL,
	"status"	INTEGER NOT NULL,
	"error"	TEXT
);`

func NewDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "logs.sqlite3")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(SQL_SCHEMA); err != nil {
		fmt.Println(err)
		return db, nil
	}

	return db, nil
}

func SetLogMiddleware(e *echo.Echo, s *sql.DB, WriteToFile bool) {
	buf := new(bytes.Buffer)

	logger := slog.New(slog.NewJSONHandler(buf, nil))
	insertStmt, err := s.Prepare("INSERT INTO logs VALUES (?, ?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		fmt.Println(err)
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogStatus:    true,
		LogURI:       true,
		LogError:     true,
		HandleError:  true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if WriteToFile {
				go insertStmt.ExecContext(context.Background(), v.StartTime.String(), v.Latency, v.RemoteIP, v.UserAgent, v.URI, v.Method, v.Status,
					func() string {
						if err != nil {
							return err.Error()
						}
						return ""
					}())
			}

			if v.Error != nil {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("remote_ip", v.RemoteIP),
					slog.String("user_agent", v.UserAgent),
					slog.String("uri", v.URI),
					slog.String("latency", v.Latency.String()),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST_INFO",
					slog.String("remote_ip", v.RemoteIP),
					slog.String("user_agent", v.UserAgent),
					slog.String("uri", v.URI),
					slog.String("latency", v.Latency.String()),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
				)
			}
			return nil
		},
	}))

}
