package db

import (
	"database/sql"
	"log/slog"

	"github.com/dtbead/moonpool/log"
	"github.com/dtbead/moonpool/media"
)

type ArchiveID int

type Database interface {
	InsertEntry(h media.Hashes, path, extension string) (ArchiveID, error)
	AddTag(tag string) error
	AddTags(tags []string) ([]media.Tag, error)
	MapTags(a ArchiveID, tags []string) ([]int, error)
	SearchTag(tag string) ([]media.Entry, error)
	Initialize() error
	Close() error
}

func NewSQLite3(filepath string, l log.Logger) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		slog.Error("unable to create new database. %v'", err)
		return SQLite3{}, err
	}

	sdb := SQLite3{db, nil, l, false}
	if err := sdb.Initialize(); err != nil {
		return SQLite3{}, err
	}

	return sdb, nil
}

func OpenSQLite3(filepath string, l log.Logger) (*SQLite3, error) {
	SQLdb, err := Open(filepath, l)
	if err != nil {
		return nil, err
	}
	return &SQLdb, nil
}
