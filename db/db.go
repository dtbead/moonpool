package db

import (
	"database/sql"
	"log/slog"

	"github.com/dtbead/moonpool/media"
)

type Database interface {
	InsertEntry(h media.Hashes, path, extension string) (int, error)
	AddTag(tag string) error
	AddTags(tags []string) ([]media.Tag, error)
	MapTags(archiveID int, tags []string) error
	MapTagsWithID(archiveID int, tags []media.Tag) error
	SearchTag(tag string) ([]media.Entry, error)
	Initialize() error
	Close() error
}

func NewSQLite3(filepath string) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		slog.Error("unable to create new database. %v'", err)
		return SQLite3{}, err
	}

	sdb := SQLite3{db, nil, false}
	if err := sdb.Initialize(); err != nil {
		return SQLite3{}, err
	}

	return sdb, nil
}

func OpenSQLite3(filepath string) (*SQLite3, error) {
	SQLdb, err := Open(filepath)
	if err != nil {
		return nil, err
	}
	return &SQLdb, nil
}
