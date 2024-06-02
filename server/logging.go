package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const SQL_SCHEMA = `CREATE TABLE "debug" (
	"time"	TEXT NOT NULL,
	"msg"	TEXT,
	"uri"	TEXT NOT NULL,
	"status"	INTEGER NOT NULL,
	"error"	TEXT
);
CREATE TABLE "warn" (
	"time"	TEXT NOT NULL,
	"msg"	TEXT,
	"uri"	TEXT NOT NULL,
	"status"	INTEGER NOT NULL,
	"error"	TEXT
);
CREATE TABLE "error" (
	"time"	TEXT NOT NULL,
	"msg"	TEXT,
	"uri"	TEXT NOT NULL,
	"status"	INTEGER NOT NULL,
	"error"	TEXT
);`

func NewDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "logs.sqlite3")
	if err != nil {
		return nil, err
	}

	/*
		if _, err := db.Exec(SQL_SCHEMA); err != nil {
			return nil, err
		}
	*/
	return db, nil
}

type JSONLog struct {
	Time      string
	Level     string
	Remote_ip string
	Latency   string
	Uri       string
	Method    string
	Status    int
	Msg       string
	Err       string
}

func SetLogMiddleware(e *echo.Echo, s *sql.DB) {
	buf := new(bytes.Buffer)

	logger := slog.New(slog.NewJSONHandler(buf, nil))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("remote_ip", v.RemoteIP),
					slog.String("uri", v.URI),
					slog.Int64("latency", v.Latency.Milliseconds()),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelDebug, "REQUEST_DEBUG",
					slog.String("remote_ip", v.RemoteIP),
					slog.String("uri", v.URI),
					slog.Int64("latency", v.Latency.Milliseconds()),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
				)
			}

			go StoreLog(buf, s)
			return nil
		},
	}))

}

func StoreLog(buf *bytes.Buffer, db *sql.DB) {
	j := new(JSONLog)

	enc := json.NewDecoder(buf)
	enc.Decode(j)

	insertDebug, _ := db.Prepare(`INSERT INTO debug (time, latency, remote_ip, msg, uri, status, error) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	_, err := insertDebug.Exec(j.Time, j.Latency, j.Remote_ip, j.Msg, j.Uri, j.Status, j.Err)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Printf(`[%s] [%s] %s | %s '%s' |  MSG: "%s"`, j.Time, j.Remote_ip, j.Level, j.Method, j.Uri, j.Err)
}
